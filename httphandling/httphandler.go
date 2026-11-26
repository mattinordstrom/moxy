package httphandling

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattinordstrom/moxy/models"
	"github.com/mattinordstrom/moxy/utils"
)

var (
	Port            = 9097
	DefaultRoute    = ""
	DefaultInsecure = false
)

var (
	ServerReadTimeout  = 60
	ServerWriteTimeout = 60
	ServerIdleTimeout  = 45
)

func ListenAddrs(allInterfaces bool) []string {
	port := strconv.Itoa(Port)
	if allInterfaces {
		return []string{":" + port}
	}

	return []string{net.JoinHostPort("127.0.0.1", port), net.JoinHostPort("::1", port)}
}

func listen(allInterfaces bool) []net.Listener {
	var listeners []net.Listener

	for i, addr := range ListenAddrs(allInterfaces) {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			// The IPv6 loopback is optional since IPv6 can be disabled on the host
			if i > 0 {
				log.Printf("Not listening on %s: %v", addr, err)

				continue
			}

			log.Fatalf("Failed to listen on %s: %v", addr, err)
		}

		listeners = append(listeners, listener)
	}

	return listeners
}

func CreateHTTPListener(sFlag bool, allInterfaces bool) {
	listeners := listen(allInterfaces)

	addrs := make([]string, 0, len(listeners))
	for _, listener := range listeners {
		addrs = append(addrs, listener.Addr().String())
	}

	fmt.Printf("Now listening on %s...\n", strings.Join(addrs, ", "))

	fs := http.FileServer(http.Dir("ui/static"))
	http.Handle("/ui/static/", http.StripPrefix("/ui/static/", revalidate(fs)))
	http.HandleFunc("/moxyws", handleWebSocket)

	http.HandleFunc("/moxywsmock", handleWebSocketWSMock)

	go handleWSMockMessages()

	go watchDefFiles()

	// http.HandleFunc("/", HTTPHandler)
	http.HandleFunc("/", RootHandler)

	server := &http.Server{
		ReadTimeout:  time.Duration(ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(ServerWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(ServerIdleTimeout) * time.Second,
	}

	serveErrs := make(chan error, len(listeners))

	for _, listener := range listeners {
		go func() {
			if sFlag {
				serveErrs <- server.ServeTLS(listener, "https/moxyserver.crt", "https/moxyserver.key")
			} else {
				serveErrs <- server.Serve(listener)
			}
		}()
	}

	log.Fatalf("Server stopped: %v", <-serveErrs)
}

func RootHandler(resWriter http.ResponseWriter, req *http.Request) {
	// The admin API must never get CORS headers, other sites could then read and change it
	if isAdminPath(req) {
		HTTPHandler(resWriter, req)

		return
	}

	origin := req.Header.Get("Origin")
	if origin == "" {
		HTTPHandler(resWriter, req)

		return
	}

	if !originAllowed(origin, req.Host) {
		reportBlockedOrigin(origin)
		http.Error(resWriter, "Forbidden: origin "+origin+" is not in cors.allowedOrigins in config.yml", http.StatusForbidden)

		return
	}

	resWriter.Header().Set("Access-Control-Allow-Origin", origin)
	resWriter.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	resWriter.Header().Set("Access-Control-Allow-Headers", "*")
	resWriter.Header().Set("Access-Control-Allow-Credentials", "true")

	// Handle preflight
	if req.Method == http.MethodOptions {
		resWriter.WriteHeader(http.StatusOK)

		return
	}

	HTTPHandler(resWriter, req)
}

// Without this the browser heuristically caches the admin UI js and silently runs a stale build.
func revalidate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resWriter http.ResponseWriter, req *http.Request) {
		resWriter.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(resWriter, req)
	})
}

