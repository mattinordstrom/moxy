package httphandling

import (
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

var AllInterfaces = false

var (
	errHostNotAllowed   = errors.New("host is not localhost")
	errCrossOrigin      = errors.New("cross-origin request")
	errContentTypeNotOK = errors.New("content type must be application/json")
)

func isAdminPath(req *http.Request) bool {
	return strings.HasPrefix(fmt.Sprint(req.URL), "/moxyadminui")
}

// Rejecting non-localhost hosts stops DNS rebinding from reaching the admin API.
func checkAdminHost(req *http.Request) error {
	if AllInterfaces || isLoopbackHost(req.Host) {
		return nil
	}

	return fmt.Errorf("%w: %s", errHostNotAllowed, req.Host)
}

func checkSameOrigin(req *http.Request) error {
	site := req.Header.Get("Sec-Fetch-Site")
	if site != "" && site != "same-origin" && site != "none" {
		return fmt.Errorf("%w: Sec-Fetch-Site %s", errCrossOrigin, site)
	}

	origin := req.Header.Get("Origin")
	if origin == "" {
		return nil
	}

	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" || !strings.EqualFold(originURL.Host, req.Host) {
		return fmt.Errorf("%w: Origin %s", errCrossOrigin, origin)
	}

	return nil
}

// A JSON content type forces a CORS preflight, which other sites can't pass.
func checkJSONContentType(req *http.Request) error {
	mediaType, _, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errContentTypeNotOK
	}

	return nil
}

func adminWebSocketAllowed(req *http.Request) bool {
	err := checkAdminHost(req)
	if err == nil {
		err = checkSameOrigin(req)
	}

	if err != nil {
		log.Printf("Blocked admin websocket from %s: %q", req.RemoteAddr, err.Error()) //nolint:gosec // %q escapes control characters

		return false
	}

	return true
}

func guardAdminRequest(resWriter http.ResponseWriter, req *http.Request) bool {
	err := checkAdminHost(req)

	if err == nil && req.Method != http.MethodGet && req.Method != http.MethodHead {
		err = checkSameOrigin(req)
		if err == nil {
			err = checkJSONContentType(req)
		}
	}

	if err == nil {
		return true
	}

	log.Printf("Blocked admin request %q %q from %s: %q", req.Method, req.URL.Path, req.RemoteAddr, err.Error()) //nolint:gosec // %q escapes control characters

	status := http.StatusForbidden
	if errors.Is(err, errContentTypeNotOK) {
		status = http.StatusUnsupportedMediaType
	}

	http.Error(resWriter, "Forbidden: "+err.Error(), status)

	return false
}
