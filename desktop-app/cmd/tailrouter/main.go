package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/giangsamne/TailRouter/internal/config"
	"github.com/giangsamne/TailRouter/internal/gateway"
	"github.com/giangsamne/TailRouter/internal/scanner"
	"github.com/giangsamne/TailRouter/internal/service"
	"github.com/giangsamne/TailRouter/internal/tailscale"
)

const Banner = `
  ⚡ TailRouter Native Engine v2.0
  ==================================
  Zero-Dependency Bare-Metal Gateway
`

func main() {
	if len(os.Args) < 2 {
		runServer(gateway.DefaultPort, "")
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "run", "start", "daemon":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		portFlag := runCmd.Int("port", gateway.DefaultPort, "Cổng lắng nghe của Gateway")
		configFlag := runCmd.String("config", "", "Đường dẫn file cấu hình routes.json")
		_ = runCmd.Parse(os.Args[2:])
		runServer(*portFlag, *configFlag)

	case "scan":
		runScan()

	case "status":
		runStatus()

	case "routes":
		runRoutes(os.Args[2:])

	case "serve":
		runServe(os.Args[2:])

	case "service":
		runService(os.Args[2:])

	case "version", "-v", "--version":
		fmt.Printf("TailRouter Native Engine v%s (%s/%s)\n", gateway.Version, os.Getenv("GOOS"), os.Getenv("GOARCH"))

	case "help", "-h", "--help":
		printHelp()

	default:
		// If argument is a port number like `tailrouter 65534`
		if p, err := strconv.Atoi(cmd); err == nil && p > 0 && p < 65536 {
			runServer(p, "")
			return
		}
		fmt.Printf("Lệnh không xác định: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(Banner)
	fmt.Println(`
Sử dụng: tailrouter [lệnh] [tùy chọn]

Các lệnh có sẵn:
  run [--port 65534]          Khởi chạy máy chủ Gateway (mặc định)
  scan                        Quét toàn bộ cổng TCP & Docker đang mở trên máy thật
  status                      Kiểm tra trạng thái hoạt động của Gateway & Tailscale
  routes list                 Xem danh sách các tuyến đường đã cấu hình
  routes add <tên> <path> <port> [serve|funnel]
                              Thêm tuyến đường chuyển tiếp mới
  routes delete <id>          Xóa tuyến đường theo ID
  serve [router|gateway|reset]
                              Cấu hình Tailscale Serve tự động
  service [install|uninstall|status]
                              Quản lý tự khởi động cùng hệ điều hành (systemd/launchd)
  version                     Xem thông tin phiên bản
  help                        Hiển thị trợ giúp này
`)
}

func runServer(port int, configPath string) {
	fmt.Print(Banner)
	srv := gateway.NewServer(port, configPath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n[TailRouter] Đang dừng Gateway an toàn...")
		srv.Close()
		os.Exit(0)
	}()

	if err := srv.Start(); err != nil {
		fmt.Printf("[Lỗi] %v\n", err)
		os.Exit(1)
	}
}

func runScan() {
	fmt.Println("🔍 Đang quét các cổng TCP & Docker trên máy thật...")
	sc := scanner.NewScanner()
	ports, err := sc.Scan()
	if err != nil {
		fmt.Printf("Lỗi quét cổng: %v\n", err)
		return
	}

	if len(ports) == 0 {
		fmt.Println("Không tìm thấy cổng nào đang mở.")
		return
	}

	fmt.Printf("\n%-7s %-20s %-25s %-20s\n", "CỔNG", "DỊCH VỤ", "MÔ TẢ / CONTAINER", "GỢI Ý PATH")
	fmt.Println("-------------------------------------------------------------------------------")
	for _, p := range ports {
		desc := p.Description
		if desc == "" {
			desc = p.Title
		}
		if len(desc) > 24 {
			desc = desc[:21] + "..."
		}
		fmt.Printf("%-7d %-20s %-25s %-20s\n", p.Port, p.Service, desc, p.SuggestedPath)
	}
	fmt.Printf("\nTổng cộng: %d cổng đang mở.\n", len(ports))
}

func runStatus() {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/status", gateway.DefaultPort))
	if err != nil {
		fmt.Printf("⚠️  TailRouter Gateway hiện KHÔNG CHẠY trên cổng %d.\n", gateway.DefaultPort)
		// Check local tailscale directly
		ts := tailscale.NewClient()
		tsStat, _ := ts.GetStatus()
		if tsStat.Running {
			fmt.Printf("   Tailscale: Đang chạy (%s, IP: %v)\n", tsStat.NodeName, tsStat.IPs)
		} else {
			fmt.Println("   Tailscale: Chưa chạy hoặc chưa cài đặt.")
		}
		return
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&data)

	fmt.Println("🟢 TailRouter Gateway: Đang chạy (Active)")
	fmt.Printf("   Phiên bản: %v\n", data["version"])
	fmt.Printf("   Uptime: %v giây\n", data["uptime_seconds"])
	fmt.Printf("   Số tuyến đường (Routes): %v\n", data["routes_count"])
	fmt.Printf("   Dashboard: http://localhost:%d/router\n", gateway.DefaultPort)

	if ts, ok := data["tailscale"].(map[string]interface{}); ok && ts["running"] == true {
		fmt.Printf("   Tailscale Node: %v (FQDN: %v)\n", ts["node_name"], ts["fqdn"])
	}
}

func runRoutes(args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}

	subCmd := args[0]
	apiURL := fmt.Sprintf("http://127.0.0.1:%d/api", gateway.DefaultPort)
	client := &http.Client{Timeout: 3 * time.Second}
	isDaemonRunning := false
	if resp, err := client.Get(apiURL + "/status"); err == nil && resp.StatusCode == http.StatusOK {
		isDaemonRunning = true
		_ = resp.Body.Close()
	}

	switch subCmd {
	case "list":
		var routes []*config.Route
		if isDaemonRunning {
			resp, err := client.Get(apiURL + "/routes")
			if err == nil && resp.StatusCode == http.StatusOK {
				_ = json.NewDecoder(resp.Body).Decode(&routes)
				_ = resp.Body.Close()
			}
		} else {
			mgr := config.NewManager("")
			routes = mgr.List()
		}

		if len(routes) == 0 {
			fmt.Println("Chưa có tuyến đường nào. Sử dụng 'tailrouter routes add' để tạo mới.")
			return
		}
		fmt.Printf("%-24s %-20s %-15s %-18s %-8s %-8s\n", "ID", "TÊN", "PATH", "ĐÍCH", "CHẾ ĐỘ", "TRẠNG THÁI")
		fmt.Println("----------------------------------------------------------------------------------------------------")
		for _, r := range routes {
			target := fmt.Sprintf("%s:%d", r.TargetHost, r.TargetPort)
			status := "Bật"
			if !r.Enabled {
				status = "Tắt"
			}
			fmt.Printf("%-24s %-20s %-15s %-18s %-8s %-8s\n", r.ID, r.Name, r.Path, target, r.Mode, status)
		}

	case "add":
		if len(args) < 4 {
			fmt.Println("Cú pháp: tailrouter routes add <tên> <path> <port> [serve|funnel]")
			return
		}
		name := args[1]
		path := args[2]
		port, err := strconv.Atoi(args[3])
		if err != nil {
			fmt.Println("Cổng không hợp lệ")
			return
		}
		mode := "serve"
		if len(args) >= 5 {
			mode = args[4]
		}

		if isDaemonRunning {
			payload := map[string]interface{}{
				"name":        name,
				"path":        path,
				"target_host": "127.0.0.1",
				"target_port": port,
				"mode":        mode,
				"notes":       "Thêm qua CLI",
			}
			bodyBytes, _ := json.Marshal(payload)
			resp, err := client.Post(apiURL+"/routes", "application/json", bytes.NewBuffer(bodyBytes))
			if err != nil {
				fmt.Printf("Lỗi kết nối tới Gateway: %v\n", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				var errResp map[string]string
				_ = json.NewDecoder(resp.Body).Decode(&errResp)
				fmt.Printf("Lỗi: %s\n", errResp["error"])
				return
			}
			var r config.Route
			_ = json.NewDecoder(resp.Body).Decode(&r)
			fmt.Printf("✅ Đã tạo route thành công: %s -> %s:%d (%s)\n", r.Path, r.TargetHost, r.TargetPort, r.Mode)
		} else {
			mgr := config.NewManager("")
			r, err := mgr.Add(name, path, "127.0.0.1", port, true, mode, "Thêm qua CLI")
			if err != nil {
				fmt.Printf("Lỗi: %v\n", err)
				return
			}
			fmt.Printf("✅ Đã tạo route thành công: %s -> %s:%d (%s)\n", r.Path, r.TargetHost, r.TargetPort, r.Mode)
		}

	case "delete", "remove", "rm":
		if len(args) < 2 {
			fmt.Println("Cú pháp: tailrouter routes delete <id hoặc path>")
			return
		}
		target := args[1]
		cleanTarget := strings.Trim(target, "/")
		if isDaemonRunning {
			req, _ := http.NewRequest(http.MethodDelete, apiURL+"/routes/"+url.PathEscape(cleanTarget), nil)
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Lỗi kết nối tới Gateway: %v\n", err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				var errResp map[string]string
				_ = json.NewDecoder(resp.Body).Decode(&errResp)
				fmt.Printf("Lỗi: %s\n", errResp["error"])
				return
			}
			fmt.Printf("✅ Đã xóa route: %s\n", target)
		} else {
			mgr := config.NewManager("")
			r, err := mgr.Delete(target)
			if err != nil {
				fmt.Printf("Lỗi: %v\n", err)
				return
			}
			fmt.Printf("✅ Đã xóa route: %s (%s)\n", r.Name, r.Path)
		}

	default:
		fmt.Printf("Lệnh routes không hợp lệ: %s\n", subCmd)
	}
}

func runServe(args []string) {
	if len(args) == 0 {
		args = []string{"router"}
	}
	ts := tailscale.NewClient()
	sub := args[0]
	switch sub {
	case "router":
		if err := ts.ConfigureServeRouter(gateway.DefaultPort); err != nil {
			fmt.Printf("Lỗi: %v\n", err)
			return
		}
		fmt.Println("✅ Đã bật Tailscale Serve cho /router")
	case "gateway":
		if err := ts.ConfigureServeGateway(gateway.DefaultPort); err != nil {
			fmt.Printf("Lỗi: %v\n", err)
			return
		}
		fmt.Println("✅ Đã ánh xạ toàn bộ Gateway qua Tailscale Serve")
	case "reset":
		if err := ts.ResetServe(); err != nil {
			fmt.Printf("Lỗi: %v\n", err)
			return
		}
		fmt.Println("✅ Đã đặt lại cấu hình Tailscale Serve")
	default:
		fmt.Println("Cú pháp: tailrouter serve [router|gateway|reset]")
	}
}

func runService(args []string) {
	if len(args) == 0 {
		args = []string{"status"}
	}
	mgr := service.NewManager()
	switch args[0] {
	case "install", "enable":
		if err := mgr.Enable(); err != nil {
			fmt.Printf("Lỗi: %v\n", err)
			return
		}
		fmt.Println("✅ Đã cài đặt tự khởi động cùng hệ điều hành thành công.")
	case "uninstall", "disable":
		if err := mgr.Disable(); err != nil {
			fmt.Printf("Lỗi: %v\n", err)
			return
		}
		fmt.Println("✅ Đã gỡ bỏ tự khởi động.")
	case "status":
		stat := mgr.GetStatus()
		fmt.Printf("Phương thức: %s (%s)\n", stat.Method, stat.Platform)
		fmt.Printf("Tự khởi động: %v\n", stat.Enabled)
		fmt.Printf("Đang hoạt động: %v\n", stat.Active)
	default:
		fmt.Println("Cú pháp: tailrouter service [install|uninstall|status]")
	}
}
