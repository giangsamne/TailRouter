package gateway

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	tmpDir, err := os.MkdirTemp("", "tailrouter_gw_test_*")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}

	cfgPath := filepath.Join(tmpDir, "routes.json")
	srv := NewServer(65534, cfgPath)

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return srv, cleanup
}

func newLocalRequest(method, target string) *http.Request {
	req := httptest.NewRequest(method, target, nil)
	req.Host = "localhost:65534"
	return req
}

func TestSecurityHeaders(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	req := newLocalRequest("GET", "/api/status")
	w := httptest.NewRecorder()

	srv.securityMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	resp := w.Result()
	if resp.Header.Get("X-Frame-Options") != "DENY" {
		t.Errorf("Expected X-Frame-Options: DENY, got %q", resp.Header.Get("X-Frame-Options"))
	}
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("Expected X-Content-Type-Options: nosniff, got %q", resp.Header.Get("X-Content-Type-Options"))
	}
}

func TestCSRFProtection(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. POST without custom header should be rejected (403 Forbidden)
	reqNoHeader := newLocalRequest("POST", "/api/routes/reset")
	wNoHeader := httptest.NewRecorder()

	srv.securityMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})).ServeHTTP(wNoHeader, reqNoHeader)

	if wNoHeader.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden without CSRF header, got %d", wNoHeader.Code)
	}

	// 2. POST with X-TailRouter-Client: webui should pass
	reqWithHeader := newLocalRequest("POST", "/api/routes/reset")
	reqWithHeader.Header.Set("X-TailRouter-Client", "webui")
	wWithHeader := httptest.NewRecorder()

	srv.securityMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})).ServeHTTP(wWithHeader, reqWithHeader)

	if wWithHeader.Code != http.StatusOK {
		t.Errorf("Expected 200 OK with CSRF header, got %d", wWithHeader.Code)
	}
}

func TestDNSRebindingProtection(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// Attacker domain Host header should be rejected (400 Bad Request)
	req := httptest.NewRequest("GET", "/api/status", nil)
	req.Host = "attacker-domain.com"
	w := httptest.NewRecorder()

	srv.securityMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for attacker Host header, got %d", w.Code)
	}
}

func TestFunnelIsolationOnRootAndAPI(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	handler := srv.securityMiddleware(http.HandlerFunc(srv.handleRoot))

	// Attempting to access /router from public internet via Tailscale Funnel
	reqRouter := newLocalRequest("GET", "/router")
	reqRouter.Header.Set("Tailscale-Funnel-Request", "1")
	wRouter := httptest.NewRecorder()
	handler.ServeHTTP(wRouter, reqRouter)

	if wRouter.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden when Funnel accesses /router, got %d", wRouter.Code)
	}

	// Attempting to access /api/status from public internet via Tailscale Funnel
	reqAPI := newLocalRequest("GET", "/api/status")
	reqAPI.Header.Set("Tailscale-Funnel-Request", "1")
	wAPI := httptest.NewRecorder()
	handler.ServeHTTP(wAPI, reqAPI)

	if wAPI.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden when Funnel accesses /api/status, got %d", wAPI.Code)
	}
}

func TestFunnelIsolationOnPrivateRoutes(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	// Add a private "serve" route
	route, err := srv.routes.Add("Internal Wiki", "/wiki", "127.0.0.1", 8080, false, "serve", "notes")
	if err != nil {
		t.Fatalf("Add route failed: %v", err)
	}

	handler := srv.securityMiddleware(http.HandlerFunc(srv.handleRoot))

	// Funnel request attempting to reach internal wiki
	req := newLocalRequest("GET", "/wiki/index.html")
	req.Header.Set("Tailscale-Funnel-Request", "true")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden when Funnel accesses private serve route %s, got %d", route.Path, w.Code)
	}
}

