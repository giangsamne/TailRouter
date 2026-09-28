package scanner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PortItem struct {
	Port          int    `json:"port"`
	Host          string `json:"host"`
	Protocol      string `json:"protocol"`
	Process       string `json:"process"`
	PID           int    `json:"pid"`
	Service       string `json:"service"`
	Description   string `json:"description"`
	Title         string `json:"title"`
	Container     string `json:"container,omitempty"`
	SuggestedPath string `json:"suggested_path"`
	SuggestedName string `json:"suggested_name"`
}

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func findBin(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	standardDirs := []string{"/usr/sbin", "/sbin", "/usr/bin", "/bin", "/usr/local/bin", "/opt/homebrew/bin"}
	for _, dir := range standardDirs {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return name
}

func runCmdWithStandardPath(name string, args ...string) ([]byte, error) {
	bin := findBin(name)
	cmd := exec.Command(bin, args...)
	pathEnv := "/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/opt/homebrew/bin"
	if curPath := os.Getenv("PATH"); curPath != "" {
		pathEnv = curPath + ":" + pathEnv
	}
	cmd.Env = append(os.Environ(), "PATH="+pathEnv)
	return cmd.CombinedOutput()
}

func (s *Scanner) Scan() ([]*PortItem, error) {
	portMap := make(map[int]*PortItem)

	// 1. Scan OS listening ports
	switch runtime.GOOS {
	case "linux":
		s.scanLinux(portMap)
	case "darwin":
		s.scanDarwin(portMap)
	case "windows":
		s.scanWindows(portMap)
	default:
		s.scanGeneric(portMap)
	}

	// 2. Scan Docker containers if docker CLI is present
	s.scanDocker(portMap)

	// 3. Remove port 65534 (the gateway itself)
	delete(portMap, 65534)

	// 4. Fallback: if no ports detected after OS scan, probe common ports
	if len(portMap) == 0 {
		s.scanGeneric(portMap)
		delete(portMap, 65534)
	}

	// 5. Concurrently inspect HTTP title and service signatures
	var wg sync.WaitGroup
	sem := make(chan struct{}, 15) // max 15 concurrent probes

	items := make([]*PortItem, 0, len(portMap))
	for _, item := range portMap {
		items = append(items, item)
	}

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(pi *PortItem) {
			defer wg.Done()
			defer func() { <-sem }()
			s.probeHTTP(pi)
		}(item)
	}
	wg.Wait()

	return items, nil
}

func (s *Scanner) scanLinux(portMap map[int]*PortItem) {
	out, err := runCmdWithStandardPath("ss", "-tulpn")
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Netid") || strings.HasPrefix(line, "State") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}
			localAddr := fields[4]
			colonIdx := strings.LastIndex(localAddr, ":")
			if colonIdx == -1 {
				continue
			}
			portStr := localAddr[colonIdx+1:]
			port, err := strconv.Atoi(portStr)
			if err != nil || port <= 0 || port > 65535 {
				continue
			}

			procName := ""
			pid := 0
			if len(fields) >= 7 {
				procField := fields[6]
				re := regexp.MustCompile(`users:\(\("([^"]+)",pid=(\d+)`)
				matches := re.FindStringSubmatch(procField)
				if len(matches) == 3 {
					procName = matches[1]
					pid, _ = strconv.Atoi(matches[2])
				}
			}

			portMap[port] = &PortItem{
				Port:     port,
				Host:     "127.0.0.1",
				Protocol: "tcp",
				Process:  procName,
				PID:      pid,
				Service:  detectKnownService(port, procName),
			}
		}
	}
}

