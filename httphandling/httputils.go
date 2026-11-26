package httphandling

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var (
	ForwardTransport         http.RoundTripper = newForwardTransport(false)
	InsecureForwardTransport http.RoundTripper = newForwardTransport(true)
)

const (
	dialTimeout         = 10 * time.Second
	tlsHandshakeTimeout = 10 * time.Second
)

// No overall timeout: a request stays open as long as the client waits, and is cancelled when it gives up.
func newForwardTransport(insecure bool) *http.Transport {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		transport = &http.Transport{}
	}

	transport = transport.Clone()
	transport.DialContext = (&net.Dialer{Timeout: dialTimeout}).DialContext
	transport.TLSHandshakeTimeout = tlsHandshakeTimeout

	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in per proxy for self-signed dev targets
	}

	return transport
}

var forwardedHeaders = []string{"X-Forwarded-For", "X-Forwarded-Host", "X-Forwarded-Proto"}

func rewriteProxyRequest(proxyReq *httputil.ProxyRequest, target *url.URL) {
	proxyReq.Out.URL = target
	proxyReq.Out.Host = ""

	// Rewrite strips these, but moxy has always passed on what the client sent
	for _, header := range forwardedHeaders {
		if values := proxyReq.In.Header.Values(header); len(values) > 0 {
			proxyReq.Out.Header[header] = values
		}
	}

	// Without this the transport cannot transparently decompress, and verbose logging would show gzip
	proxyReq.Out.Header.Del("Accept-Encoding")
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
