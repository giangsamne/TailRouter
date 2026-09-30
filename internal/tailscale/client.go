package tailscale

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type StatusInfo struct {
	Installed       bool     `json:"installed"`
	Running         bool     `json:"running"`
	BackendState    string   `json:"backend_state"`
	NodeName        string   `json:"node_name"`
	FQDN            string   `json:"fqdn"`
	IPs             []string `json:"ips"`
	MagicDNSEnabled bool     `json:"magic_dns"`
	ServeConfigured bool     `json:"serve_configured"`
	RemoteRouterURL string   `json:"remote_router_url,omitempty"`
	NeedsOperator   bool     `json:"needs_operator"`
	RawJSON         string   `json:"-"`
}

type Client struct {
	binaryPath string
}

func NewClient() *Client {
	c := &Client{}
	c.binaryPath = c.findBinary()
	return c
}

func (c *Client) findBinary() string {
	// 1. Check PATH
	if p, err := exec.LookPath("tailscale"); err == nil {
		return p
	}

	// 2. OS specific known paths
	switch runtime.GOOS {
	case "darwin":
		paths := []string{
			"/Applications/Tailscale.app/Contents/MacOS/Tailscale",
			"/usr/local/bin/tailscale",
			"/opt/homebrew/bin/tailscale",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "windows":
		paths := []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Tailscale", "tailscale.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Tailscale", "tailscale.exe"),
			filepath.Join(os.Getenv("LocalAppData"), "Tailscale", "tailscale.exe"),
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	case "linux":
		paths := []string{
			"/usr/bin/tailscale",
			"/usr/local/bin/tailscale",
			"/bin/tailscale",
		}
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

func (c *Client) RunCmd(args ...string) (string, error) {
	bin := c.binaryPath
	if bin == "" {
		bin = c.findBinary()
		c.binaryPath = bin
	}
	if bin == "" {
		return "", errors.New("không tìm thấy lệnh tailscale trên hệ thống")
	}

	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "SHLVL=1")
	out, err := cmd.CombinedOutput()
	outputStr := string(out)
	if err != nil {
		if strings.Contains(outputStr, "Access denied") || strings.Contains(outputStr, "operator") {
			if sudoPath, sErr := exec.LookPath("sudo"); sErr == nil {
				sudoArgs := append([]string{"-n", bin}, args...)
				sCmd := exec.Command(sudoPath, sudoArgs...)
				sCmd.Env = append(os.Environ(), "SHLVL=1")
				if sOut, sErr2 := sCmd.CombinedOutput(); sErr2 == nil {
					return string(sOut), nil
				}
			}
		}
		cleanOut := strings.TrimSpace(outputStr)
		if cleanOut != "" {
			return outputStr, errors.New(cleanOut)
		}
		return outputStr, fmt.Errorf("tailscale error: %v", err)
	}
	return outputStr, nil
}

func (c *Client) GetStatus() (*StatusInfo, error) {
	info := &StatusInfo{
		Installed:    false,
		Running:      false,
		BackendState: "NoBinary",
		IPs:          []string{},
	}

	if c.binaryPath == "" {
		c.binaryPath = c.findBinary()
	}
	if c.binaryPath == "" {
		return info, nil
	}
	info.Installed = true

	out, err := c.RunCmd("status", "--json")
	if err != nil {
		info.BackendState = "Stopped"
		return info, nil
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(out), &data); err != nil {
		return info, err
	}

	info.Running = true
	if state, ok := data["BackendState"].(string); ok {
		info.BackendState = state
	}

	if self, ok := data["Self"].(map[string]interface{}); ok {
		if hostName, ok := self["HostName"].(string); ok {
			info.NodeName = hostName
		}
		if dnsName, ok := self["DNSName"].(string); ok {
			info.FQDN = strings.TrimRight(dnsName, ".")
		}
		if ips, ok := self["TailscaleIPs"].([]interface{}); ok {
			for _, ip := range ips {
				if s, ok := ip.(string); ok {
					info.IPs = append(info.IPs, s)
				}
			}
		}
	}

	if magic, ok := data["MagicDNSEnabled"].(bool); ok {
		info.MagicDNSEnabled = magic
	}

	// Detect Tailscale Serve status
	serveOut, serveErr := c.RunCmd("serve", "status", "--json")
	if serveErr == nil && serveOut != "" {
		var serveData map[string]interface{}
		if err := json.Unmarshal([]byte(serveOut), &serveData); err == nil {
			if web, ok := serveData["Web"].(map[string]interface{}); ok && len(web) > 0 {
				for hostKey, v := range web {
					if info.FQDN == "" {
						hostPort := strings.Split(hostKey, ":")
						if len(hostPort) > 0 && strings.Contains(hostPort[0], ".") {
							info.FQDN = hostPort[0]
						}
					}
					if hostConfig, ok := v.(map[string]interface{}); ok {
						if handlers, ok := hostConfig["Handlers"].(map[string]interface{}); ok {
							if _, hasRouter := handlers["/router"]; hasRouter {
								info.ServeConfigured = true
							}
							if _, hasRoot := handlers["/"]; hasRoot {
								info.ServeConfigured = true
							}
						}
					}
				}
			}
		}
	} else if serveErr != nil {
		errStr := serveErr.Error() + " " + serveOut
		if strings.Contains(errStr, "Access denied") || strings.Contains(errStr, "operator") {
			info.NeedsOperator = true
		}
	}

	if !info.ServeConfigured {
		serveText, tErr := c.RunCmd("serve", "status")
		if tErr != nil {
			errStr := tErr.Error() + " " + serveText
			if strings.Contains(errStr, "Access denied") || strings.Contains(errStr, "operator") {
				info.NeedsOperator = true
			}
		} else {
			if strings.Contains(serveText, "/router") || strings.Contains(serveText, "proxy http://") {
				info.ServeConfigured = true
			}
		}
	}

	if info.ServeConfigured {
		if info.FQDN != "" {
			info.RemoteRouterURL = fmt.Sprintf("https://%s", info.FQDN)
		} else if len(info.IPs) > 0 {
			info.RemoteRouterURL = fmt.Sprintf("https://%s", info.IPs[0])
		}
	}

	return info, nil
}

func (c *Client) ConfigureServeRouter(port int) error {
	return c.ConfigureServeGateway(port)
}

func (c *Client) ConfigureServeGateway(port int) error {
	// Map root to 127.0.0.1:port
	target := fmt.Sprintf("http://127.0.0.1:%d", port)
	_, err := c.RunCmd("serve", "--bg", "--yes", "--https=443", target)
	return err
}

func (c *Client) ResetServe() error {
	_, err := c.RunCmd("serve", "reset")
	return err
}

func (c *Client) ApplyRoute(path string, targetHost string, targetPort int, mode string) error {
	cleanPath := "/" + strings.Trim(strings.TrimSpace(path), "/")
	if cleanPath == "" {
		cleanPath = "/"
	}
	target := fmt.Sprintf("http://%s:%d", targetHost, targetPort)
	if mode == "funnel" {
		_, err := c.RunCmd("funnel", "--bg", "--yes", "--https=443", fmt.Sprintf("--set-path=%s", cleanPath), target)
		return err
	}
	// mode serve
	_, err := c.RunCmd("serve", "--bg", "--yes", "--https=443", fmt.Sprintf("--set-path=%s", cleanPath), target)
	return err
}

func (c *Client) RemoveRoute(path string) error {
	cleanPath := "/" + strings.Trim(strings.TrimSpace(path), "/")
	if cleanPath == "" {
		cleanPath = "/"
	}
	_, _ = c.RunCmd("funnel", "--https=443", fmt.Sprintf("--set-path=%s", cleanPath), "off")
	_, err := c.RunCmd("serve", "--https=443", fmt.Sprintf("--set-path=%s", cleanPath), "off")
	return err
}
