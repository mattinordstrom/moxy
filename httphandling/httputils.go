package httphandling

import (
	"fmt"
	"maps"
	"net/http"
)

var ForwardClient *http.Client

func createReqFromReq(req *http.Request, newURL string) (*http.Request, error) {
	freq, reqerror := http.NewRequest(req.Method, newURL, req.Body) //nolint:gosec // SSRF is intentional — Moxy is a proxy that forwards to configured targets
	if reqerror != nil {
		return nil, fmt.Errorf("failed to create forward request: %w", reqerror)
	}

	freq.Header = make(http.Header)
	maps.Copy(freq.Header, req.Header)

	// DisableCompression, if true, prevents the Transport from
	// requesting compression with an "Accept-Encoding: gzip"
	// request header when the Request contains no existing
	// Accept-Encoding value. If the Transport requests gzip on
	// its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body. However, if the user
	// explicitly requested gzip it is not automatically
	// uncompressed.
	//
	// https://go.dev/src/net/http/transport.go#L190
	freq.Header.Del("Accept-Encoding")

	return freq, nil
}
