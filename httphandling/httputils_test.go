package httphandling_test

import (
	"net/http"
	"testing"

	"github.com/mattinordstrom/moxy/httphandling"
)

func TestCreateReqFromReqHost(t *testing.T) {
	t.Parallel()

	const (
		incomingHost = "apps-develop.alfa.fnox.se"
		targetHost   = "route-utv.fnox.se"
	)

	tests := []struct {
		name         string
		incomingHost string
		preserveHost bool
		expectedHost string
	}{
		{"target host when not preserving", incomingHost, false, targetHost},
		{"incoming host when preserving", incomingHost, true, incomingHost},
		{"incoming host with port when preserving", incomingHost + ":8443", true, incomingHost + ":8443"},
		{"target host when incoming is localhost", "localhost:9097", true, targetHost},
		{"target host when incoming is 127.0.0.1", "127.0.0.1:9097", true, targetHost},
		{"target host when incoming is ::1", "[::1]:9097", true, targetHost},
		{"target host when incoming is empty", "", true, targetHost},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "/api/common/companysettings", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Host = test.incomingHost
			req.Header.Set("X-Token", "abc")

			freq, err := httphandling.CreateReqFromReq(req, "http://route-utv.fnox.se/api/common/companysettings", test.preserveHost)
			if err != nil {
				t.Fatal(err)
			}

			if freq.Host != test.expectedHost {
				t.Errorf("expected Host %q, got %q", test.expectedHost, freq.Host)
			}

			if freq.URL.Host != targetHost {
				t.Errorf("expected request to be sent to route-utv.fnox.se, got %q", freq.URL.Host)
			}

			if freq.Header.Get("X-Token") != "abc" {
				t.Errorf("expected headers to be copied, got %v", freq.Header)
			}
		})
	}
}
