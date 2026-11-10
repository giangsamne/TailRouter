package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAuthManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tailrouter_auth_test_*")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	authPath := filepath.Join(tmpDir, "auth.json")
	mgr := NewManager(authPath)

	// Initially disabled
	if mgr.IsEnabled() {
		t.Errorf("Expected auth disabled initially")
	}

	// Set password
	if err := mgr.SetPassword("supersecret123"); err != nil {
		t.Fatalf("SetPassword failed: %v", err)
	}
	if !mgr.IsEnabled() {
		t.Errorf("Expected auth enabled after SetPassword")
	}

	// Verify wrong password
	if mgr.VerifyPassword("wrongpass") {
		t.Errorf("VerifyPassword should fail with wrong password")
	}

	// Verify correct password
	if !mgr.VerifyPassword("supersecret123") {
		t.Errorf("VerifyPassword should succeed with correct password")
	}

	// Session creation and 10-minute validation
	sess := mgr.CreateSession()
	if sess.Token == "" {
		t.Fatalf("CreateSession returned empty token")
	}
	if sess.RemainingSeconds <= 590 || sess.RemainingSeconds > 600 {
		t.Errorf("Expected session remaining ~600 seconds, got %d", sess.RemainingSeconds)
	}

	valid, remaining := mgr.ValidateSession(sess.Token)
	if !valid || remaining <= 0 {
		t.Errorf("Expected session valid, got valid=%v, remaining=%d", valid, remaining)
	}

	// Extend session
	extended, err := mgr.ExtendSession(sess.Token)
	if err != nil {
		t.Fatalf("ExtendSession failed: %v", err)
	}
	if extended.RemainingSeconds <= 590 {
		t.Errorf("Expected extended session remaining ~600 seconds, got %d", extended.RemainingSeconds)
	}

	// Revoke session
	mgr.RevokeSession(sess.Token)
	valid, _ = mgr.ValidateSession(sess.Token)
	if valid {
		t.Errorf("Expected session invalid after revocation")
	}

	// 10-minute Temporary Password (OTP)
	otp, exp, err := mgr.GenerateTempPassword()
	if err != nil {
		t.Fatalf("GenerateTempPassword failed: %v", err)
	}
	if otp == "" {
		t.Fatalf("OTP is empty")
	}
	if exp.Before(time.Now()) {
		t.Errorf("Expected OTP expiration in future")
	}
	if !mgr.VerifyPassword(otp) {
		t.Errorf("Expected OTP to verify successfully")
	}

	// File permissions test
	info, err := os.Stat(authPath)
	if err != nil {
		t.Fatalf("Stat auth.json failed: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Expected permissions 0600 on auth.json, got %04o", info.Mode().Perm())
	}
}
