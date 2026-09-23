package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var ReservedPaths = map[string]bool{
	"/":            true,
	"/router":      true,
	"/api":         true,
	"/favicon.ico": true,
	"/favicon.svg": true,
}

type Route struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Path          string  `json:"path"`
	TargetHost    string  `json:"target_host"`
	TargetPort    int     `json:"target_port"`
	StripPrefix   bool    `json:"strip_prefix"`
	Enabled       bool    `json:"enabled"`
	CreatedAt     string  `json:"created_at"`
	Hits          int64   `json:"hits"`
	LastStatus    string  `json:"last_status"`
	LastLatencyMs float64 `json:"last_latency_ms"`
	Notes         string  `json:"notes"`
	Mode          string  `json:"mode"` // "serve" or "funnel"
}

type Manager struct {
	mu         sync.RWMutex
	configPath string
	routes     map[string]*Route
}

func NewManager(configPath string) *Manager {
	if configPath == "" {
		execDir, err := os.Executable()
		if err != nil {
			execDir = "."
		} else {
			execDir = filepath.Dir(execDir)
		}
		configPath = filepath.Join(execDir, "config", "routes.json")
	}

	m := &Manager{
		configPath: configPath,
		routes:     make(map[string]*Route),
	}
	_ = m.Load()
	return m
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dir := filepath.Dir(m.configPath)
	_ = os.MkdirAll(dir, 0755)

	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		m.routes = make(map[string]*Route)
		return m.saveLocked()
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}

	var rawList []*Route
	if err := json.Unmarshal(data, &rawList); err != nil {
		// try unmarshaling as map
		var rawMap map[string]*Route
		if err2 := json.Unmarshal(data, &rawMap); err2 == nil {
			for _, r := range rawMap {
				rawList = append(rawList, r)
			}
		} else {
			m.routes = make(map[string]*Route)
			return err
		}
	}

	m.routes = make(map[string]*Route)
	for _, r := range rawList {
		if r == nil || r.ID == "" {
			continue
		}
		// Strict filter: Never allow routing to port 65534 or reserved paths
		if r.TargetPort == 65534 || ReservedPaths[r.Path] {
			continue
		}
		m.routes[r.ID] = r
	}

	return nil
}

func (m *Manager) saveLocked() error {
	dir := filepath.Dir(m.configPath)
	_ = os.MkdirAll(dir, 0755)

	list := make([]*Route, 0, len(m.routes))
	for _, r := range m.routes {
		list = append(list, r)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt < list[j].CreatedAt
	})

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.configPath, data, 0644)
}

func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveLocked()
}

func (m *Manager) List() []*Route {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*Route, 0, len(m.routes))
	for _, r := range m.routes {
		// clone
		cp := *r
		list = append(list, &cp)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt < list[j].CreatedAt
	})
	return list
}

func (m *Manager) Get(id string) (*Route, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.routes[id]
	if !ok {
		return nil, false
	}
	cp := *r
	return &cp, true
}

func ValidatePath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = strings.TrimRight(p, "/")
	if p == "" {
		p = "/"
	}

	if ReservedPaths[p] {
		return "", fmt.Errorf("đường dẫn '%s' là đường dẫn hệ thống được bảo lưu", p)
	}

	validPattern := regexp.MustCompile(`^/[a-zA-Z0-9_\-\.\/]+$`)
	if !validPattern.MatchString(p) {
		return "", errors.New("đường dẫn chứa ký tự không hợp lệ")
	}
	return p, nil
}

