package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/giangsamne/TailRouter/internal/config"
	"github.com/giangsamne/TailRouter/internal/gateway"
	"github.com/giangsamne/TailRouter/internal/scanner"
	"github.com/giangsamne/TailRouter/internal/service"
	"github.com/giangsamne/TailRouter/internal/tailscale"
)

const Banner = `
  ⚡ TailRouter Desktop Edition v2.0
  ==================================
  Zero-Dependency Bare-Metal Gateway & Tray App
`

func runCLI() {
	cmd := os.Args[1]
	switch cmd {
	case "run", "start", "daemon":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		portFlag := runCmd.Int("port", gateway.DefaultPort, "Cổng lắng nghe của Gateway")
		configFlag := runCmd.String("config", "", "Đường dẫn file cấu hình routes.json")
		_ = runCmd.Parse(os.Args[2:])
		runServerCLI(*portFlag, *configFlag)

	case "scan":
		runScanCLI()

	case "status":
		runStatusCLI()

	case "routes":
		runRoutesCLI(os.Args[2:])

	case "serve":
		runServeCLI(os.Args[2:])

	case "service":
		runServiceCLI(os.Args[2:])

	case "version", "-v", "--version":
		fmt.Printf("TailRouter Desktop Edition v%s (%s/%s)\n", gateway.Version, os.Getenv("GOOS"), os.Getenv("GOARCH"))

	case "help", "-h", "--help":
		printHelpCLI()

	default:
		if p, err := strconv.Atoi(cmd); err == nil && p > 0 && p < 65536 {
			runServerCLI(p, "")
			return
		}
		fmt.Printf("Lệnh không xác định: %s\n", cmd)
		printHelpCLI()
		os.Exit(1)
	}
}

func printHelpCLI() {
	fmt.Print(Banner)
	fmt.Println(`
Sử dụng: TailRouter [lệnh] [tùy chọn]

Các lệnh CLI:
  scan                        Quét cổng đang mở trên máy thật
  status                      Kiểm tra trạng thái hoạt động của Gateway & Tailscale
  routes list                 Xem danh sách các tuyến đường
  routes add <tên> <path> <port> [serve|funnel]
                              Thêm tuyến đường mới
  routes delete <id>          Xóa tuyến đường theo ID
  serve [router|gateway|reset]
                              Cấu hình Tailscale Serve
  service [install|uninstall|status]
                              Quản lý tự khởi động cùng máy
  version                     Xem thông tin phiên bản
`)
}

func runServerCLI(port int, configPath string) {
	fmt.Print(Banner)
	srv := gateway.NewServer(port, configPath)
	if err := srv.Start(); err != nil {
		fmt.Printf("[Lỗi] %v\n", err)
		os.Exit(1)
	}
}

func runScanCLI() {
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

func runStatusCLI() {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/status", gateway.DefaultPort))
	if err != nil {
		fmt.Printf("⚠️  TailRouter Gateway hiện KHÔNG CHẠY trên cổng %d.\n", gateway.DefaultPort)
		return
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&data)

	fmt.Println("🟢 TailRouter Gateway: Đang chạy (Active)")
	fmt.Printf("   Phiên bản: %v\n", data["version"])
	fmt.Printf("   Uptime: %v giây\n", data["uptime_seconds"])
	fmt.Printf("   Số tuyến đường: %v\n", data["routes_count"])
	fmt.Printf("   Dashboard: http://localhost:%d/router\n", gateway.DefaultPort)
}

func runRoutesCLI(args []string) {
	if len(args) == 0 {
		args = []string{"list"}
	}
	mgr := config.NewManager("")
	switch args[0] {
	case "list":
		routes := mgr.List()
		fmt.Printf("%-20s %-15s %-20s %-8s %-8s\n", "TÊN", "PATH", "ĐÍCH", "CHẾ ĐỘ", "TRẠNG THÁI")
		fmt.Println("---------------------------------------------------------------------------")
		for _, r := range routes {
			target := fmt.Sprintf("%s:%d", r.TargetHost, r.TargetPort)
			status := "Bật"
			if !r.Enabled {
				status = "Tắt"
			}
			fmt.Printf("%-20s %-15s %-20s %-8s %-8s\n", r.Name, r.Path, target, r.Mode, status)
		}
	}
}

func runServeCLI(args []string) {
	if len(args) == 0 {
		args = []string{"router"}
	}
	ts := tailscale.NewClient()
	switch args[0] {
	case "router":
		_ = ts.ConfigureServeRouter(gateway.DefaultPort)
		fmt.Println("✅ Đã bật Tailscale Serve cho /router")
	case "gateway":
		_ = ts.ConfigureServeGateway(gateway.DefaultPort)
		fmt.Println("✅ Đã ánh xạ toàn bộ Gateway qua Tailscale Serve")
	case "reset":
		_ = ts.ResetServe()
		fmt.Println("✅ Đã đặt lại cấu hình Tailscale Serve")
	}
}

func runServiceCLI(args []string) {
	mgr := service.NewManager()
	if len(args) > 0 && (args[0] == "install" || args[0] == "enable") {
		_ = mgr.Enable()
		fmt.Println("✅ Đã cài đặt tự khởi động cùng hệ thống.")
	} else if len(args) > 0 && (args[0] == "uninstall" || args[0] == "disable") {
		_ = mgr.Disable()
		fmt.Println("✅ Đã gỡ bỏ tự khởi động.")
	} else {
		stat := mgr.GetStatus()
		fmt.Printf("Tự khởi động: %v\n", stat.Enabled)
	}
}
