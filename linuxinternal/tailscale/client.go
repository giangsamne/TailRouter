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
		return outputStr, fmt.Errorf("tailscale error: %v, output: %s", err, outputStr)
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

	return info, nil
}

func (c *Client) ConfigureServeRouter(port int) error {
	// Map /router to 127.0.0.1:port/router
	target := fmt.Sprintf("http://127.0.0.1:%d/router", port)
	_, err := c.RunCmd("serve", "--bg", "--https=443", "/router", target)
	return err
}

func (c *Client) ConfigureServeGateway(port int) error {
	// Map / to 127.0.0.1:port
	target := fmt.Sprintf("http://127.0.0.1:%d", port)
	_, err := c.RunCmd("serve", "--bg", "--https=443", "/", target)
	return err
}

func (c *Client) ResetServe() error {
	_, err := c.RunCmd("serve", "reset")
	return err
}

func (c *Client) ApplyRoute(path string, targetHost string, targetPort int, mode string) error {
	if mode == "funnel" {
		target := fmt.Sprintf("http://%s:%d", targetHost, targetPort)
		if _, err := c.RunCmd("serve", "--bg", "--https=443", path, target); err != nil {
			return err
		}
		_, err := c.RunCmd("funnel", "--bg", "--https=443", path, "on")
		return err
	}
	// mode serve
	target := fmt.Sprintf("http://%s:%d", targetHost, targetPort)
	_, err := c.RunCmd("serve", "--bg", "--https=443", path, target)
	return err
}

func (c *Client) RemoveRoute(path string) error {
	// Tailscale serve allows removing path by resetting or reconfiguring
	// tailscale serve --https=443 <path> off
	_, _ = c.RunCmd("funnel", "--https=443", path, "off")
	_, err := c.RunCmd("serve", "--https=443", path, "off")
	return err
}