func (m *Manager) Add(name, path, targetHost string, targetPort int, stripPrefix bool, mode, notes string) (*Route, error) {
	normPath, err := ValidatePath(path)
	if err != nil {
		return nil, err
	}

	if targetPort <= 0 || targetPort > 65535 {
		return nil, errors.New("cổng không hợp lệ (1-65535)")
	}
	if targetPort == 65534 {
		return nil, errors.New("không thể đăng ký chính cổng Gateway (65534) làm cổng đích")
	}

	if targetHost == "" {
		targetHost = "127.0.0.1"
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, r := range m.routes {
		if r.Path == normPath {
			return nil, fmt.Errorf("đường dẫn '%s' đã được sử dụng bởi route '%s'", normPath, r.Name)
		}
	}

	if mode != "funnel" {
		mode = "serve"
	}

	now := time.Now().UTC().Format(time.RFC3339)
	routeID := fmt.Sprintf("route_%d_%d", time.Now().Unix(), targetPort)
	if name == "" {
		name = fmt.Sprintf("Port %d", targetPort)
	}

	r := &Route{
		ID:            routeID,
		Name:          name,
		Path:          normPath,
		TargetHost:    targetHost,
		TargetPort:    targetPort,
		StripPrefix:   stripPrefix,
		Enabled:       true,
		CreatedAt:     now,
		Hits:          0,
		LastStatus:    "unknown",
		LastLatencyMs: 0,
		Notes:         notes,
		Mode:          mode,
	}

	m.routes[routeID] = r
	_ = m.saveLocked()
	return r, nil
}

func (m *Manager) Update(id string, updates map[string]interface{}) (*Route, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.routes[id]
	if !ok {
		return nil, errors.New("route không tồn tại")
	}

	if name, ok := updates["name"].(string); ok && name != "" {
		r.Name = name
	}
	if p, ok := updates["path"].(string); ok {
		normPath, err := ValidatePath(p)
		if err != nil {
			return nil, err
		}
		for rID, existing := range m.routes {
			if rID != id && existing.Path == normPath {
				return nil, fmt.Errorf("đường dẫn '%s' đã được sử dụng", normPath)
			}
		}
		r.Path = normPath
	}
	if port, ok := updates["target_port"].(float64); ok {
		iPort := int(port)
		if iPort <= 0 || iPort > 65535 || iPort == 65534 {
			return nil, errors.New("cổng không hợp lệ")
		}
		r.TargetPort = iPort
	}
	if host, ok := updates["target_host"].(string); ok && host != "" {
		r.TargetHost = host
	}
	if strip, ok := updates["strip_prefix"].(bool); ok {
		r.StripPrefix = strip
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		r.Enabled = enabled
	}
	if notes, ok := updates["notes"].(string); ok {
		r.Notes = notes
	}
	if mode, ok := updates["mode"].(string); ok {
		if mode == "funnel" || mode == "serve" {
			r.Mode = mode
		}
	}

	_ = m.saveLocked()
	return r, nil
}

func (m *Manager) Delete(id string) (*Route, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.routes[id]
	if !ok {
		return nil, errors.New("route không tồn tại")
	}
	delete(m.routes, id)
	_ = m.saveLocked()
	return r, nil
}

func (m *Manager) Toggle(id string) (*Route, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.routes[id]
	if !ok {
		return nil, errors.New("route không tồn tại")
	}
	r.Enabled = !r.Enabled
	_ = m.saveLocked()
	return r, nil
}

func (m *Manager) ToggleMode(id string) (*Route, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, ok := m.routes[id]
	if !ok {
		return nil, errors.New("route không tồn tại")
	}
	if r.Mode == "serve" {
		r.Mode = "funnel"
	} else {
		r.Mode = "serve"
	}
	_ = m.saveLocked()
	return r, nil
}

// FindMatchingRoute matches the longest matching path
func (m *Manager) FindMatchingRoute(reqPath string) (*Route, string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var bestRoute *Route
	var bestMatch string

	for _, r := range m.routes {
		if !r.Enabled {
			continue
		}
		p := r.Path
		if reqPath == p || strings.HasPrefix(reqPath, p+"/") {
			if len(p) > len(bestMatch) {
				bestMatch = p
				bestRoute = r
			}
		}
	}

	if bestRoute == nil {
		return nil, "", false
	}

	remainder := strings.TrimPrefix(reqPath, bestMatch)
	if !strings.HasPrefix(remainder, "/") {
		remainder = "/" + remainder
	}
	cp := *bestRoute
	return &cp, remainder, true
}

func (m *Manager) IncrementHits(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.routes[id]; ok {
		r.Hits++
	}
}

func (m *Manager) UpdateStatus(id string, status string, latencyMs float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.routes[id]; ok {
		r.LastStatus = status
		r.LastLatencyMs = latencyMs
	}
}

func (m *Manager) PingRoute(id string) (string, float64, error) {
	m.mu.RLock()
	r, ok := m.routes[id]
	m.mu.RUnlock()

	if !ok {
		return "unknown", 0, errors.New("route không tồn tại")
	}

	start := time.Now()
	addr := fmt.Sprintf("%s:%d", r.TargetHost, r.TargetPort)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	latency := float64(time.Since(start).Microseconds()) / 1000.0

	status := "online"
	if err != nil {
		status = "offline"
	} else {
		_ = conn.Close()
	}

	m.UpdateStatus(id, status, latency)
	_ = m.Save()
	return status, latency, nil
}
