package gateway

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/giangsamne/TailRouter/internal/config"
	"github.com/giangsamne/TailRouter/internal/scanner"
	"github.com/giangsamne/TailRouter/internal/service"
	"github.com/giangsamne/TailRouter/internal/tailscale"
	"github.com/giangsamne/TailRouter/web"
)

const (
	DefaultPort = 65534
	Version = "2.0.1-go-native"
	MaxLogSize  = 5 * 1024 * 1024 // 5MB max log file
)

type Server struct {
	port       int
	routes     *config.Manager
	scanner    *scanner.Scanner
	tailscale  *tailscale.Client
	service    *service.Manager
	startTime  time.Time
	logFile    *os.File
	logger     *log.Logger
	mu         sync.Mutex
	httpServer *http.Server
}

func NewServer(port int, configPath string) *Server {
	if port <= 0 {
		port = DefaultPort
	}

	routesMgr := config.NewManager(configPath)
	scannerMgr := scanner.NewScanner()
	tsClient := tailscale.NewClient()
	svcMgr := service.NewManager()

	s := &Server{
		port:      port,
		routes:    routesMgr,
		scanner:   scannerMgr,
		tailscale: tsClient,
		service:   svcMgr,
		startTime: time.Now(),
	}
	s.setupLogger()
	return s
}

func (s *Server) setupLogger() {
	execDir, err := os.Executable()
	if err != nil {
		execDir = "."
	} else {
		execDir = filepath.Dir(execDir)
	}
	logPath := filepath.Join(execDir, "tailrouter.log")

	// Rotate log if exceeds 5MB
	if fi, err := os.Stat(logPath); err == nil && fi.Size() > MaxLogSize {
		_ = os.Rename(logPath, logPath+".1")
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		s.logger = log.New(os.Stdout, "[TailRouter] ", log.LstdFlags)
		return
	}
	s.logFile = f
	multi := io.MultiWriter(os.Stdout, f)
	s.logger = log.New(multi, "[TailRouter] ", log.LstdFlags)
}

func (s *Server) Close() {
	if s.httpServer != nil {
		_ = s.httpServer.Close()
	}
	if s.logFile != nil {
		_ = s.logFile.Close()
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.logger.Printf("Khởi chạy TailRouter Native Engine trên cổng %d (Dual-Stack IPv4/IPv6)...", s.port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.corsMiddleware(mux),
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Dual-stack listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("không thể mở cổng %d: %w", s.port, err)
	}

	s.logger.Printf("TailRouter Gateway sẵn sàng tại: http://localhost:%d/router", s.port)
	return s.httpServer.Serve(listener)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-TailRouter-Hop")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	reqPath := r.URL.Path

	// 1. Web Dashboard routes
	if reqPath == "/" || reqPath == "/router" || reqPath == "/router/" || strings.HasPrefix(reqPath, "/router/") {
		s.serveWebDashboard(w, r)
		return
	}

	// 2. Favicon
	if reqPath == "/favicon.ico" || reqPath == "/favicon.svg" {
		w.Header().Set("Content-Type", "image/svg+xml")
		fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#2563eb" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>`)
		return
	}

	// 3. REST API routes
	if strings.HasPrefix(reqPath, "/api/") {
		s.handleAPI(w, r)
		return
	}

	// 4. Reverse Proxying for matched routes
	s.handleReverseProxy(w, r)
}

func (s *Server) serveWebDashboard(w http.ResponseWriter, r *http.Request) {
	indexData, err := web.Content.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Không tìm thấy giao diện Web Dashboard nhúng", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(indexData)
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		s.jsonResponse(w, http.StatusOK, map[string]string{"message": "TailRouter Native API Ready"})
		return
	}

	switch parts[0] {
	case "status":
		s.handleAPIStatus(w, r)
	case "routes":
		s.handleAPIRoutes(w, r, parts[1:])
	case "scan":
		s.handleAPIScan(w, r)
	case "tailscale":
		s.handleAPITailscale(w, r, parts[1:])
	case "autostart":
		s.handleAPIAutostart(w, r)
	default:
		s.jsonResponse(w, http.StatusNotFound, map[string]string{"error": "API endpoint không tồn tại"})
	}
}

func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	tsStatus, _ := s.tailscale.GetStatus()
	autoStatus := s.service.GetStatus()
	routesList := s.routes.List()

	data := map[string]interface{}{
		"status":         "online",
		"version":        Version,
		"uptime_seconds": int(time.Since(s.startTime).Seconds()),
		"routes_count":   len(routesList),
		"routes":         routesList,
		"platform":       runtime.GOOS,
		"arch":           runtime.GOARCH,
		"tailscale":      tsStatus,
		"autostart":      autoStatus,
	}
	s.jsonResponse(w, http.StatusOK, data)
}