func HTTPHandler(resWriter http.ResponseWriter, req *http.Request) {
	// Before loading the defs, so a broken mockdef.json can still be fixed from the admin UI
	if isAdminPath(req) {
		handleAdminReq(resWriter, req)

		return
	}

	mockObjArr, mockErr := utils.GetMockJSON()
	if mockErr != nil {
		utils.LogError("Failed to load mock definitions: ", mockErr)
		recordEvent(LogEntry{
			Type: utils.EventTypeError, Method: req.Method, URL: redactURL(req.URL.RequestURI()),
			Status: http.StatusInternalServerError, Message: "Failed to load mock definitions: " + mockErr.Error(),
		})
		resWriter.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(resWriter, "Error loading mock definitions")

		return
	}

	proxyObjArr, proxyErr := utils.GetProxyJSON()
	if proxyErr != nil {
		utils.LogError("Failed to load proxy definitions: ", proxyErr)
		recordEvent(LogEntry{
			Type: utils.EventTypeError, Method: req.Method, URL: redactURL(req.URL.RequestURI()),
			Status: http.StatusInternalServerError, Message: "Failed to load proxy definitions: " + proxyErr.Error(),
		})
		resWriter.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(resWriter, "Error loading proxy definitions")

		return
	}

	// RequestURI, since a proxy-style absolute URL would otherwise end up inside the target URL
	reqURL := req.URL.RequestURI()

	////////////////////////////////////////
	// Loop mockdef json obj
	for _, mockEntity := range mockObjArr {
		if mockEntity.Active && req.Method == mockEntity.Method && urlPartMatches(mockEntity.URLPart, reqURL) {
			serveMock(resWriter, req, mockEntity)

			return
		}
	}

	////////////////////////////////////////
	// Loop proxydef json obj
	useProxyForReq(resWriter, req, proxyObjArr, reqURL)
}

func useProxyForReq(resWriter http.ResponseWriter, req *http.Request, objArr []models.Proxy, reqURL string) {
	newURL := ""
	termURL := utils.SanitizeTerminal(reqURL)
	rule := defaultRouteRule
	insecure := DefaultInsecure

	for _, proxyEntity := range objArr {
		if proxyEntity.Active && urlPartMatches(proxyEntity.URLPart, reqURL) {
			newURL = strings.TrimRight(proxyEntity.Target, "/") + "/" + strings.TrimLeft(reqURL, "/")
			rule = proxyEntity.URLPart
			insecure = proxyEntity.Insecure

			if proxyEntity.Verbose {
				fmt.Println(" ")
				fmt.Println(utils.GetProxyEventString(termURL, proxyEntity.Target, utils.SanitizeTerminal(req.Method)+" ", false))
				// headers
				var sBuilder strings.Builder

				for name, values := range req.Header {
					hValues := strings.Join(values, ", ")
					fmt.Fprintf(&sBuilder, "%s: %s | ", name, hValues)
				}

				headerString := sBuilder.String()
				if sBuilder.Len() > 0 {
					headerString = headerString[:len(headerString)-2]
				}

				fmt.Printf("\x1B[3m%s\x1B[23m\n", utils.SanitizeTerminal(headerString))

				// body
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					utils.LogError("Error reading body for logging:", err)

					return
				}

				bodyStr := string(bodyBytes)
				fmt.Println(utils.SanitizeTerminal(bodyStr))

				extras := map[string]any{"body": bodyStr, "headers": headerString, "httpMethod": req.Method}

				updateAdminWithLatest(reqURL+utils.RightArrow+newURL, utils.EventTypeProxy, extras)

				// Restore the body for further processing
				req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

				fmt.Println(" ")
			} else {
				updateAdminWithLatest(reqURL+utils.RightArrow+newURL, utils.EventTypeProxy, map[string]any{})
				fmt.Println(utils.GetProxyEventString(termURL, proxyEntity.Target, "", false))
			}

			break
		}
	}

	if newURL == "" {
		newURL = DefaultRoute + reqURL
		updateAdminWithLatest(reqURL+utils.RightArrow+newURL, utils.EventTypeProxy, map[string]any{})
		fmt.Println(utils.GetProxyEventString(termURL, DefaultRoute, "", true))
	}

	forwardReq(resWriter, req, newURL, rule, insecure)
}

