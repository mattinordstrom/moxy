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
		"other localhost port": {headerOrigin: "http://localhost:3000"},
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
		req.Host = adminHost
		req.Header.Set(headerOrigin, "http://localhost:3000")

		res := httptest.NewRecorder()
		httphandling.RootHandler(res, req)

		return res
	}

	if res := preflight("/api/anything"); res.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
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
