package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	SessionDuration     = 10 * time.Minute
	TempPasswordDuration = 10 * time.Minute
	TokenByteLen        = 32
)

type AuthData struct {
	Enabled       bool      `json:"enabled"`
	PasswordHash  string    `json:"password_hash"`
	PasswordSalt  string    `json:"password_salt"`
	TempPassword  string    `json:"temp_password,omitempty"`
	TempExpiresAt time.Time `json:"temp_expires_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SessionInfo struct {
	Token            string    `json:"token"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	RemainingSeconds int       `json:"remaining_seconds"`
}

type StatusInfo struct {
	Enabled             bool   `json:"enabled"`
	Authenticated       bool   `json:"authenticated"`
	HasMasterPassword   bool   `json:"has_master_password"`
	HasTempPassword     bool   `json:"has_temp_password"`
	TempExpiresIn       int    `json:"temp_expires_in"` // seconds
	TempPasswordMasked  string `json:"temp_password_masked,omitempty"`
	SessionRemainingSec int    `json:"session_remaining_sec"`
}

type Manager struct {
	mu         sync.RWMutex
	configPath string
	lastLoaded time.Time
	data       AuthData
	sessions   map[string]time.Time // token -> expiresAt
}

func DefaultAuthPath(routesConfigPath string) string {
	dir := filepath.Dir(routesConfigPath)
	return filepath.Join(dir, "auth.json")
}

func NewManager(authPath string) *Manager {
	if authPath == "" {
		home, _ := os.UserHomeDir()
		authPath = filepath.Join(home, ".config", "tailrouter", "auth.json")
	}

	m := &Manager{
		configPath: authPath,
		sessions:   make(map[string]time.Time),
	}
	_ = m.loadLocked()
	return m
}

func (m *Manager) loadLocked() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.data = AuthData{Enabled: false}
			return nil
		}
		return err
	}
	var loaded AuthData
	if err := json.Unmarshal(data, &loaded); err != nil {
		return err
	}
	m.data = loaded
	if fi, err := os.Stat(m.configPath); err == nil {
		m.lastLoaded = fi.ModTime()
	}
	return nil
}

func (m *Manager) checkReloadLocked() {
	fi, err := os.Stat(m.configPath)
	if err != nil {
		if os.IsNotExist(err) && m.data.Enabled {
			m.data = AuthData{Enabled: false}
		}
		return
	}
	if fi.ModTime().After(m.lastLoaded) {
		_ = m.loadLocked()
	}
}

func (m *Manager) saveLocked() error {
	dir := filepath.Dir(m.configPath)
	_ = os.MkdirAll(dir, 0700)

	m.data.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(m.data, "", "  ")
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, ".auth-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}
	_ = tmpFile.Chmod(0600)
	_ = tmpFile.Sync()
	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, m.configPath); err != nil {
		return err
	}
	_ = os.Chmod(m.configPath, 0600)
	if fi, err := os.Stat(m.configPath); err == nil {
		m.lastLoaded = fi.ModTime()
	}
	return nil
}

