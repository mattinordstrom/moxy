package httphandling_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattinordstrom/moxy/httphandling"
)

type upstreamSeen struct {
	contentLength    int64
	transferEncoding []string
	body             string
}

func startUpstream(t *testing.T) *upstreamSeen {
	t.Helper()

	seen := &upstreamSeen{}

	upstream := httptest.NewServer(http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		seen.contentLength = req.ContentLength
		seen.transferEncoding = req.TransferEncoding
		seen.body = string(body)

		resWriter.Header().Set("Access-Control-Allow-Origin", "https://prod.example")
		resWriter.Header().Set("Access-Control-Expose-Headers", "X-Upstream")
		resWriter.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	useUpstream(t, upstream)

	return seen
}

func useUpstream(t *testing.T, upstream *httptest.Server) {
	t.Helper()

	origRoute, origTransport := httphandling.DefaultRoute, httphandling.ForwardTransport
	httphandling.DefaultRoute = upstream.URL
	httphandling.ForwardTransport = upstream.Client().Transport

	t.Cleanup(func() {
		httphandling.DefaultRoute, httphandling.ForwardTransport = origRoute, origTransport
	})
}

//nolint:paralleltest // swaps the global default route and forward client
func TestProxyKeepsContentLength(t *testing.T) {
	seen := startUpstream(t)

	body := `{"a":1}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/upstream/echo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	httphandling.RootHandler(httptest.NewRecorder(), req)

	if seen.contentLength != int64(len(body)) || len(seen.transferEncoding) != 0 {
		t.Errorf("expected Content-Length %d and no Transfer-Encoding, got %d and %v",
			len(body), seen.contentLength, seen.transferEncoding)
	}

	if seen.body != body {
		t.Errorf("expected body %q upstream, got %q", body, seen.body)
	}
}

//nolint:paralleltest // swaps the global default route and forward client
func TestProxySendsSingleCORSOrigin(t *testing.T) {
	startUpstream(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/upstream/echo", nil)
	req.Header.Set("Origin", appOrigin)

	res := httptest.NewRecorder()
	httphandling.RootHandler(res, req)

	if got := res.Header().Values("Access-Control-Allow-Origin"); len(got) != 1 || got[0] != appOrigin {
		t.Errorf("expected only moxy's CORS origin, got %v", got)
	}

	if got := res.Header().Get("Access-Control-Expose-Headers"); got != "X-Upstream" {
		t.Errorf("expected upstream CORS headers moxy doesn't set to pass through, got %q", got)
	}
}