func (s *Scanner) scanDarwin(portMap map[int]*PortItem) {
	out, _ := runCmdWithStandardPath("lsof", "-iTCP", "-sTCP:LISTEN", "-n", "-P")
	if len(out) > 0 {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "COMMAND") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 9 {
				continue
			}
			procName := fields[0]
			pid, _ := strconv.Atoi(fields[1])
			addr := fields[8]
			colonIdx := strings.LastIndex(addr, ":")
			if colonIdx == -1 {
				continue
			}
			port, err := strconv.Atoi(addr[colonIdx+1:])
			if err != nil || port <= 0 || port > 65535 {
				continue
			}

			portMap[port] = &PortItem{
				Port:     port,
				Host:     "127.0.0.1",
				Protocol: "tcp",
				Process:  procName,
				PID:      pid,
				Service:  detectKnownService(port, procName),
			}
		}
	}

	netOut, _ := exec.Command("/usr/sbin/netstat", "-an", "-p", "tcp").CombinedOutput()
	if len(netOut) == 0 {
		netOut, _ = runCmdWithStandardPath("netstat", "-an", "-p", "tcp")
	}
	if len(netOut) > 0 {
		scanner := bufio.NewScanner(strings.NewReader(string(netOut)))
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, "LISTEN") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}
			addr := fields[3]
			dotIdx := strings.LastIndex(addr, ".")
			if dotIdx == -1 {
				continue
			}
			port, err := strconv.Atoi(addr[dotIdx+1:])
			if err != nil || port <= 0 || port > 65535 {
				continue
			}
			if _, exists := portMap[port]; !exists {
				portMap[port] = &PortItem{
					Port:     port,
					Host:     "127.0.0.1",
					Protocol: "tcp",
					Service:  detectKnownService(port, ""),
				}
			}
		}
	}
}

func (s *Scanner) scanWindows(portMap map[int]*PortItem) {
	out, err := runCmdWithStandardPath("netstat", "-ano", "-p", "tcp")
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.Contains(line, "LISTENING") {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}
			addr := fields[1]
			colonIdx := strings.LastIndex(addr, ":")
			if colonIdx == -1 {
				continue
			}
			port, err := strconv.Atoi(addr[colonIdx+1:])
			if err != nil || port <= 0 || port > 65535 {
				continue
			}
			pid, _ := strconv.Atoi(fields[4])

			portMap[port] = &PortItem{
				Port:     port,
				Host:     "127.0.0.1",
				Protocol: "tcp",
				PID:      pid,
				Service:  detectKnownService(port, ""),
			}
		}
	}
}

func (s *Scanner) scanGeneric(portMap map[int]*PortItem) {
	commonPorts := []int{
		22, 80, 443, 3000, 3001, 4000, 5000, 5173, 8000, 8080, 8081,
		8085, 8088, 8090, 8123, 8443, 9000, 9090, 9443, 11434, 45068, 49152, 65312,
	}
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, p := range commonPorts {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			addr := fmt.Sprintf("127.0.0.1:%d", port)
			conn, err := net.DialTimeout("tcp", addr, 400*time.Millisecond)
			if err == nil {
				_ = conn.Close()
				mu.Lock()
				if _, exists := portMap[port]; !exists {
					portMap[port] = &PortItem{
						Port:     port,
						Host:     "127.0.0.1",
						Protocol: "tcp",
						Service:  detectKnownService(port, ""),
					}
				}
				mu.Unlock()
			}
		}(p)
	}
	wg.Wait()
}

func (s *Scanner) scanDocker(portMap map[int]*PortItem) {
	out, err := runCmdWithStandardPath("docker", "ps", "--format", "{{json .}}")
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	rePort := regexp.MustCompile(`(?:0\.0\.0\.0|:::?|127\.0\.0\.1):(\d+)->`)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var c struct {
			Names string `json:"Names"`
			Image string `json:"Image"`
			Ports string `json:"Ports"`
		}
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			continue
		}

		matches := rePort.FindAllStringSubmatch(c.Ports, -1)
		for _, m := range matches {
			if len(m) >= 2 {
				port, _ := strconv.Atoi(m[1])
				if port > 0 && port != 65534 {
					if existing, ok := portMap[port]; ok {
						existing.Container = c.Names
						existing.Description = fmt.Sprintf("Docker: %s (%s)", c.Names, c.Image)
						if existing.Service == "" || existing.Service == "TCP Port" {
							existing.Service = c.Names
						}
					} else {
						portMap[port] = &PortItem{
							Port:        port,
							Host:        "127.0.0.1",
							Protocol:    "tcp",
							Container:   c.Names,
							Description: fmt.Sprintf("Docker: %s (%s)", c.Names, c.Image),
							Service:     c.Names,
						}
					}
				}
			}
		}
	}
}

