package httphandling_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattinordstrom/moxy/httphandling"
	testhelper "github.com/mattinordstrom/moxy/testhelpers"
)

const mockdefURL = "/moxyadminui/mockdef"

func adminRequest(t *testing.T, method, host string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, mockdefURL, strings.NewReader("not json"))
	req.Host = host
	req.RemoteAddr = localRemote

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	return testhelper.GetRecorder(req)
}

func TestAdminRejectsNonLocalhostHost(t *testing.T) {
	t.Parallel()

	for _, host := range []string{"attacker.example:9097", "10.0.0.5:9097", "localhost.attacker.example"} {
		if res := adminRequest(t, http.MethodGet, host, nil); res.Code != http.StatusForbidden {
			t.Errorf("host %s: expected 403, got %d", host, res.Code)
		}
	}

	for _, host := range []string{adminHost, ipv4Host, ipv6Host} {
		if res := adminRequest(t, http.MethodGet, host, nil); res.Code == http.StatusForbidden {
			t.Errorf("host %s: expected to be allowed, got 403", host)
		}
	}
}

func TestAdminWriteRejectsOtherOrigins(t *testing.T) {
	t.Parallel()

	tests := map[string]map[string]string{
		"other site":           {headerOrigin: "https://evil.example"},
		"other localhost port": {headerOrigin: appOrigin},
		"null origin":          {headerOrigin: "null"},
		"cross-site fetch":     {headerFetch: "cross-site"},
		"same-site fetch":      {headerFetch: "same-site"},
	}

	for name, headers := range tests {
		headers["Content-Type"] = "application/json"
		if res := adminRequest(t, http.MethodPost, adminHost, headers); res.Code != http.StatusForbidden {
			t.Errorf("%s: expected 403, got %d", name, res.Code)
		}
	}
}

func TestAdminWriteRequiresJSON(t *testing.T) {
	t.Parallel()

	for _, contentType := range []string{"", "text/plain", "application/x-www-form-urlencoded", "multipart/form-data"} {
		headers := map[string]string{headerOrigin: "http://localhost:9097", "Content-Type": contentType}
		if res := adminRequest(t, http.MethodPost, adminHost, headers); res.Code != http.StatusUnsupportedMediaType {
			t.Errorf("content type %q: expected 415, got %d", contentType, res.Code)
		}
	}
}

func TestAdminWriteAllowsSameOrigin(t *testing.T) {
	t.Parallel()

	headers := map[string]string{
		headerOrigin:   "http://" + adminHost,
		headerFetch:    "same-origin",
		"Content-Type": "application/json; charset=utf-8",
	}

	// The invalid JSON body proves the request got past the guard to the handler
	if res := adminRequest(t, http.MethodPost, adminHost, headers); res.Code != http.StatusBadRequest {
		t.Errorf("expected the handler's 400 for invalid JSON, got %d", res.Code)
	}
}

func TestCORSOnlyOnMockAndProxyRoutes(t *testing.T) {
	t.Parallel()

	preflight := func(path string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodOptions, path, nil)
		req.RemoteAddr = localRemote
		req.Host = adminHost
		req.Header.Set(headerOrigin, appOrigin)

		res := httptest.NewRecorder()
		httphandling.RootHandler(res, req)

		return res
	}

	if res := preflight("/api/anything"); res.Header().Get("Access-Control-Allow-Origin") != appOrigin {
		t.Errorf("mock/proxy route: expected CORS for the app's origin, got %q", res.Header().Get("Access-Control-Allow-Origin"))
	}

	res := preflight(mockdefURL)
	if acao := res.Header().Get("Access-Control-Allow-Origin"); acao != "" {
		t.Errorf("admin route: expected no CORS header, got %q", acao)
	}

	if res.Code != http.StatusForbidden {
		t.Errorf("admin route: expected cross-origin preflight to get 403, got %d", res.Code)
	}
}

func TestAdminOnlyReachableFromThisMachine(t *testing.T) {
	t.Parallel()

	request := func(remote string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, mockdefURL, nil)
		req.Host = adminHost
		req.RemoteAddr = remote

		return testhelper.GetRecorder(req)
	}

	for _, remote := range []string{"192.168.1.20:50000", "10.0.0.5:50000", "[fe80::1]:50000", "", "garbage"} {
		if res := request(remote); res.Code != http.StatusForbidden {
			t.Errorf("remote %q: expected 403, got %d", remote, res.Code)
		}
	}

	for _, remote := range []string{localRemote, "[::1]:50000", "[::ffff:127.0.0.1]:50000"} {
		if res := request(remote); res.Code == http.StatusForbidden {
			t.Errorf("remote %q: expected to be allowed, got 403", remote)
		}
	}
}

func TestAdminWebSocketOnlyFromThisMachine(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/moxyws", nil)
	req.Host = adminHost
	req.RemoteAddr = "192.168.1.20:50000"
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	res := httptest.NewRecorder()
	httphandling.HandleAdminWebSocket(res, req)

	if res.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", res.Code)
	}
}
