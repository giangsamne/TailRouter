package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", "/dashboard", false},
		{"valid subpath", "/app/v1", false},
		{"empty path", "", false}, // defaults to "/"
		{"root path", "/", false},
		{"traversal dotdot", "/../secret", true},
		{"double slash", "//app", true},
		{"reserved api prefix", "/api", true},
		{"reserved api subpath", "/api/test", true},
		{"reserved router prefix", "/router", true},
		{"reserved router subpath", "/router/admin", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTargetHost(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{"valid localhost", "127.0.0.1", false},
		{"valid local name", "localhost", false},
		{"valid lan ip", "192.168.1.50", false},
		{"valid domain", "my-nas.local", false},
		{"empty host defaults to 127.0.0.1", "", false},
		{"cloud metadata ip", "169.254.169.254", true},
		{"cloud metadata prefix", "169.254.1.1", true},
		{"cloud metadata name", "metadata.google.internal", true},
		{"has port", "127.0.0.1:8080", true},
		{"has uri scheme", "http://localhost", true},
		{"has url path", "localhost/test", true},
		{"has user credentials", "user:pass@127.0.0.1", true},
		{"has control chars", "127.0.0.1\r\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateTargetHost(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTargetHost(%q) error = %v, wantErr %v", tt.host, err, tt.wantErr)
			}
		})
	}
}

func TestRouteStoreAtomicPermissions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tailrouter_test_*")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfgPath := filepath.Join(tmpDir, "sub", "routes.json")
	store := NewManager(cfgPath)

	route, err := store.Add("Test Service", "/testsvc", "127.0.0.1", 8080, false, "serve", "notes")
	if err != nil {
		t.Fatalf("Add route failed: %v", err)
	}
	if route == nil || route.ID == "" {
		t.Fatalf("Expected valid route, got nil or empty ID")
	}

	info, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatalf("Stat config failed: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("Expected config permissions 0600, got %04o", perm)
	}
}
