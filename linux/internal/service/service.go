package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Status struct {
	Supported bool   `json:"supported"`
	Platform  string `json:"platform"`
	Enabled   bool   `json:"enabled"`
	Active    bool   `json:"active"`
	Method    string `json:"method"`
}

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) GetStatus() *Status {
	s := &Status{
		Supported: true,
		Platform:  runtime.GOOS,
		Enabled:   false,
		Active:    false,
	}

	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("systemctl"); err == nil {
			s.Method = "systemd"
			home, _ := os.UserHomeDir()
			unitFile := filepath.Join(home, ".config", "systemd", "user", "tailrouter.service")
			if _, err := os.Stat(unitFile); os.IsNotExist(err) {
				unitFile = filepath.Join(home, ".config", "systemd", "user", "tailscale-port-router.service")
			}
			if _, err := os.Stat(unitFile); err == nil {
				s.Enabled = true
			}
			out, err := exec.Command("systemctl", "--user", "is-active", "tailrouter").CombinedOutput()
			if err == nil && strings.TrimSpace(string(out)) == "active" {
				s.Active = true
			} else {
				out2, err2 := exec.Command("systemctl", "--user", "is-active", "tailscale-port-router").CombinedOutput()
				if err2 == nil && strings.TrimSpace(string(out2)) == "active" {
					s.Active = true
				}
			}
		} else if _, err := exec.LookPath("rc-service"); err == nil || fileExists("/sbin/rc-service") {
			s.Method = "openrc"
			if _, err := os.Stat("/etc/init.d/tailrouter"); err == nil {
				s.Enabled = true
			}
			out, err := exec.Command("rc-service", "tailrouter", "status").CombinedOutput()
			if err == nil && strings.Contains(string(out), "started") {
				s.Active = true
			}
		} else {
			s.Method = "unsupported"
			s.Supported = false
		}
	case "darwin":
		s.Method = "launchd"
		home, _ := os.UserHomeDir()
		plistFile := filepath.Join(home, "Library", "LaunchAgents", "com.tailrouter.gateway.plist")
		if _, err := os.Stat(plistFile); err == nil {
			s.Enabled = true
		}
		s.Active = s.Enabled
	case "windows":
		s.Method = "registry"
		out, err := exec.Command("reg", "query", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "TailRouter").CombinedOutput()
		if err == nil && strings.Contains(string(out), "TailRouter") {
			s.Enabled = true
			s.Active = true
		}
	default:
		s.Supported = false
	}

	return s
}

func (m *Manager) Enable() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, _ = filepath.Abs(execPath)
	execDir := filepath.Dir(execPath)

	switch runtime.GOOS {
	case "linux":
		home, _ := os.UserHomeDir()
		systemdDir := filepath.Join(home, ".config", "systemd", "user")
		_ = os.MkdirAll(systemdDir, 0755)

		unitContent := fmt.Sprintf(`[Unit]
Description=TailRouter High-Performance Native Gateway
After=network.target tailscaled.service

[Service]
Type=simple
WorkingDirectory=%s
ExecStart=%s run
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=default.target
`, execDir, execPath)

		unitPath := filepath.Join(systemdDir, "tailrouter.service")
		if err := os.WriteFile(unitPath, []byte(unitContent), 0644); err != nil {
			return err
		}

		_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
		_ = exec.Command("systemctl", "--user", "enable", "tailrouter").Run()
		_ = exec.Command("systemctl", "--user", "start", "tailrouter").Run()

		user := os.Getenv("USER")
		if user != "" {
			_ = exec.Command("loginctl", "enable-linger", user).Run()
		}
		return nil

	case "darwin":
		home, _ := os.UserHomeDir()
		agentDir := filepath.Join(home, "Library", "LaunchAgents")
		_ = os.MkdirAll(agentDir, 0755)

		plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.tailrouter.gateway</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>run</string>
    </array>
    <key>WorkingDirectory</key>
    <string>%s</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>/tmp/tailrouter.err.log</string>
    <key>StandardOutPath</key>
    <string>/tmp/tailrouter.out.log</string>
</dict>
</plist>
`, execPath, execDir)

		plistPath := filepath.Join(agentDir, "com.tailrouter.gateway.plist")
		if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
			return err
		}
		_ = exec.Command("launchctl", "load", "-w", plistPath).Run()
		return nil

	case "windows":
		cmdArg := fmt.Sprintf(`"%s" run`, execPath)
		return exec.Command("reg", "add", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "TailRouter", "/t", "REG_SZ", "/d", cmdArg, "/f").Run()
	}

	return fmt.Errorf("hệ điều hành %s chưa được hỗ trợ tự khởi động", runtime.GOOS)
}

func (m *Manager) Disable() error {
	switch runtime.GOOS {
	case "linux":
		_ = exec.Command("systemctl", "--user", "stop", "tailrouter").Run()
		_ = exec.Command("systemctl", "--user", "disable", "tailrouter").Run()
		home, _ := os.UserHomeDir()
		unitPath := filepath.Join(home, ".config", "systemd", "user", "tailrouter.service")
		_ = os.Remove(unitPath)
		_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
		return nil

	case "darwin":
		home, _ := os.UserHomeDir()
		plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.tailrouter.gateway.plist")
		// Safe disable: remove plist file so it doesn't boot next time, but avoid killing current running process
		_ = os.Remove(plistPath)
		return nil

	case "windows":
		return exec.Command("reg", "delete", `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, "/v", "TailRouter", "/f").Run()
	}
	return nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