func (s *Scanner) probeHTTP(pi *PortItem) {
	suggestName := pi.Title
	if suggestName == "" {
		if pi.Container != "" {
			suggestName = pi.Container
		} else if pi.Service != "" && pi.Service != "TCP Port" {
			suggestName = pi.Service
		} else if pi.Process != "" {
			suggestName = pi.Process
		} else {
			suggestName = fmt.Sprintf("Port %d", pi.Port)
		}
	}
	pi.SuggestedName = suggestName
	pi.SuggestedPath = "/" + formatSlug(suggestName, pi.Port)

	client := &http.Client{
		Timeout: 500 * time.Millisecond,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 2 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/", pi.Port)
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "TailRouter-Scanner/2.0")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 16384))
	bodyStr := string(bodyBytes)

	reTitle := regexp.MustCompile(`(?i)<title[^>]*>(.*?)</title>`)
	matchTitle := reTitle.FindStringSubmatch(bodyStr)
	if len(matchTitle) >= 2 {
		cleanTitle := strings.TrimSpace(matchTitle[1])
		cleanTitle = strings.ReplaceAll(cleanTitle, "\n", " ")
		cleanTitle = strings.ReplaceAll(cleanTitle, "\r", " ")
		if cleanTitle != "" {
			pi.Title = cleanTitle
			pi.SuggestedName = cleanTitle
			pi.SuggestedPath = "/" + formatSlug(cleanTitle, pi.Port)
		}
	}

	serverHeader := resp.Header.Get("Server")
	if serverHeader != "" && pi.Description == "" {
		pi.Description = "Server: " + serverHeader
	}
}

func formatSlug(name string, port int) string {
	slug := strings.ToLower(name)
	regNonAlpha := regexp.MustCompile(`[^a-z0-9]+`)
	slug = regNonAlpha.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" || slug == "router" {
		slug = fmt.Sprintf("app-%d", port)
	}
	return slug
}

func detectKnownService(port int, proc string) string {
	lowProc := strings.ToLower(proc)
	if strings.Contains(lowProc, "bambu") {
		return "Bambu 3D Printer"
	}
	if strings.Contains(lowProc, "homeassistant") || strings.Contains(lowProc, "hass") {
		return "Home Assistant"
	}
	if strings.Contains(lowProc, "code-server") || strings.Contains(lowProc, "vscode") {
		return "VS Code Server"
	}
	if strings.Contains(lowProc, "ollama") {
		return "Ollama AI"
	}
	if strings.Contains(lowProc, "ssh") {
		return "OpenSSH"
	}

	switch port {
	case 22:
		return "SSH"
	case 80:
		return "HTTP Web"
	case 443:
		return "HTTPS Web"
	case 3000:
		return "Node.js / React"
	case 5000:
		return "Flask / FastAPI"
	case 8080:
		return "Bambu Timelapse / Web Proxy"
	case 8081:
		return "Web Service"
	case 8085:
		return "Cube 3D / App"
	case 8088:
		return "Tank Simulator / Web"
	case 8090:
		return "FPV Drone Simulator"
	case 8123:
		return "Home Assistant"
	case 9000:
		return "Portainer"
	case 11434:
		return "Ollama Local AI"
	default:
		if proc != "" {
			return proc
		}
		return "TCP Port"
	}
}
