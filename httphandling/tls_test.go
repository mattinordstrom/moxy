package httphandling_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattinordstrom/moxy/httphandling"
)

func startTLSUpstream(t *testing.T) *httptest.Server {
	t.Helper()

	upstream := httptest.NewTLSServer(http.HandlerFunc(func(resWriter http.ResponseWriter, _ *http.Request) {
		resWriter.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)

	return upstream
}

func forwardTo(t *testing.T, target string, rule string, insecure bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/secure", nil)
	res := httptest.NewRecorder()
	httphandling.ForwardReq(res, req, target+"/api/secure", rule, insecure)

	return res
}

//nolint:paralleltest // reads the global event log
func TestProxyRejectsUntrustedCertificate(t *testing.T) {
	httphandling.ResetEventLog(t)

	upstream := startTLSUpstream(t)

	res := forwardTo(t, upstream.URL, "/api/secure", false)
	if res.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", res.Code)
	}

	if body := res.Body.String(); !strings.Contains(body, "Untrusted certificate") || !strings.Contains(body, `"insecure": true on the proxy for /api/secure`) {
		t.Errorf("expected a hint about the proxy's insecure flag, got %q", body)
	}

	entries := entriesOfType(readEventLog(t, "0"), "error")
	if len(entries) != 1 || !strings.HasPrefix(entries[0].Message, "Untrusted certificate") {
		t.Errorf("expected an untrusted certificate entry, got %+v", entries)
	}
}

func TestDefaultRouteCertificateHintPointsToConfig(t *testing.T) {
	t.Parallel()

	upstream := startTLSUpstream(t)

	res := forwardTo(t, upstream.URL, "(default route)", false)
	if body := res.Body.String(); !strings.Contains(body, "insecure: true under defaults in config.yml") {
		t.Errorf("expected a hint about config.yml, got %q", body)
	}
}

func TestInsecureProxyAcceptsSelfSignedCertificate(t *testing.T) {
	t.Parallel()

	upstream := startTLSUpstream(t)

	if res := forwardTo(t, upstream.URL, "/api/secure", true); res.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", res.Code, res.Body.String())
	}
}
