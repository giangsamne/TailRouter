package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/giangsamne/TailRouter/internal/browser"
	"github.com/giangsamne/TailRouter/internal/config"
	"github.com/giangsamne/TailRouter/internal/gateway"
	"github.com/giangsamne/TailRouter/internal/scanner"
	"github.com/giangsamne/TailRouter/internal/service"
	"github.com/giangsamne/TailRouter/internal/tailscale"
)

const Banner = `
  ⚡ TailRouter Native Engine v2.1.0
  ==================================
  Zero-Dependency Bare-Metal Gateway
`

func main() {
	if len(os.Args) < 2 {
		// Khi chạy trực tiếp không có tham số (hoặc click đúp file): mở server và tự động bật Web Dashboard
		runServer(gateway.DefaultPort, "", true)
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "run", "start", "daemon":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		portFlag := runCmd.Int("port", gateway.DefaultPort, "Cổng lắng nghe của Gateway")
		configFlag := runCmd.String("config", "", "Đường dẫn file cấu hình routes.json")
		openFlag := runCmd.Bool("open", false, "Tự động mở Web Dashboard trên trình duyệt")
		_ = runCmd.Parse(os.Args[2:])
		runServer(*portFlag, *configFlag, *openFlag)

	case "stop", "kill", "down":
		runStop(gateway.DefaultPort)

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

	case "uninstall", "remove", "purge":
		runUninstall(os.Args[2:])

	case "version", "-v", "--version":
		fmt.Printf("TailRouter Native Engine v%s (%s/%s)\n", gateway.Version, os.Getenv("GOOS"), os.Getenv("GOARCH"))

	case "help", "-h", "--help":
		printHelp()

	default:
		// If argument is a port number like `tailrouter 65534`
		if p, err := strconv.Atoi(cmd); err == nil && p > 0 && p < 65536 {
			runServer(p, "", true)
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
  (không tham số)             Khởi chạy máy chủ và tự động mở Web Dashboard
  run [--port 65534] [--open] Khởi chạy máy chủ Gateway
  stop                        Dừng dịch vụ Gateway đang chạy
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
  uninstall [-y]              Gỡ cài đặt và xóa sạch toàn bộ dữ liệu TailRouter khỏi máy
  version                     Xem thông tin phiên bản
  help                        Hiển thị trợ giúp này
`)
}

func findPIDByPort(port int) int {
	// 1. Thử fuser (nhanh và chuẩn xác nhất trên Linux)
	if out, err := exec.Command("fuser", fmt.Sprintf("%d/tcp", port)).CombinedOutput(); err == nil {
		fields := strings.Fields(strings.TrimSpace(string(out)))
		for _, f := range fields {
			if p, err := strconv.Atoi(f); err == nil && p > 0 {
				return p
			}
		}
	}
	// 2. Thử ss -ltnp
	if out, err := exec.Command("ss", "-ltnp", fmt.Sprintf("sport = :%d", port)).Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "pid=") {
				idx := strings.Index(line, "pid=")
				sub := line[idx+4:]
				end := strings.IndexAny(sub, ",)")
				if end != -1 {
					sub = sub[:end]
				}
				if p, err := strconv.Atoi(sub); err == nil && p > 0 {
					return p
				}
			}
		}
	}
	// 3. Thử ss -tulpn chung
	if out, err := exec.Command("ss", "-tulpn").Output(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, fmt.Sprintf(":%d ", port)) && strings.Contains(line, "pid=") {
				idx := strings.Index(line, "pid=")
				sub := line[idx+4:]
				end := strings.IndexAny(sub, ",)")
				if end != -1 {
					sub = sub[:end]
				}
				if p, err := strconv.Atoi(sub); err == nil && p > 0 {
					return p
				}
			}
		}
	}
	// 4. Thử lsof lắng nghe TCP
	if out, err := exec.Command("lsof", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-t").Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if p, err := strconv.Atoi(strings.TrimSpace(line)); err == nil && p > 0 {
				return p
			}
		}
	}
	return 0
}

func checkPortOccupied(port int) (bool, int) {
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/api/status", port))
	if err == nil {
		defer resp.Body.Close()
		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			if p, ok := data["pid"].(float64); ok && p > 0 {
				return true, int(p)
			}
		}
		pid := findPIDByPort(port)
		return true, pid
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 500*time.Millisecond)
	if err == nil {
		_ = conn.Close()
		pid := findPIDByPort(port)
		return true, pid
	}
	return false, 0
}

func runStop(port int) {
	occupied, pid := checkPortOccupied(port)
	if !occupied && pid <= 0 {
		fmt.Print(Banner)
		fmt.Printf("⚠️  TailRouter Gateway hiện KHÔNG CHẠY trên cổng %d.\n", port)
		return
	}
	if pid <= 0 {
		pid = findPIDByPort(port)
	}
	if pid > 0 {
		fmt.Printf("🛑 Đang dừng tiến trình TailRouter (PID: %d)...\n", pid)
		p, err := os.FindProcess(pid)
		if err == nil {
			_ = p.Signal(syscall.SIGTERM)
			for i := 0; i < 6; i++ {
				time.Sleep(500 * time.Millisecond)
				if err := p.Signal(syscall.Signal(0)); err != nil {
					break
				}
			}
			_ = p.Signal(syscall.SIGKILL)
		}
		fmt.Println("✅ Đã dừng thành công dịch vụ TailRouter!")
	} else {
		_ = exec.Command("pkill", "-f", "tailrouter").Run()
		fmt.Println("✅ Đã dừng thành công dịch vụ TailRouter!")
	}
}

func runServer(port int, configPath string, openBrowser bool) {
	occupied, pid := checkPortOccupied(port)
	if occupied {
		fmt.Print(Banner)
		if pid > 0 {
			fmt.Printf("⚠️  Cổng %d đã có tiến trình đang chạy (PID: %d).\n", port, pid)
		} else {
			fmt.Printf("⚠️  Cổng %d đã có tiến trình đang chạy.\n", port)
		}
		fmt.Printf("👉 Truy cập ngay tại: http://localhost:%d/router\n", port)
		ts := tailscale.NewClient()
		if tsStat, err := ts.GetStatus(); err == nil && tsStat.Running && tsStat.FQDN != "" {
			fmt.Printf("👉 Tailscale Serve:   https://%s/router\n", tsStat.FQDN)
		}
		return
	}

	fmt.Print(Banner)
	srv := gateway.NewServer(port, configPath)

	if openBrowser {
		go func() {
			time.Sleep(500 * time.Millisecond)
			dashboardURL := fmt.Sprintf("http://localhost:%d/router", port)
			_ = browser.Open(dashboardURL)
		}()
	}

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

	fmt.Printf("\n%-7s %-20s %-20s %-28s %-20s\n", "CỔNG", "PHẦN MỀM / TIẾN TRÌNH", "DỊCH VỤ", "MÔ TẢ / CONTAINER", "GỢI Ý PATH")
	fmt.Println("---------------------------------------------------------------------------------------------------------")
	for _, p := range ports {
		proc := p.Process
		if proc == "" && p.Container != "" {
			proc = "docker:" + p.Container
		}
		if proc == "" {
			proc = "-"
		}
		if p.PID > 0 {
			proc = fmt.Sprintf("%s (%d)", proc, p.PID)
		}
		if len(proc) > 19 {
			proc = proc[:16] + "..."
		}

		desc := p.Description
		if desc == "" {
			desc = p.Title
		}
		if len(desc) > 27 {
			desc = desc[:24] + "..."
		}
		service := p.Service
		if len(service) > 19 {
			service = service[:16] + "..."
		}

		fmt.Printf("%-7d %-20s %-20s %-28s %-20s\n", p.Port, proc, service, desc, p.SuggestedPath)
	}
	fmt.Printf("\nTổng cộng: %d cổng đang mở (đã sắp xếp từ bé đến lớn).\n", len(ports))
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
		os.Exit(1)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&data)

	fmt.Println("🟢 TailRouter Gateway: Đang chạy (Active)")
	fmt.Printf("   Phiên bản: %v\n", data["version"])
	fmt.Printf("   Uptime: %v giây\n", data["uptime_seconds"])
	fmt.Printf("   Số tuyến đường (Routes): %v\n", data["routes_count"])
	fmt.Printf("   Dashboard: http://localhost:%d\n", gateway.DefaultPort)

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
	case "router", "gateway":
		if err := ts.ConfigureServeGateway(gateway.DefaultPort); err != nil {
			fmt.Printf("Lỗi: %v\n", err)
			return
		}
		fmt.Println("✅ Đã bật Tailscale Serve cho TailRouter")
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

func runUninstall(args []string) {
	force := false
	for _, a := range args {
		if a == "-y" || a == "--yes" || a == "-f" || a == "--force" {
			force = true
			break
		}
	}

	fmt.Print(Banner)
	if !force {
		fmt.Println("⚠️  CẢNH BÁO: BẠN CÓ CHẮC CHẮN MUỐN GỠ CÀI ĐẶT TAILROUTER KHỎI MÁY?")
		fmt.Println("   Thao tác này sẽ xóa sạch cấu hình, nhật ký log, dịch vụ chạy ngầm,")
		fmt.Println("   đặt lại Tailscale Serve và xóa bỏ file thực thi tailrouter.")
		fmt.Print("\n👉 Bạn có muốn tiếp tục? (y/N): ")
		var resp string
		_, _ = fmt.Scanln(&resp)
		resp = strings.ToLower(strings.TrimSpace(resp))
		if resp != "y" && resp != "yes" {
			fmt.Println("Đã hủy bỏ thao tác gỡ cài đặt.")
			return
		}
	}

	fmt.Println("\n💥 Đang tiến hành gỡ cài đặt và dọn dẹp sạch sẽ...")

	// 1. Dừng daemon nếu đang chạy
	occupied, pid := checkPortOccupied(gateway.DefaultPort)
	if occupied || pid > 0 {
		fmt.Println("🛑 Đang dừng tiến trình TailRouter...")
		runStop(gateway.DefaultPort)
	}

	// 2. Đặt lại Tailscale Serve
	fmt.Println("🌐 Đang đặt lại cấu hình Tailscale Serve...")
	ts := tailscale.NewClient()
	_ = ts.ResetServe()

	// 3. Vô hiệu hóa service autostart
	fmt.Println("🚀 Đang gỡ bỏ dịch vụ tự khởi động cùng hệ thống...")
	svcMgr := service.NewManager()
	_ = svcMgr.Disable()

	// 4. Xóa cấu hình
	fmt.Println("🗑️  Đang xóa cấu hình routes.json và thư mục dữ liệu...")
	cfgDir := filepath.Dir(config.DefaultConfigPath())
	if cfgDir != "" && cfgDir != "/" && cfgDir != "." {
		_ = os.RemoveAll(cfgDir)
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		_ = os.RemoveAll(filepath.Join(home, ".config", "tailrouter"))
		_ = os.RemoveAll(filepath.Join(home, ".tailrouter"))
		_ = os.Remove(filepath.Join(home, ".tailrouter.log"))
	}

	// 5. Xóa log
	execDir, err := os.Executable()
	if err == nil {
		execDir = filepath.Dir(execDir)
		_ = os.Remove(filepath.Join(execDir, "tailrouter.log"))
		_ = os.Remove(filepath.Join(execDir, "tailrouter.log.1"))
	}
	_ = os.Remove("/tmp/tailrouter.err.log")
	_ = os.Remove("/tmp/tailrouter.out.log")

	// 6. Xóa binary
	fmt.Println("📦 Đang gỡ bỏ file thực thi TailRouter...")
	_ = svcMgr.UninstallBinary()

	fmt.Println("\n✅ Đã hoàn tất gỡ cài đặt TailRouter!")
	fmt.Println("🎉 Toàn bộ dữ liệu, dịch vụ và file thực thi đã được xóa sạch khỏi máy.")
}
