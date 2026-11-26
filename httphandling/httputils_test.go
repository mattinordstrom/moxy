package httphandling_test

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/mattinordstrom/moxy/httphandling"
)

func TestRewriteProxyRequestUsesTargetHost(t *testing.T) {
	t.Parallel()

	const targetHost = "route-utv.fnox.se"

	for _, incomingHost := range []string{"apps-develop.alfa.fnox.se", adminHost, ""} {
		t.Run(incomingHost, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "/api/common/companysettings", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Host = incomingHost
			req.Header.Set("X-Token", "abc")

			target, err := url.Parse("http://route-utv.fnox.se/api/common/companysettings")
			if err != nil {
				t.Fatal(err)
			}

			proxyReq := &httputil.ProxyRequest{In: req, Out: req.Clone(t.Context())}
			httphandling.RewriteProxyRequest(proxyReq, target)
			freq := proxyReq.Out

			// An empty Host makes the transport send the target's host
			if freq.Host != "" {
				t.Errorf("expected the target's Host to be used, got %q", freq.Host)
			}

			if freq.URL.Host != targetHost {
				t.Errorf("expected request to be sent to %s, got %q", targetHost, freq.URL.Host)
			}

			if freq.Header.Get("X-Token") != "abc" {
				t.Errorf("expected headers to be copied, got %v", freq.Header)
			}
		})
	}
}