func forwardReq(resWriter http.ResponseWriter, req *http.Request, newURL string, rule string, insecure bool) {
	start := time.Now()
	entry := LogEntry{
		Type: utils.EventTypeProxy, Method: req.Method, URL: redactURL(req.URL.RequestURI()),
		Rule: rule, Target: redactURL(newURL),
	}

	target, err := url.Parse(newURL)
	if err != nil || target.Host == "" {
		utils.LogError("Invalid proxy target "+strconv.Quote(newURL)+":", err)
		http.Error(resWriter, "Error: Failed to create request for "+newURL, http.StatusBadGateway)
		updateAdminWithLatest("Error: Failed to create request for "+newURL, utils.EventTypeError, map[string]any{})

		entry.Type, entry.Status, entry.Message = utils.EventTypeError, http.StatusBadGateway, "Invalid proxy target"
		recordEvent(entry)

		return
	}

	clearWriteDeadline(resWriter)

	transport := ForwardTransport
	if insecure {
		transport = InsecureForwardTransport
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(proxyReq *httputil.ProxyRequest) {
			rewriteProxyRequest(proxyReq, target)
		},
		Transport: transport,
		// Flush every write so streamed responses reach the client as they arrive
		FlushInterval: -1,
		ModifyResponse: func(res *http.Response) error {
			entry.Status, entry.DurationMS = res.StatusCode, millisSince(start)
			recordEvent(entry)

			// Browsers reject responses with two CORS values, so moxy's own CORS headers win
			for key := range res.Header {
				if strings.HasPrefix(key, "Access-Control-") && resWriter.Header().Get(key) != "" {
					res.Header.Del(key)
				}
			}

			return nil
		},
		ErrorHandler: func(resWriter http.ResponseWriter, req *http.Request, err error) {
			entry.DurationMS = millisSince(start)

			if errors.Is(req.Context().Err(), context.Canceled) {
				entry.Message = "Client cancelled before the target answered"
				recordEvent(entry)

				return
			}

			utils.LogError("", err)

			msg := "Error: No response from " + newURL
			entry.Message = "No response: " + err.Error()

			if certErr, ok := errors.AsType[*tls.CertificateVerificationError](err); ok {
				hint := insecureHint(rule)
				msg = "Error: Untrusted certificate from " + newURL + ": " + certErr.Err.Error() + ". " + hint
				entry.Message = "Untrusted certificate: " + certErr.Err.Error() + ". " + hint
			}

			http.Error(resWriter, msg, http.StatusBadGateway)
			updateAdminWithLatest(msg, utils.EventTypeError, map[string]any{})

			entry.Type, entry.Status = utils.EventTypeError, http.StatusBadGateway
			recordEvent(entry)
		},
		ErrorLog: log.New(io.Discard, "", 0),
	}

	proxy.ServeHTTP(resWriter, req)
}

func insecureHint(rule string) string {
	if rule == defaultRouteRule {
		return "If you trust it, set insecure: true under defaults in config.yml"
	}

	return `If you trust it, set "insecure": true on the proxy for ` + rule
}

// Event streams and dev-server sockets stay open far longer than the server's write timeout.
func clearWriteDeadline(resWriter http.ResponseWriter) {
	err := http.NewResponseController(resWriter).SetWriteDeadline(time.Time{})
	if err != nil && !errors.Is(err, http.ErrNotSupported) {
		utils.LogError("Error clearing write deadline:", err)
	}
}

var (
	urlPatterns   = map[string]*regexp.Regexp{}
	urlPatternsMu sync.Mutex
)

func urlPartMatches(urlPart string, reqURL string) bool {
	trimmedURL := strings.TrimRight(strings.TrimLeft(urlPart, "/"), "/")
	if strings.Contains(reqURL, trimmedURL) {
		return true
	}

	if !strings.Contains(trimmedURL, ".*") {
		return false
	}

	regex := compileURLPattern(trimmedURL)

	return regex != nil && regex.MatchString(reqURL)
}

