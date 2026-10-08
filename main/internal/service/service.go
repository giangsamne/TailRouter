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
	Supported   bool   `json:"supported"`
	Platform    string `json:"platform"`
	Enabled     bool   `json:"enabled"`
	Active      bool   `json:"active"`
	Method      string `json:"method"`
	ServiceType string `json:"service_type,omitempty"`
	ServiceFile string `json:"service_file,omitempty"`
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
		home, _ := os.UserHomeDir()
		// 1. Check systemd user service
		if _, err := exec.LookPath("systemctl"); err == nil {
			s.Method = "systemd"
			s.ServiceType = "systemd"
			unitFile := filepath.Join(home, ".config", "systemd", "user", "tailrouter.service")
			if _, err := os.Stat(unitFile); os.IsNotExist(err) {
				unitFile = filepath.Join(home, ".config", "systemd", "user", "tailscale-port-router.service")
			}
			if _, err := os.Stat(unitFile); err == nil {
				s.Enabled = true
				s.ServiceFile = unitFile
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
		}

		// 2. Check openrc service
		if !s.Enabled {
			if _, err := exec.LookPath("rc-service"); err == nil || fileExists("/sbin/rc-service") {
				if _, err := os.Stat("/etc/init.d/tailrouter"); err == nil {
					s.Enabled = true
					s.Active = true
					s.Method = "openrc"
					s.ServiceType = "openrc"
					s.ServiceFile = "/etc/init.d/tailrouter"
				}
			}
		}

		// 3. Check user crontab (@reboot)
		if !s.Enabled {
			cronOut, _ := exec.Command("crontab", "-l").CombinedOutput()
			cronStr := string(cronOut)
			if strings.Contains(cronStr, "tailrouter run") || strings.Contains(cronStr, "tailscale-port-router run") {
				s.Enabled = true
				s.Active = true
				s.Method = "crontab"
				s.ServiceType = "crontab"
				s.ServiceFile = "crontab (@reboot)"
			}
		}
		if !s.Enabled && os.Geteuid() == 0 {
			cTabs, _ := filepath.Glob("/var/spool/cron/crontabs/*")
			for _, ct := range cTabs {
				content, _ := os.ReadFile(ct)
				if strings.Contains(string(content), "tailrouter run") || strings.Contains(string(content), "tailscale-port-router run") {
					s.Enabled = true
					s.Active = true
					s.Method = "crontab"
					s.ServiceType = "crontab"
					s.ServiceFile = ct + " (@reboot)"
					break
				}
			}
		}

		// 4. Check ~/.profile, ~/.ashrc, ~/.bashrc
		if !s.Enabled {
			var checkProfiles []string
			if home != "" {
				for _, f := range []string{".profile", ".ashrc", ".bashrc"} {
					checkProfiles = append(checkProfiles, filepath.Join(home, f))
				}
			}
			if os.Geteuid() == 0 {
				for _, f := range []string{".profile", ".ashrc", ".bashrc"} {
					checkProfiles = append(checkProfiles, filepath.Join("/root", f))
				}
				userHomes, _ := filepath.Glob("/home/*")
				for _, uh := range userHomes {
					for _, f := range []string{".profile", ".ashrc", ".bashrc"} {
						checkProfiles = append(checkProfiles, filepath.Join(uh, f))
					}
				}
			}
			for _, p := range checkProfiles {
				content, err := os.ReadFile(p)
				if err == nil {
					cStr := string(content)
					if strings.Contains(cStr, "tailrouter run") || strings.Contains(cStr, "tailscale-port-router run") {
						s.Enabled = true
						s.Active = true
						s.Method = "profile"
						s.ServiceType = "profile"
						s.ServiceFile = p
						break
					}
				}
			}
		}

		if !s.Enabled {
			if _, err := exec.LookPath("systemctl"); err == nil {
				s.Method = "systemd"
				s.ServiceType = "systemd"
			} else {
				s.Method = "crontab"
				s.ServiceType = "crontab"
			}
		}

	case "darwin":
		s.Method = "launchd"
		s.ServiceType = "launchd"
		home, _ := os.UserHomeDir()
		plistFile := filepath.Join(home, "Library", "LaunchAgents", "com.tailrouter.gateway.plist")
		if _, err := os.Stat(plistFile); err == nil {
			s.Enabled = true
			s.ServiceFile = plistFile
		}
		s.Active = s.Enabled

	case "windows":
		s.Method = "registry"
		s.ServiceType = "registry"
		s.ServiceFile = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run\TailRouter`
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

func cleanFileLines(filePath string, keywords ...string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	var newLines []string
	hasChanged := false
	for _, l := range lines {
		matched := false
		for _, kw := range keywords {
			if strings.Contains(l, kw) {
				matched = true
				hasChanged = true
				break
			}
		}
		if !matched {
			newLines = append(newLines, l)
		}
	}
	if !hasChanged {
		return nil
	}
	return os.WriteFile(filePath, []byte(strings.Join(newLines, "\n")), 0644)
}

func updateCrontabFile(lines []string) error {
	cleaned := strings.TrimSpace(strings.Join(lines, "\n"))
	if cleaned == "" {
		return exec.Command("crontab", "-r").Run()
	}
	tmpFile, err := os.CreateTemp("", "tailrouter_cron_*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(cleaned + "\n"); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()
	return exec.Command("crontab", tmpFile.Name()).Run()
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
		// If systemctl is available, use systemd user service
		if _, err := exec.LookPath("systemctl"); err == nil {
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
		}

		// Fallback for non-systemd (Alpine Linux / OpenRC / crontab)
		if _, err := exec.LookPath("crontab"); err == nil {
			cronEntry := fmt.Sprintf("@reboot sleep 5 && %s run > %s 2>&1", execPath, filepath.Join(home, ".tailrouter.log"))
			out, _ := exec.Command("crontab", "-l").CombinedOutput()
			currCron := string(out)
			if !strings.Contains(currCron, "tailrouter run") && !strings.Contains(currCron, "tailscale-port-router run") {
				var lines []string
				if strings.TrimSpace(currCron) != "" {
					lines = strings.Split(strings.TrimSpace(currCron), "\n")
				}
				lines = append(lines, cronEntry)
				_ = updateCrontabFile(lines)
			}
			return nil
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
		home, _ := os.UserHomeDir()
		// 1. Stop & remove systemd user unit if exists
		if _, err := exec.LookPath("systemctl"); err == nil {
			_ = exec.Command("systemctl", "--user", "stop", "tailrouter").Run()
			_ = exec.Command("systemctl", "--user", "disable", "tailrouter").Run()
			if home != "" {
				_ = os.Remove(filepath.Join(home, ".config", "systemd", "user", "tailrouter.service"))
			}
			_ = exec.Command("systemctl", "--user", "daemon-reload").Run()
		}
		if os.Geteuid() == 0 {
			_ = exec.Command("systemctl", "stop", "tailrouter").Run()
			_ = exec.Command("systemctl", "disable", "tailrouter").Run()
			_ = os.Remove("/etc/systemd/system/tailrouter.service")
			_ = exec.Command("systemctl", "daemon-reload").Run()
		}

		// 2. OpenRC service
		if fileExists("/etc/init.d/tailrouter") {
			_ = exec.Command("rc-service", "tailrouter", "stop").Run()
			_ = exec.Command("rc-update", "del", "tailrouter", "default").Run()
			_ = os.Remove("/etc/init.d/tailrouter")
		}

		// 3. Remove from crontab using updateCrontabFile
		if _, err := exec.LookPath("crontab"); err == nil {
			out, err := exec.Command("crontab", "-l").CombinedOutput()
			if err == nil {
				var remaining []string
				for _, l := range strings.Split(string(out), "\n") {
					if !strings.Contains(l, "tailrouter run") && !strings.Contains(l, "tailscale-port-router run") {
						remaining = append(remaining, l)
					}
				}
				_ = updateCrontabFile(remaining)
			}
		}

		// If running as root, clean root crontab AND all crontabs in /var/spool/cron/crontabs/
		if os.Geteuid() == 0 {
			cTabs, _ := filepath.Glob("/var/spool/cron/crontabs/*")
			for _, ct := range cTabs {
				_ = cleanFileLines(ct, "tailrouter run", "tailscale-port-router run")
			}
			cTabs2, _ := filepath.Glob("/var/spool/cron/*")
			for _, ct := range cTabs2 {
				if !strings.HasSuffix(ct, "crontabs") {
					_ = cleanFileLines(ct, "tailrouter run", "tailscale-port-router run")
				}
			}
		}

		// 4. Remove from ~/.profile, ~/.ashrc, ~/.bashrc, ~/.bash_profile
		var profiles []string
		if home != "" {
			for _, f := range []string{".profile", ".ashrc", ".bashrc", ".bash_profile", ".zshrc"} {
				profiles = append(profiles, filepath.Join(home, f))
			}
		}
		if os.Geteuid() == 0 {
			for _, f := range []string{".profile", ".ashrc", ".bashrc", ".bash_profile"} {
				profiles = append(profiles, filepath.Join("/root", f))
			}
			userHomes, _ := filepath.Glob("/home/*")
			for _, uh := range userHomes {
				for _, f := range []string{".profile", ".ashrc", ".bashrc", ".bash_profile"} {
					profiles = append(profiles, filepath.Join(uh, f))
				}
			}
		}
		for _, p := range profiles {
			_ = cleanFileLines(p, "tailrouter run", "tailscale-port-router run")
		}
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

// UninstallBinary removes installed tailrouter executable files from standard installation directories
func (m *Manager) UninstallBinary() error {
	var candidates []string

	home, _ := os.UserHomeDir()
	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, ".local", "bin", "tailrouter"),
			filepath.Join(home, ".local", "bin", "tailrouter-arm64"),
			filepath.Join(home, ".local", "bin", "tailrouter-amd64"),
		)
	}
	candidates = append(candidates,
		"/usr/local/bin/tailrouter",
		"/usr/local/bin/tailrouter-arm64",
		"/usr/local/bin/tailrouter-amd64",
		"/usr/bin/tailrouter",
	)

	if execPath, err := os.Executable(); err == nil {
		execPath, _ = filepath.Abs(execPath)
		candidates = append(candidates, execPath)
	}

	for _, p := range candidates {
		if fileExists(p) {
			_ = os.Remove(p)
		}
	}
	return nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
