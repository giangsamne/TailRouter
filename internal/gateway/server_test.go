package gateway

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
