import AppKit
import Foundation

class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem!
    var timer: Timer?
    var serverProcess: Process?
    var baseDir: String = ""
    
    // Menu items
    let statusMenuItem = NSMenuItem(title: "🟡 Đang kết nối...", action: nil, keyEquivalent: "")
    let ipMenuItem = NSMenuItem(title: "🌐 IP: Đang kiểm tra...", action: nil, keyEquivalent: "")
    let routesMenuItem = NSMenuItem(title: "⚡ Tuyến đường: 0", action: nil, keyEquivalent: "")
    let autostartMenuItem = NSMenuItem(title: "✓ Tự khởi động cùng Mac", action: #selector(toggleAutostart), keyEquivalent: "s")
    let toggleServerMenuItem = NSMenuItem(title: "⏸ Dừng Gateway", action: #selector(toggleServer), keyEquivalent: "")
    
    var isServerRunning = false
    var isAutostartEnabled = false

    func applicationDidFinishLaunching(_ notification: Notification) {
        // Xác định thư mục chứa mã nguồn server.py
        let appBundlePath = Bundle.main.bundlePath
        let resourcesPath = Bundle.main.resourcePath ?? appBundlePath
        
        if FileManager.default.fileExists(atPath: "\(resourcesPath)/server.py") {
            baseDir = resourcesPath
        } else if FileManager.default.fileExists(atPath: "\(appBundlePath)/server.py") {
            baseDir = appBundlePath
        } else {
            let userHome = FileManager.default.homeDirectoryForCurrentUser.path
            let candidate = "\(userHome)/TailRouter"
            if FileManager.default.fileExists(atPath: "\(candidate)/server.py") {
                baseDir = candidate
            } else {
                baseDir = FileManager.default.currentDirectoryPath
            }
        }

        // Tạo Menu Bar Item
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        if let button = statusItem.button {
            if let image = NSImage(systemSymbolName: "arrow.triangle.swap", accessibilityDescription: "TailRouter") {
                image.isTemplate = true
                button.image = image
            } else {
                button.title = "⛭ TailRouter"
            }
            button.toolTip = "TailRouter - Tailscale Port Router & Gateway"
        }

        buildMenu()
        
        // Khởi động server nếu chưa chạy
        startServerInBackgroundIfNeeded()

        // Kiểm tra trạng thái định kỳ 3 giây/lần
        updateStatus()
        timer = Timer.scheduledTimer(withTimeInterval: 3.0, repeats: true) { [weak self] _ in
            self?.updateStatus()
        }
    }

    func buildMenu() {
        let menu = NSMenu()
        
        statusMenuItem.isEnabled = false
        menu.addItem(statusMenuItem)
        
        ipMenuItem.isEnabled = false
        menu.addItem(ipMenuItem)
        
        routesMenuItem.isEnabled = false
        menu.addItem(routesMenuItem)
        
        menu.addItem(NSMenuItem.separator())
        
        let openDashboardItem = NSMenuItem(title: "🚀 Mở Bảng Quản Trị (/router)", action: #selector(openDashboard), keyEquivalent: "o")
        openDashboardItem.target = self
        menu.addItem(openDashboardItem)
        
        let scanItem = NSMenuItem(title: "🔍 Quét Cổng Đang Mở", action: #selector(openScan), keyEquivalent: "")
        scanItem.target = self
        menu.addItem(scanItem)
        
        menu.addItem(NSMenuItem.separator())
        
        autostartMenuItem.target = self
        menu.addItem(autostartMenuItem)
        
        toggleServerMenuItem.target = self
        menu.addItem(toggleServerMenuItem)
        
        let restartItem = NSMenuItem(title: "🔄 Khởi động lại Gateway", action: #selector(restartServer), keyEquivalent: "r")
        restartItem.target = self
        menu.addItem(restartItem)
        
        menu.addItem(NSMenuItem.separator())
        
        let helpItem = NSMenuItem(title: "📖 Xem GitHub & Hướng Dẫn", action: #selector(openHelp), keyEquivalent: "")
        helpItem.target = self
        menu.addItem(helpItem)
        
        let quitItem = NSMenuItem(title: "❌ Thoát TailRouter", action: #selector(quitApp), keyEquivalent: "q")
        quitItem.target = self
        menu.addItem(quitItem)
        
        statusItem.menu = menu
    }

    func updateStatus() {
        guard let url = URL(string: "http://127.0.0.1:65534/api/status") else { return }
        
        var request = URLRequest(url: url)
        request.timeoutInterval = 2.0
        
        let task = URLSession.shared.dataTask(with: request) { [weak self] data, response, error in
            DispatchQueue.main.async {
                guard let self = self else { return }
                
                if let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200, let data = data {
                    self.isServerRunning = true
                    if let button = self.statusItem.button {
                        if let img = NSImage(systemSymbolName: "arrow.triangle.swap", accessibilityDescription: "TailRouter") {
                            img.isTemplate = true
                            button.image = img
                        }
                    }
                    self.toggleServerMenuItem.title = "⏸ Tạm Dừng Gateway"
                    
                    if let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] {
                        let activeRoutes = json["active_routes"] as? Int ?? 0
                        let totalRoutes = json["total_routes"] as? Int ?? 0
                        self.routesMenuItem.title = "⚡ Tuyến đường: \(activeRoutes)/\(totalRoutes) hoạt động"
                        
                        if let ts = json["tailscale"] as? [String: Any] {
                            let ip = ts["ipv4"] as? String ?? "N/A"
                            let fqdn = ts["fqdn"] as? String ?? ""
                            self.ipMenuItem.title = "🌐 IP: \(ip) (\(fqdn.isEmpty ? "Local" : "Tailnet"))"
                        }
                        
                        if let auto = json["autostart"] as? [String: Any] {
                            self.isAutostartEnabled = auto["enabled"] as? Bool ?? false
                            self.autostartMenuItem.state = self.isAutostartEnabled ? .on : .off
                        }
                        
                        self.statusMenuItem.title = "🟢 TailRouter: Đang chạy (Cổng 65534)"
                    }
                } else {
                    self.isServerRunning = false
                    self.statusMenuItem.title = "🔴 TailRouter: Đã dừng"
                    self.ipMenuItem.title = "🌐 IP: Offline"
                    self.routesMenuItem.title = "⚡ Tuyến đường: Offline"
                    self.toggleServerMenuItem.title = "▶️ Khởi Động Gateway"
                    if let button = self.statusItem.button {
                        if let img = NSImage(systemSymbolName: "exclamationmark.triangle", accessibilityDescription: "TailRouter Stopped") {
                            img.isTemplate = true
                            button.image = img
                        }
                    }
                }
            }
        }
        task.resume()
    }

    func startServerInBackgroundIfNeeded() {
        // Kiểm tra xem cổng 65534 đã có tiến trình nào lắng nghe chưa
        let check = Process()
        check.executableURL = URL(fileURLWithPath: "/usr/sbin/lsof")
        check.arguments = ["-i", ":65534"]
        let pipe = Pipe()
        check.standardOutput = pipe
        try? check.run()
        check.waitUntilExit()
        
        let output = String(data: pipe.fileHandleForReading.readDataToEndOfFile(), encoding: .utf8) ?? ""
        if output.contains("LISTEN") {
            // Server đã đang chạy
            return
        }

        // Khởi động server.py bằng python3
        let serverScript = "\(baseDir)/server.py"
        guard FileManager.default.fileExists(atPath: serverScript) else { return }

        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: "/usr/bin/python3")
        proc.arguments = [serverScript]
        proc.currentDirectoryURL = URL(fileURLWithPath: baseDir)
        
        var env = ProcessInfo.processInfo.environment
        env["PYTHONUNBUFFERED"] = "1"
        proc.environment = env

        try? proc.run()
        self.serverProcess = proc
    }

    @objc func openDashboard() {
        if let url = URL(string: "http://localhost:65534/router") {
            NSWorkspace.shared.open(url)
        }
    }

    @objc func openScan() {
        if let url = URL(string: "http://localhost:65534/router") {
            NSWorkspace.shared.open(url)
        }
    }

    @objc func toggleAutostart() {
        let willEnable = !isAutostartEnabled
        guard let url = URL(string: "http://127.0.0.1:65534/api/autostart") else { return }
        
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let body = ["enable": willEnable]
        request.httpBody = try? JSONSerialization.data(withJSONObject: body)
        
        URLSession.shared.dataTask(with: request) { [weak self] data, response, _ in
            DispatchQueue.main.async {
                self?.updateStatus()
            }
        }.resume()
    }

    @objc func toggleServer() {
        if isServerRunning {
            // Dừng server
            let stopProc = Process()
            stopProc.executableURL = URL(fileURLWithPath: "/usr/bin/pkill")
            stopProc.arguments = ["-f", "server.py"]
            try? stopProc.run()
            stopProc.waitUntilExit()
            self.serverProcess = nil
            self.updateStatus()
        } else {
            // Khởi động server
            startServerInBackgroundIfNeeded()
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) { [weak self] in
                self?.updateStatus()
            }
        }
    }

    @objc func restartServer() {
        let stopProc = Process()
        stopProc.executableURL = URL(fileURLWithPath: "/usr/bin/pkill")
        stopProc.arguments = ["-f", "server.py"]
        try? stopProc.run()
        stopProc.waitUntilExit()
        self.serverProcess = nil

        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) { [weak self] in
            self?.startServerInBackgroundIfNeeded()
            self?.updateStatus()
        }
    }

    @objc func openHelp() {
        if let url = URL(string: "https://github.com/giangsamne/TailRouter") {
            NSWorkspace.shared.open(url)
        }
    }

    @objc func quitApp() {
        timer?.invalidate()
        if let proc = serverProcess, proc.isRunning {
            proc.terminate()
        }
        NSApplication.shared.terminate(nil)
    }
}

// Khởi chạy ứng dụng
let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate
app.run()
