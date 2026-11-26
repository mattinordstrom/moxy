package httphandling

import (
	"fmt"
	"maps"
	"net"
	"net/http"
)

var ForwardClient *http.Client

func createReqFromReq(req *http.Request, newURL string, preserveHost bool) (*http.Request, error) {
	freq, reqerror := http.NewRequest(req.Method, newURL, req.Body) //nolint:gosec // SSRF is intentional — Moxy is a proxy that forwards to configured targets
	if reqerror != nil {
		return nil, fmt.Errorf("failed to create forward request: %w", reqerror)
	}

	// Without this the body is sent chunked, which some upstreams reject
	freq.ContentLength = req.ContentLength

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

	// Go keeps the incoming Host in req.Host, not in req.Header, so it is lost unless copied explicitly.
	if preserveHost && req.Host != "" && !isLoopbackHost(req.Host) {
		freq.Host = req.Host
	}

	return freq, nil
}

func isLoopbackHost(host string) bool {
	hostname := host
	if splitHost, _, err := net.SplitHostPort(host); err == nil {
		hostname = splitHost
	}

	if hostname == "localhost" {
		return true
	}

	ip := net.ParseIP(hostname)

	return ip != nil && ip.IsLoopback()
}