func hashPassword(plain, saltHex string) string {
	salt, _ := hex.DecodeString(saltHex)
	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(plain))
	return hex.EncodeToString(h.Sum(nil))
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func generateTempOTP() string {
	// Generates e.g. TR-839215
	maxVal := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return fmt.Sprintf("TR-%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("TR-%06d", n.Int64()+100000)
}

func (m *Manager) IsEnabled() bool {
	m.mu.Lock()
	m.checkReloadLocked()
	defer m.mu.Unlock()
	return m.data.Enabled
}

func (m *Manager) SetPassword(plain string) error {
	plain = strings.TrimSpace(plain)
	if len(plain) < 4 {
		return errors.New("mật khẩu phải có ít nhất 4 ký tự")
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkReloadLocked()

	salt := generateRandomHex(16)
	hash := hashPassword(plain, salt)

	m.data.Enabled = true
	m.data.PasswordSalt = salt
	m.data.PasswordHash = hash
	return m.saveLocked()
}

func (m *Manager) ClearPassword() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkReloadLocked()

	m.data.Enabled = false
	m.data.PasswordHash = ""
	m.data.PasswordSalt = ""
	m.data.TempPassword = ""
	m.data.TempExpiresAt = time.Time{}
	m.sessions = make(map[string]time.Time)
	return m.saveLocked()
}

func (m *Manager) GenerateTempPassword() (string, time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkReloadLocked()

	otp := generateTempOTP()
	expires := time.Now().Add(TempPasswordDuration)

	m.data.Enabled = true
	m.data.TempPassword = otp
	m.data.TempExpiresAt = expires

	if err := m.saveLocked(); err != nil {
		return "", time.Time{}, err
	}
	return otp, expires, nil
}

func (m *Manager) VerifyPassword(input string) bool {
	m.mu.Lock()
	m.checkReloadLocked()
	defer m.mu.Unlock()

	if !m.data.Enabled {
		return true
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return false
	}

	// 1. Check Master Password
	if m.data.PasswordHash != "" && m.data.PasswordSalt != "" {
		expectedHash := m.data.PasswordHash
		actualHash := hashPassword(input, m.data.PasswordSalt)
		if subtle.ConstantTimeCompare([]byte(expectedHash), []byte(actualHash)) == 1 {
			return true
		}
	}

	// 2. Check 10-minute Temporary OTP
	if m.data.TempPassword != "" && time.Now().Before(m.data.TempExpiresAt) {
		if subtle.ConstantTimeCompare([]byte(m.data.TempPassword), []byte(input)) == 1 {
			return true
		}
	}

	return false
}

func (m *Manager) CreateSession() SessionInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	token := generateRandomHex(TokenByteLen)
	expiresAt := time.Now().Add(SessionDuration)
	m.sessions[token] = expiresAt

	return SessionInfo{
		Token:            token,
		CreatedAt:        time.Now(),
		ExpiresAt:        expiresAt,
		RemainingSeconds: int(SessionDuration.Seconds()),
	}
}

func (m *Manager) ValidateSession(token string) (bool, int) {
	if token == "" {
		return false, 0
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt, ok := m.sessions[token]
	if !ok {
		return false, 0
	}

	now := time.Now()
	if now.After(expiresAt) {
		delete(m.sessions, token)
		return false, 0
	}

	remaining := int(expiresAt.Sub(now).Seconds())
	if remaining < 0 {
		remaining = 0
	}
	return true, remaining
}

func (m *Manager) ExtendSession(token string) (SessionInfo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	expiresAt, ok := m.sessions[token]
	if !ok || time.Now().After(expiresAt) {
		delete(m.sessions, token)
		return SessionInfo{}, errors.New("phiên đăng nhập đã hết hạn hoặc không tồn tại")
	}

	newExpiry := time.Now().Add(SessionDuration)
	m.sessions[token] = newExpiry

	return SessionInfo{
		Token:            token,
		CreatedAt:        time.Now(),
		ExpiresAt:        newExpiry,
		RemainingSeconds: int(SessionDuration.Seconds()),
	}, nil
}

func (m *Manager) RevokeSession(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
}

func (m *Manager) RevokeAllSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = make(map[string]time.Time)
}

func (m *Manager) GetStatus(token string) StatusInfo {
	m.mu.Lock()
	m.checkReloadLocked()
	defer m.mu.Unlock()

	now := time.Now()
	hasTemp := m.data.TempPassword != "" && now.Before(m.data.TempExpiresAt)
	tempRemaining := 0
	tempMasked := ""
	if hasTemp {
		tempRemaining = int(m.data.TempExpiresAt.Sub(now).Seconds())
		if len(m.data.TempPassword) > 4 {
			tempMasked = m.data.TempPassword[:3] + "***"
		} else {
			tempMasked = "***"
		}
	}

	authenticated := false
	sessionRemaining := 0
	if !m.data.Enabled {
		authenticated = true
		sessionRemaining = 600
	} else if token != "" {
		if exp, ok := m.sessions[token]; ok && now.Before(exp) {
			authenticated = true
			sessionRemaining = int(exp.Sub(now).Seconds())
		}
	}

	return StatusInfo{
		Enabled:             m.data.Enabled,
		Authenticated:       authenticated,
		HasMasterPassword:   m.data.PasswordHash != "",
		HasTempPassword:     hasTemp,
		TempExpiresIn:       tempRemaining,
		TempPasswordMasked:  tempMasked,
		SessionRemainingSec: sessionRemaining,
	}
}

func (m *Manager) GetConfigPath() string {
	return m.configPath
}