func TestPasswordProtectionAnd10MinuteSession(t *testing.T) {
	srv, cleanup := setupTestServer(t)
	defer cleanup()

	handler := srv.securityMiddleware(http.HandlerFunc(srv.handleRoot))

	// 1. Enable password protection
	if err := srv.auth.SetPassword("AdminMasterPass99"); err != nil {
		t.Fatalf("SetPassword failed: %v", err)
	}

	// 2. Unauthenticated request to /api/status should return 401 Unauthorized
	reqUnauth := newLocalRequest("GET", "/api/status")
	wUnauth := httptest.NewRecorder()
	handler.ServeHTTP(wUnauth, reqUnauth)

	if wUnauth.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized without auth, got %d", wUnauth.Code)
	}

	// 3. Login with wrong password should fail
	reqBadLogin := newLocalRequest("POST", "/api/auth/login")
	reqBadLogin.Header.Set("X-TailRouter-Client", "webui")
	reqBadLogin.Body = http.NoBody
	wBadLogin := httptest.NewRecorder()
	handler.ServeHTTP(wBadLogin, reqBadLogin)

	if wBadLogin.Code != http.StatusBadRequest && wBadLogin.Code != http.StatusUnauthorized {
		t.Errorf("Expected 400 or 401 on bad login, got %d", wBadLogin.Code)
	}

	// 4. Login with correct password
	reqLogin := newLocalRequest("POST", "/api/auth/login")
	reqLogin.Header.Set("X-TailRouter-Client", "webui")
	reqLogin.Header.Set("Content-Type", "application/json")
	loginBody := `{"password":"AdminMasterPass99"}`
	reqLogin.Body = io.NopCloser(strings.NewReader(loginBody))
	wLogin := httptest.NewRecorder()
	handler.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on correct login, got %d: %s", wLogin.Code, wLogin.Body.String())
	}

	var loginResp struct {
		Success          bool   `json:"success"`
		Token            string `json:"token"`
		RemainingSeconds int    `json:"remaining_seconds"`
	}
	if err := json.NewDecoder(wLogin.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}
	if !loginResp.Success || loginResp.Token == "" {
		t.Fatalf("Expected login success with token, got: %+v", loginResp)
	}
	if loginResp.RemainingSeconds <= 590 || loginResp.RemainingSeconds > 600 {
		t.Errorf("Expected remaining seconds ~600, got %d", loginResp.RemainingSeconds)
	}

	// 5. Authenticated request using X-TailRouter-Auth
	reqAuth := newLocalRequest("GET", "/api/status")
	reqAuth.Header.Set("X-TailRouter-Auth", loginResp.Token)
	wAuth := httptest.NewRecorder()
	handler.ServeHTTP(wAuth, reqAuth)

	if wAuth.Code != http.StatusOK {
		t.Errorf("Expected 200 OK with valid session token, got %d", wAuth.Code)
	}

	// 6. Extend 10-minute session
	reqExtend := newLocalRequest("POST", "/api/auth/extend")
	reqExtend.Header.Set("X-TailRouter-Client", "webui")
	reqExtend.Header.Set("X-TailRouter-Auth", loginResp.Token)
	wExtend := httptest.NewRecorder()
	handler.ServeHTTP(wExtend, reqExtend)

	if wExtend.Code != http.StatusOK {
		t.Errorf("Expected 200 OK on extend session, got %d", wExtend.Code)
	}

	// 7. Logout revokes 10-minute session
	reqLogout := newLocalRequest("POST", "/api/auth/logout")
	reqLogout.Header.Set("X-TailRouter-Client", "webui")
	reqLogout.Header.Set("X-TailRouter-Auth", loginResp.Token)
	wLogout := httptest.NewRecorder()
	handler.ServeHTTP(wLogout, reqLogout)

	if wLogout.Code != http.StatusOK {
		t.Errorf("Expected 200 OK on logout, got %d", wLogout.Code)
	}

	// 8. Subsequent call with revoked token should fail with 401
	reqPostLogout := newLocalRequest("GET", "/api/status")
	reqPostLogout.Header.Set("X-TailRouter-Auth", loginResp.Token)
	wPostLogout := httptest.NewRecorder()
	handler.ServeHTTP(wPostLogout, reqPostLogout)

	if wPostLogout.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized after logout, got %d", wPostLogout.Code)
	}
}
