package httphandling_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/mattinordstrom/moxy/config"
	"github.com/mattinordstrom/moxy/httphandling"
)

const corsMockURL = "/api/test/mocktesting"

func useAllowedOrigins(t *testing.T, origins ...string) {
	t.Helper()

	orig := config.AppConfig.CORS.AllowedOrigins
	config.AppConfig.CORS.AllowedOrigins = origins

	t.Cleanup(func() { config.AppConfig.CORS.AllowedOrigins = orig })
}

func corsRequest(t *testing.T, method string, host string, origin string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, corsMockURL, nil)
	req.Host = host

	if origin != "" {
		req.Header.Set(headerOrigin, origin)
	}

	res := httptest.NewRecorder()
	httphandling.RootHandler(res, req)

	return res
}

//nolint:paralleltest // swaps the global config
func TestCORSAllowsLocalAndListedOrigins(t *testing.T) {
	useAllowedOrigins(t, "http://myapp.local:3000/", "https://dev.example")

	origins := map[string]string{
		"localhost":         "http://localhost:3000",
		"127.0.0.1":         "http://127.0.0.1:5173",
		"ipv6 loopback":     "http://[::1]:8080",
		"listed with slash": "http://myapp.local:3000",
		"listed":            "https://dev.example",
		"same origin":       "http://moxy.lan:9097",
	}

	for name, origin := range origins {
		t.Run(name, func(t *testing.T) {
			for _, method := range []string{http.MethodGet, http.MethodOptions} {
				res := corsRequest(t, method, "moxy.lan:9097", origin)

				if got := res.Header().Get("Access-Control-Allow-Origin"); got != origin {
					t.Errorf("%s: Access-Control-Allow-Origin = %q, want %q", method, got, origin)
				}
			}

			if res := corsRequest(t, http.MethodGet, "moxy.lan:9097", origin); res.Code != http.StatusNotFound {
				t.Errorf("expected the mock's 404, got %d", res.Code)
			}
		})
	}
}

//nolint:paralleltest // swaps the global config
func TestCORSBlocksOtherOrigins(t *testing.T) {
	useAllowedOrigins(t, "https://dev.example")

	origins := []string{
		"https://evil.example",
		"http://localhost.evil.example",
		"http://127.0.0.1.evil.example",
		"https://dev.example.evil.example",
		"null",
	}

	for _, origin := range origins {
		t.Run(origin, func(t *testing.T) {
			for _, method := range []string{http.MethodGet, http.MethodOptions} {
				res := corsRequest(t, method, adminHost, origin)

				if res.Code != http.StatusForbidden {
					t.Errorf("%s: expected 403, got %d", method, res.Code)
				}

				if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
					t.Errorf("%s: expected no CORS header, got %q", method, got)
				}
			}
		})
	}
}

//nolint:paralleltest // swaps the global config
func TestCORSWildcardAllowsAnyOrigin(t *testing.T) {
	useAllowedOrigins(t, "*")

	res := corsRequest(t, http.MethodGet, adminHost, "https://anything.example")
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "https://anything.example" {
		t.Errorf("Access-Control-Allow-Origin = %q", got)
	}
}

//nolint:paralleltest // swaps the global config
func TestCORSWithoutOriginServesRequest(t *testing.T) {
	useAllowedOrigins(t)

	res := corsRequest(t, http.MethodGet, adminHost, "")
	if res.Code != http.StatusNotFound {
		t.Errorf("expected the mock's 404, got %d", res.Code)
	}
}

//nolint:paralleltest // swaps the global config
func TestWSMockChecksOrigin(t *testing.T) {
	useAllowedOrigins(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/moxywsmock", httphandling.HandleWSMock)

	moxy := httptest.NewServer(mux)
	t.Cleanup(moxy.Close)

	wsURL := "ws" + strings.TrimPrefix(moxy.URL, "http") + "/moxywsmock"
	dial := func(origin string) error {
		conn, res, err := websocket.DefaultDialer.Dial(wsURL, http.Header{headerOrigin: {origin}})
		if res != nil {
			res.Body.Close()
		}

		if err == nil {
			conn.Close()
		}

		return err //nolint:wrapcheck // only checked for nil
	}

	if err := dial("http://localhost:3000"); err != nil {
		t.Errorf("local origin rejected: %v", err)
	}

	if err := dial("https://evil.example"); err == nil {
		t.Error("other origin was accepted")
	}
}