// An invalid pattern only disables its own def, and is reported once instead of on every request.
func compileURLPattern(pattern string) *regexp.Regexp {
	urlPatternsMu.Lock()
	defer urlPatternsMu.Unlock()

	regex, known := urlPatterns[pattern]
	if known {
		return regex
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		utils.LogError("Ignoring def with invalid urlpart regex "+strconv.Quote(pattern)+":", err)
		recordEvent(LogEntry{Type: utils.EventTypeError, Rule: pattern, Message: "Ignoring def with invalid urlpart regex: " + err.Error()})
		updateAdminWithLatest("Ignoring def with invalid urlpart regex "+strconv.Quote(pattern)+": "+err.Error(),
			utils.EventTypeError, map[string]any{})
	}

	urlPatterns[pattern] = regex

	return regex
}

const (
	minStatusCode = 100
	maxStatusCode = 999
)

func serveMock(resWriter http.ResponseWriter, req *http.Request, mockEntity models.Mock) {
	start := time.Now()
	entry := LogEntry{Type: utils.EventTypeMock, Method: req.Method, URL: redactURL(req.URL.RequestURI()), Rule: mockEntity.URLPart}

	if mockEntity.FreezeTimeMS > 0 {
		time.Sleep(time.Duration(mockEntity.FreezeTimeMS) * time.Millisecond)
	}

	if mockEntity.StatusCode == 0 {
		mockEntity.StatusCode = http.StatusOK
	}

	// WriteHeader panics outside this range
	if mockEntity.StatusCode < minStatusCode || mockEntity.StatusCode > maxStatusCode {
		reportMockError(resWriter, entry, fmt.Errorf("%w %d in mock for %s", errInvalidStatusCode,
			mockEntity.StatusCode, mockEntity.URLPart))

		return
	}

	body, logPayload, err := mockBody(mockEntity)
	if err != nil {
		reportMockError(resWriter, entry, err)

		return
	}

	resWriter.Header().Set("Content-Type", "application/json")
	resWriter.WriteHeader(mockEntity.StatusCode)

	_, err = resWriter.Write(body)
	if err != nil {
		utils.LogError("Error occurred during write: ", err)

		return
	}

	entry.Status, entry.DurationMS = mockEntity.StatusCode, millisSince(start)
	recordEvent(entry)

	updateAdminWithLatest(utils.GetMockEventString(mockEntity, false, logPayload), utils.EventTypeMock, map[string]any{})
	fmt.Println(utils.GetMockEventString(mockEntity, true, logPayload))
}

var errInvalidStatusCode = errors.New("invalid statuscode")

func mockBody(mockEntity models.Mock) ([]byte, string, error) {
	if !utils.UsePayloadFromFile(mockEntity) {
		body, err := json.Marshal(mockEntity.Payload)
		if err != nil {
			return nil, "", fmt.Errorf("invalid payload in mock for %s: %w", mockEntity.URLPart, err)
		}

		return body, string(body), nil
	}

	body, err := utils.GetJSONPayloadFromAbsolutePath(mockEntity.PayloadFromFile)
	if err != nil {
		return nil, "", fmt.Errorf("payloadFromFile in mock for %s: %w", mockEntity.URLPart, err)
	}

	return body, mockEntity.PayloadFromFile, nil
}

func reportMockError(resWriter http.ResponseWriter, entry LogEntry, err error) {
	entry.Type, entry.Status, entry.Message = utils.EventTypeError, http.StatusInternalServerError, err.Error()
	recordEvent(entry)

	utils.LogError("Error:", err)
	updateAdminWithLatest(err.Error(), utils.EventTypeError, map[string]any{})
	http.Error(resWriter, err.Error(), http.StatusInternalServerError)
}
