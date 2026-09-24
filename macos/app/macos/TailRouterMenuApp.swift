import Cocoa

class AppDelegate: NSObject, NSApplicationDelegate {
    var statusItem: NSStatusItem!
    var serverProcess: Process?
    var timer: Timer?
    var isServerRunning = false
    var isAutostartEnabled = false

    let statusMenuItem = NSMenuItem(title: "🟡 Đang kiểm tra Gateway...", action: nil, keyEquivalent: "")
    let ipMenuItem = NSMenuItem(title: "🌐 IP: Đang tải...", action: nil, keyEquivalent: "")
    let routesMenuItem = NSMenuItem(title: "⚡ Tuyến đường: Đang tải...", action: nil, keyEquivalent: "")
    let autostartMenuItem = NSMenuItem(title: "🚀 Tự khởi động cùng Mac", action: #selector(toggleAutostart), keyEquivalent: "")
    let toggleServerMenuItem = NSMenuItem(title: "▶️ Khởi Động Gateway", action: #selector(toggleServer), keyEquivalent: "s")

    func applicationDidFinishLaunching(_ aNotification: Notification) {
        statusItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
        
        if let button = statusItem.button {
            if let img = NSImage(systemSymbolName: "arrow.triangle.swap", accessibilityDescription: "TailRouter") {
                img.isTemplate = true
                button.image = img
            } else {
                button.title = "TR"
            }
        }

        buildMenu()
        startServerInBackgroundIfNeeded()

        updateStatus()
        timer = Timer.scheduledTimer(withTimeInterval: 3.0, repeats: true) { [weak self] _ in
            self?.updateStatus()
        }
    }

    func buildMenu() {
        let menu = NSMenu()
        
        let titleItem = NSMenuItem(title: "TailRouter Native Engine v2.0", action: nil, keyEquivalent: "")
        titleItem.isEnabled = false
        menu.addItem(titleItem)
        
        menu.addItem(NSMenuItem.separator())
        menu.addItem(statusMenuItem)
        menu.addItem(ipMenuItem)
        menu.addItem(routesMenuItem)
        
        menu.addItem(NSMenuItem.separator())
        
        let openDashboardItem = NSMenuItem(title: "🌐 Mở Bảng Quản Trị (/router)", action: #selector(openDashboard), keyEquivalent: "o")
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
                        let routesCount = json["routes_count"] as? Int ?? 0
                        self.routesMenuItem.title = "⚡ Tuyến đường: \(routesCount) đang hoạt động"
                        
                        if let ts = json["tailscale"] as? [String: Any] {
                            let ips = ts["ips"] as? [String] ?? []
                            let ip = ips.first ?? "N/A"
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
        let check = Process()
        check.executableURL = URL(fileURLWithPath: "/usr/sbin/lsof")
        check.arguments = ["-i", ":65534"]
        let pipe = Pipe()
        check.standardOutput = pipe
        try? check.run()
        check.waitUntilExit()
        
        let output = String(data: pipe.fileHandleForReading.readDataToEndOfFile(), encoding: .utf8) ?? ""
        if output.contains("LISTEN") {
            return
        }

        let bundleURL = Bundle.main.bundleURL
        var binPath = bundleURL.appendingPathComponent("Contents/MacOS/tailrouter-server").path
        if !FileManager.default.fileExists(atPath: binPath) {
            binPath = bundleURL.appendingPathComponent("Contents/Resources/tailrouter-server").path
        }
        guard FileManager.default.fileExists(atPath: binPath) else { return }

        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: binPath)
        proc.arguments = ["run"]
        proc.currentDirectoryURL = bundleURL.appendingPathComponent("Contents/Resources")

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
        let body = ["enabled": willEnable]
        request.httpBody = try? JSONSerialization.data(withJSONObject: body)
        
        URLSession.shared.dataTask(with: request) { [weak self] data, response, _ in
            DispatchQueue.main.async {
                self?.updateStatus()
            }
        }.resume()
    }

    @objc func toggleServer() {
        if isServerRunning {
            stopServer()
            self.updateStatus()
        } else {
            startServerInBackgroundIfNeeded()
            DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) { [weak self] in
                self?.updateStatus()
            }
        }
    }

    @objc func restartServer() {
        stopServer()
        DispatchQueue.main.asyncAfter(deadline: .now() + 1.0) { [weak self] in
            self?.startServerInBackgroundIfNeeded()
            self?.updateStatus()
        }
    }

    func stopServer() {
        if let proc = serverProcess, proc.isRunning {
            proc.terminate()
        }
        let stopProc = Process()
        stopProc.executableURL = URL(fileURLWithPath: "/usr/bin/pkill")
        stopProc.arguments = ["-f", "tailrouter-server"]
        try? stopProc.run()
        stopProc.waitUntilExit()
        self.serverProcess = nil
    }

    @objc func openHelp() {
        if let url = URL(string: "https://github.com/giangsamne/TailRouter") {
            NSWorkspace.shared.open(url)
        }
    }

    @objc func quitApp() {
        timer?.invalidate()
        stopServer()
        NSApplication.shared.terminate(nil)
    }
}

let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate
app.run()
