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
	Version     = "2.0.0-go-native"
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
			s.jsonResponse(w, http.StatusOK, s.routes.List())
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
				s.jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Dữ liệu JSON không hợp lệ"})
				return
			}
			route, err := s.routes.Add(body.Name, body.Path, body.TargetHost, body.TargetPort, body.StripPrefix, body.Mode, body.Notes)
			if err != nil {
				s.jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}

			// Apply to Tailscale Serve/Funnel in background
			go func() {
				_ = s.tailscale.ApplyRoute(route.Path, route.TargetHost, route.TargetPort, route.Mode)
			}()

			s.jsonResponse(w, http.StatusCreated, route)
			return
		}
		s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Phương thức không được hỗ trợ"})
		return
	}

	routeID := subParts[0]

	if len(subParts) == 1 {
		// /api/routes/{id}
		if r.Method == http.MethodGet {
			if route, ok := s.routes.Get(routeID); ok {
				s.jsonResponse(w, http.StatusOK, route)
				return
			}
			s.jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Không tìm thấy route"})
			return
		}
		if r.Method == http.MethodPut {
			var updates map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
				s.jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "Dữ liệu JSON không hợp lệ"})
				return
			}
			updated, err := s.routes.Update(routeID, updates)
			if err != nil {
				s.jsonResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
				return
			}
			s.jsonResponse(w, http.StatusOK, updated)
			return
		}
		if r.Method == http.MethodDelete {
			route, err := s.routes.Delete(routeID)
			if err != nil {
				s.jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
				return
			}
			// Remove from Tailscale in background
			go func() {
				_ = s.tailscale.RemoveRoute(route.Path)
			}()
			s.jsonResponse(w, http.StatusOK, map[string]string{"message": "Đã xóa route thành công"})
			return
		}
	}

	// Actions on /api/routes/{id}/{action}
	action := subParts[1]
	switch action {
	case "toggle":
		if r.Method != http.MethodPost {
			s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Cần phương thức POST"})
			return
		}
		route, err := s.routes.Toggle(routeID)
		if err != nil {
			s.jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		s.jsonResponse(w, http.StatusOK, route)

	case "toggle-mode":
		if r.Method != http.MethodPost {
			s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Cần phương thức POST"})
			return
		}
		route, err := s.routes.ToggleMode(routeID)
		if err != nil {
			s.jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		go func() {
			_ = s.tailscale.ApplyRoute(route.Path, route.TargetHost, route.TargetPort, route.Mode)
		}()
		s.jsonResponse(w, http.StatusOK, route)

	case "ping":
		if r.Method != http.MethodPost {
			s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Cần phương thức POST"})
			return
		}
		status, latency, err := s.routes.PingRoute(routeID)
		if err != nil {
			s.jsonResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		s.jsonResponse(w, http.StatusOK, map[string]interface{}{
			"status":     status,
			"latency_ms": latency,
		})

	default:
		s.jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Hành động không hợp lệ"})
	}
}

func (s *Server) handleAPIScan(w http.ResponseWriter, r *http.Request) {
	ports, err := s.scanner.Scan()
	if err != nil {
		s.jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.jsonResponse(w, http.StatusOK, map[string]interface{}{
		"ports": ports,
		"count": len(ports),
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
			s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Cần phương thức POST"})
			return
		}
		var req struct {
			Target string `json:"target"` // "router" or "gateway"
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		var err error
		if req.Target == "gateway" {
			err = s.tailscale.ConfigureServeGateway(s.port)
		} else {
			err = s.tailscale.ConfigureServeRouter(s.port)
		}
		if err != nil {
			s.jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		s.jsonResponse(w, http.StatusOK, map[string]string{"message": "Đã cấu hình Tailscale Serve thành công"})

	case "reset":
		if r.Method != http.MethodPost {
			s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Cần phương thức POST"})
			return
		}
		if err := s.tailscale.ResetServe(); err != nil {
			s.jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		s.jsonResponse(w, http.StatusOK, map[string]string{"message": "Đã đặt lại Tailscale Serve thành công"})

	default:
		s.jsonResponse(w, http.StatusNotFound, map[string]string{"error": "Hành động Tailscale không hợp lệ"})
	}
}

func (s *Server) handleAPIAutostart(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.jsonResponse(w, http.StatusOK, s.service.GetStatus())
		return
	}

	if r.Method == http.MethodPost {
		var req struct {
			Enabled bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.jsonResponse(w, http.StatusBadRequest, map[string]string{"error": "JSON không hợp lệ"})
			return
		}

		var err error
		if req.Enabled {
			err = s.service.Enable()
		} else {
			err = s.service.Disable()
		}

		if err != nil {
			s.jsonResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		s.jsonResponse(w, http.StatusOK, s.service.GetStatus())
		return
	}

	s.jsonResponse(w, http.StatusMethodNotAllowed, map[string]string{"error": "Phương thức không được hỗ trợ"})
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