func (s *Server) handleAPIRoutes(w http.ResponseWriter, r *http.Request, subParts []string) {
	if len(subParts) == 0 {
		// /api/routes
		if r.Method == http.MethodGet {
			s.jsonResponse(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"routes":  s.routes.List(),
			})
			return
		}
		if r.Method == http.MethodPost {
			var body struct {
				Name        string `json:"name"`
				Path        string `json:"path"`
				TargetHost  string `json:"target_host"`
				TargetPort  int    `json:"target_port"`
				StripPrefix bool   `json:"strip_prefix"`
				Mode        string `json:"mode"`
				Notes       string `json:"notes"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				s.jsonResponse(w, http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"message": "Dữ liệu JSON không hợp lệ",
					"error":   "Dữ liệu JSON không hợp lệ",
				})
				return
			}
			route, err := s.routes.Add(body.Name, body.Path, body.TargetHost, body.TargetPort, body.StripPrefix, body.Mode, body.Notes)
			if err != nil {
				s.jsonResponse(w, http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"message": err.Error(),
					"error":   err.Error(),
				})
				return
			}

			// Apply to Tailscale Serve/Funnel in background
			go func() {
				_ = s.tailscale.ApplyRoute(route.Path, route.TargetHost, route.TargetPort, route.Mode)
			}()

			s.jsonResponse(w, http.StatusCreated, map[string]interface{}{
				"success":     true,
				"message":     "Đã thêm tuyến đường thành công",
				"route":       route,
				"id":          route.ID,
				"name":        route.Name,
				"path":        route.Path,
				"target_host": route.TargetHost,
				"target_port": route.TargetPort,
				"mode":        route.Mode,
				"enabled":     route.Enabled,
			})
			return
		}
		s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"success": false,
			"error":   "Phương thức không được hỗ trợ",
			"message": "Phương thức không được hỗ trợ",
		})
		return
	}

	if r.Method == http.MethodDelete {
		rawTarget := strings.TrimPrefix(r.URL.Path, "/api/routes/")
		rawTarget = strings.Trim(rawTarget, "/")
		route, err := s.routes.Delete(rawTarget)
		if err != nil {
			s.jsonResponse(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": err.Error(),
				"error":   err.Error(),
			})
			return
		}
		// Remove from Tailscale in background
		go func() {
			_ = s.tailscale.RemoveRoute(route.Path)
		}()
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Đã xóa route thành công",
		})
		return
	}

	routeID := subParts[0]

	// Handle /api/routes/sync-tailscale
	if routeID == "sync-tailscale" {
		_ = s.routes.Load()
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Đã đồng bộ tuyến đường thành công",
			"routes":  s.routes.List(),
		})
		return
	}

	if len(subParts) == 1 {
		// /api/routes/{id}
		if r.Method == http.MethodGet {
			if route, ok := s.routes.Get(routeID); ok {
				s.jsonResponse(w, http.StatusOK, route)
				return
			}
			s.jsonResponse(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"error":   "Không tìm thấy route",
				"message": "Không tìm thấy route",
			})
			return
		}
		if r.Method == http.MethodPut {
			var updates map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
				s.jsonResponse(w, http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"message": "Dữ liệu JSON không hợp lệ",
					"error":   "Dữ liệu JSON không hợp lệ",
				})
				return
			}
			updated, err := s.routes.Update(routeID, updates)
			if err != nil {
				s.jsonResponse(w, http.StatusBadRequest, map[string]interface{}{
					"success": false,
					"message": err.Error(),
					"error":   err.Error(),
				})
				return
			}
			go func() {
				if updated.Enabled {
					_ = s.tailscale.ApplyRoute(updated.Path, updated.TargetHost, updated.TargetPort, updated.Mode)
				} else {
					_ = s.tailscale.RemoveRoute(updated.Path)
				}
			}()
			s.jsonResponse(w, http.StatusOK, map[string]interface{}{
				"success":     true,
				"message":     "Đã cập nhật tuyến đường thành công",
				"route":       updated,
				"id":          updated.ID,
				"name":        updated.Name,
				"path":        updated.Path,
				"target_host": updated.TargetHost,
				"target_port": updated.TargetPort,
				"mode":        updated.Mode,
				"enabled":     updated.Enabled,
			})
			return
		}
	}

	// Actions on /api/routes/{id}/{action}
	if len(subParts) >= 2 {
		action := subParts[1]
		switch action {
		case "toggle":
			if r.Method != http.MethodPost {
				s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]interface{}{"error": "Cần phương thức POST"})
				return
			}
			route, err := s.routes.Toggle(routeID)
			if err != nil {
				s.jsonResponse(w, http.StatusNotFound, map[string]interface{}{
					"success": false,
					"message": err.Error(),
					"error":   err.Error(),
				})
				return
			}
			go func() {
				if route.Enabled {
					_ = s.tailscale.ApplyRoute(route.Path, route.TargetHost, route.TargetPort, route.Mode)
				} else {
					_ = s.tailscale.RemoveRoute(route.Path)
				}
			}()
			actionMsg := "Đã bật route"
			if !route.Enabled {
				actionMsg = "Đã tắt route"
			}
			s.jsonResponse(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": actionMsg,
				"enabled": route.Enabled,
				"route":   route,
			})
			return

		case "toggle-mode":
			if r.Method != http.MethodPost {
				s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]interface{}{"error": "Cần phương thức POST"})
				return
			}
			route, err := s.routes.ToggleMode(routeID)
			if err != nil {
				s.jsonResponse(w, http.StatusNotFound, map[string]interface{}{
					"success": false,
					"message": err.Error(),
					"error":   err.Error(),
				})
				return
			}
			go func() {
				_ = s.tailscale.ApplyRoute(route.Path, route.TargetHost, route.TargetPort, route.Mode)
			}()
			s.jsonResponse(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("Đã chuyển chế độ sang %s", route.Mode),
				"mode":    route.Mode,
				"route":   route,
			})
			return

		case "ping":
			if r.Method != http.MethodPost {
				s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]interface{}{"error": "Cần phương thức POST"})
				return
			}
			status, latency, err := s.routes.PingRoute(routeID)
			if err != nil {
				s.jsonResponse(w, http.StatusNotFound, map[string]interface{}{
					"success":    false,
					"online":     false,
					"status":     "offline",
					"latency_ms": 0,
					"error":      err.Error(),
					"message":    err.Error(),
				})
				return
			}
			s.jsonResponse(w, http.StatusOK, map[string]interface{}{
				"success":    true,
				"online":     status == "online",
				"status":     status,
				"latency_ms": latency,
			})
			return

		default:
			s.jsonResponse(w, http.StatusNotFound, map[string]interface{}{"error": "Hành động không hợp lệ"})
			return
		}
	}
}

func (s *Server) handleAPIScan(w http.ResponseWriter, r *http.Request) {
	ports, err := s.scanner.Scan()
	if err != nil {
		s.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{"error": err.Error(), "message": err.Error(), "success": false})
		return
	}
	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"ports":   ports,
		"count":   len(ports),
	})
}

func (s *Server) handleAPITailscale(w http.ResponseWriter, r *http.Request, subParts []string) {
	if len(subParts) == 0 {
		status, _ := s.tailscale.GetStatus()
		s.jsonResponse(w, http.StatusOK, status)
		return
	}

	action := subParts[0]
	switch action {
	case "serve":
		if r.Method != http.MethodPost {
			s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]interface{}{"error": "Cần phương thức POST", "success": false})
			return
		}
		var req struct {
			Target string `json:"target"` // "router" or "gateway"
			Action string `json:"action"` // "serve_router" or "serve_all"
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		var err error
		if req.Target == "gateway" || req.Action == "serve_all" || req.Action == "gateway" {
			err = s.tailscale.ConfigureServeGateway(s.port)
		} else {
			err = s.tailscale.ConfigureServeRouter(s.port)
		}
		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "Access denied") || strings.Contains(msg, "operator") {
				msg = "Cần cấp quyền Operator cho Tailscale: Chạy 'sudo tailscale set --operator=$USER' trên terminal một lần"
			}
			s.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "error": msg, "message": msg})
			return
		}
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Đã cấu hình Tailscale Serve thành công"})

	case "reset":
		if r.Method != http.MethodPost {
			s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]interface{}{"error": "Cần phương thức POST", "success": false})
			return
		}
		if err := s.tailscale.ResetServe(); err != nil {
			msg := err.Error()
			if strings.Contains(msg, "Access denied") || strings.Contains(msg, "operator") {
				msg = "Cần cấp quyền Operator cho Tailscale: Chạy 'sudo tailscale set --operator=$USER' trên terminal một lần"
			}
			s.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": false, "error": msg, "message": msg})
			return
		}
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{"success": true, "message": "Đã đặt lại Tailscale Serve thành công"})

	case "sync":
		s.handleAPIRoutes(w, r, []string{"sync-tailscale"})
		return

	default:
		s.jsonResponse(w, http.StatusNotFound, map[string]interface{}{"error": "Hành động Tailscale không hợp lệ", "success": false})
	}
}

func (s *Server) handleAPIAutostart(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.jsonResponse(w, http.StatusOK, s.service.GetStatus())
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Enabled *bool  `json:"enabled"`
			Enable  *bool  `json:"enable"`
			Action  string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.jsonResponse(w, http.StatusBadRequest, map[string]interface{}{"success": false, "error": "JSON không hợp lệ", "message": "JSON không hợp lệ"})
			return
		}

		willEnable := false
		if req.Enabled != nil {
			willEnable = *req.Enabled
		} else if req.Enable != nil {
			willEnable = *req.Enable
		} else if req.Action == "enable" {
			willEnable = true
		}

		var err error
		if willEnable {
			err = s.service.Enable()
		} else {
			err = s.service.Disable()
		}

		status := s.service.GetStatus()
		if err != nil {
			s.jsonResponse(w, http.StatusInternalServerError, map[string]interface{}{
				"success":   false,
				"error":     err.Error(),
				"message":   err.Error(),
				"autostart": status,
			})
			return
		}
		msg := "Đã bật tự khởi động cùng hệ điều hành"
		if !willEnable {
			msg = "Đã tắt tự khởi động cùng hệ điều hành"
		}
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"success":   true,
			"message":   msg,
			"enabled":   status.Enabled,
			"autostart": status,
		})
		return
	}

	s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]interface{}{"error": "Phương thức không được hỗ trợ", "success": false})
}

func (s *Server) handleReverseProxy(w http.ResponseWriter, r *http.Request) {
	// 1. Loop Protection Check
	hopStr := r.Header.Get("X-TailRouter-Hop")
	hops := 0
	if hopStr != "" {
		hops, _ = strconv.Atoi(hopStr)
	}
	if hops >= 5 {
		http.Error(w, "508 Loop Detected: Phát hiện vòng lặp chuyển tiếp proxy", http.StatusLoopDetected)
		return
	}

	route, remainderPath, found := s.routes.FindMatchingRoute(r.URL.Path)
	if !found {
		http.NotFound(w, r)
		return
	}

	// 2. Strict Port Loop Protection
	if route.TargetPort == s.port || (route.TargetHost == "127.0.0.1" && route.TargetPort == s.port) {
		http.Error(w, "508 Loop Detected: Không được chuyển tiếp tới chính cổng Gateway 65534", http.StatusLoopDetected)
		return
	}

	targetURL, err := url.Parse(fmt.Sprintf("http://%s:%d", route.TargetHost, route.TargetPort))
	if err != nil {
		http.Error(w, "Cấu hình URL đích không hợp lệ", http.StatusInternalServerError)
		return
	}

	// Increment hits
	s.routes.IncrementHits(route.ID)

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	originalDirector := proxy.Director

	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		if route.StripPrefix {
			req.URL.Path = remainderPath
		}
		req.Host = fmt.Sprintf("%s:%d", route.TargetHost, route.TargetPort)
		req.Header.Set("X-Forwarded-Host", r.Host)
		req.Header.Set("X-Forwarded-Proto", "http")
		if r.TLS != nil {
			req.Header.Set("X-Forwarded-Proto", "https")
		}
		req.Header.Set("X-TailRouter-Hop", strconv.Itoa(hops+1))
	}

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		s.routes.UpdateStatus(route.ID, "offline", 0)
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(rw, `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>502 Bad Gateway - TailRouter</title>
<style>body{font-family:sans-serif;background:#0f172a;color:#f8fafc;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;}
.box{text-align:center;padding:2rem;background:#1e293b;border-radius:12px;border:1px solid #334155;max-width:500px;}
h1{color:#ef4444;margin:0 0 1rem;}p{color:#94a3b8;line-height:1.5;}</style></head>
<body><div class="box"><h1>502 Bad Gateway</h1>
<p>Dịch vụ đích <strong>%s:%d</strong> hiện không thể kết nối hoặc đã dừng hoạt động.</p>
<p style="font-size:0.875rem;color:#64748b;">TailRouter Native Gateway</p></div></body></html>`, route.TargetHost, route.TargetPort)
	}

	proxy.ServeHTTP(w, r)
}

func (s *Server) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
