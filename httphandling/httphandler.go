package httphandling

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattinordstrom/moxy/models"
	"github.com/mattinordstrom/moxy/utils"
)

var (
	Port                = 9097
	DefaultRoute        = ""
	DefaultPreserveHost = false
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
	AllInterfaces = allInterfaces
	listeners := listen(allInterfaces)

	addrs := make([]string, 0, len(listeners))
	for _, listener := range listeners {
		addrs = append(addrs, listener.Addr().String())
	}

	fmt.Printf("Now listening on %s...\n", strings.Join(addrs, ", "))

	transp := &http.Transport{
		// #nosec G402
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	ForwardClient = &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: transp,
	}

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
	if origin != "" {
		resWriter.Header().Set("Access-Control-Allow-Origin", origin)
	} else {
		resWriter.Header().Set("Access-Control-Allow-Origin", "*")
	}

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
	mockObjArr, mockErr := utils.GetMockJSON()
	if mockErr != nil {
		utils.LogError("Failed to load mock definitions: ", mockErr)
		resWriter.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(resWriter, "Error loading mock definitions")

		return
	}

	proxyObjArr, proxyErr := utils.GetProxyJSON()
	if proxyErr != nil {
		utils.LogError("Failed to load proxy definitions: ", proxyErr)
		resWriter.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(resWriter, "Error loading proxy definitions")

		return
	}

	reqURL := fmt.Sprint(req.URL)

	if isAdminPath(req) {
		handleAdminReq(resWriter, req)

		return
	}

	////////////////////////////////////////
	// Loop mockdef json obj
	for _, mockEntity := range mockObjArr {
		trimmedURL := strings.TrimRight(strings.TrimLeft(mockEntity.URLPart, "/"), "/")
		isActiveAndMethodMatch := mockEntity.Active && (req.Method == mockEntity.Method)
		isMatch := isActiveAndMethodMatch && strings.Contains(reqURL, trimmedURL)

		if strings.Contains(trimmedURL, ".*") && !isMatch {
			regex, err := regexp.Compile(trimmedURL)
			if err != nil {
				utils.LogError("Error compiling regex: ", err)
				http.Error(resWriter, "Error compiling regex in mock definition", http.StatusInternalServerError)

				return
			}

			isMatch = isActiveAndMethodMatch && regex.MatchString(reqURL)
		}

		if isMatch {
			freezetime := mockEntity.FreezeTimeMS
			if freezetime != 0 {
				dur, _ := time.ParseDuration(strconv.Itoa(freezetime) + "ms")
				time.Sleep(dur)
			}

			resWriter.Header().Set("Content-Type", "application/json")
			resWriter.WriteHeader(mockEntity.StatusCode)

			if !utils.UsePayloadFromFile(mockEntity) {
				jsonPayload, err := json.Marshal(mockEntity.Payload)
				if err != nil {
					utils.LogError("Error occurred during marshalling: ", err)

					return
				}

				_, wErr := resWriter.Write(jsonPayload)
				if wErr != nil {
					utils.LogError("Error occurred during write: ", wErr)

					return
				}

				updateAdminWithLatest(utils.GetMockEventString(mockEntity, false, string(jsonPayload)),
					utils.EventTypeMock, map[string]any{})

				fmt.Println(utils.GetMockEventString(mockEntity, true, string(jsonPayload)))
			} else {
				// Payload is a json from separate file
				payloadPath := mockEntity.PayloadFromFile

				payloadFromFile, err := utils.GetJSONPayloadFromAbsolutePath(payloadPath)
				if err != nil {
					utils.LogError("Error: ", err)
					updateAdminWithLatest(err.Error(), utils.EventTypeError, map[string]any{})

					return
				}

				_, wErr := resWriter.Write(payloadFromFile)
				if wErr != nil {
					utils.LogError("Error occurred during write (payloadFromFile): ", wErr)

					return
				}

				updateAdminWithLatest(utils.GetMockEventString(mockEntity, false, payloadPath),
					utils.EventTypeMock, map[string]any{})

				fmt.Println(utils.GetMockEventString(mockEntity, true, payloadPath))
			}

			return
		}
	}

	////////////////////////////////////////
	// Loop proxydef json obj
	useProxyForReq(resWriter, req, proxyObjArr, reqURL)
}

func useProxyForReq(resWriter http.ResponseWriter, req *http.Request, objArr []models.Proxy, reqURL string) {
	newURL := ""
	preserveHost := DefaultPreserveHost

	for _, proxyEntity := range objArr {
		trimmedURL := strings.TrimRight(strings.TrimLeft(proxyEntity.URLPart, "/"), "/")
		isMatch := proxyEntity.Active && strings.Contains(reqURL, trimmedURL)

		if strings.Contains(trimmedURL, ".*") && !isMatch {
			regex, err := regexp.Compile(trimmedURL)
			if err != nil {
				utils.LogError("Error compiling regex (proxy): ", err)
				http.Error(resWriter, "Error compiling regex in proxy definition", http.StatusInternalServerError)

				return
			}

			isMatch = proxyEntity.Active && regex.MatchString(reqURL)
		}

		if isMatch {
			newURL = strings.TrimRight(proxyEntity.Target, "/") + "/" + strings.TrimLeft(reqURL, "/")
			preserveHost = proxyEntity.PreserveHost

			if proxyEntity.Verbose {
				fmt.Println(" ")
				fmt.Println(utils.GetProxyEventString(reqURL, proxyEntity.Target, req.Method+" ", false))
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

				fmt.Printf("\x1B[3m%s\x1B[23m\n", headerString)

				// body
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					utils.LogError("Error reading body for logging:", err)

					return
				}

				bodyStr := string(bodyBytes)
				fmt.Println(bodyStr)

				extras := map[string]any{"body": bodyStr, "headers": headerString, "httpMethod": req.Method}

				updateAdminWithLatest(reqURL+utils.RightArrow+newURL, utils.EventTypeProxy, extras)

				// Restore the body for further processing
				req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

				fmt.Println(" ")
			} else {
				updateAdminWithLatest(reqURL+utils.RightArrow+newURL, utils.EventTypeProxy, map[string]any{})
				fmt.Println(utils.GetProxyEventString(reqURL, proxyEntity.Target, "", false))
			}

			break
		}
	}

	if newURL == "" {
		newURL = DefaultRoute + reqURL
		updateAdminWithLatest(reqURL+utils.RightArrow+newURL, utils.EventTypeProxy, map[string]any{})
		fmt.Println(utils.GetProxyEventString(reqURL, DefaultRoute, "", true))
	}

	forwardReq(resWriter, req, newURL, preserveHost)
}

func forwardReq(resWriter http.ResponseWriter, req *http.Request, newURL string, preserveHost bool) {
	freq, err := createReqFromReq(req, newURL, preserveHost)
	if err != nil {
		utils.LogError("", err)
		http.Error(resWriter, "Error: Failed to create request for "+newURL, http.StatusBadGateway)
		updateAdminWithLatest("Error: Failed to create request for "+newURL, utils.EventTypeError, map[string]any{})

		return
	}

	fresp, resperr := ForwardClient.Do(freq) //nolint:gosec // SSRF is intentional — Moxy is a proxy that forwards to configured targets
	if resperr != nil {
		utils.LogError("", resperr)
		http.Error(resWriter, "Error: No response from "+newURL, http.StatusBadGateway)

		updateAdminWithLatest("Error: No response from "+newURL, utils.EventTypeError, map[string]any{})

		return
	}

	defer fresp.Body.Close()

	for key, values := range fresp.Header {
		// Browsers reject responses with two CORS values, so moxy's own CORS headers win
		if strings.HasPrefix(key, "Access-Control-") && resWriter.Header().Get(key) != "" {
			continue
		}

		for _, value := range values {
			resWriter.Header().Add(key, value)
		}
	}

	resWriter.WriteHeader(fresp.StatusCode)

	// NOTE: Use this to temporarily log response body for specific URL
	urlMatchForLogging := "api/someurl-test-123-test"
	if strings.Contains(newURL, urlMatchForLogging) {
		bodyBytes, err := io.ReadAll(fresp.Body)
		if err != nil {
			utils.LogError("Failed to read response body: ", err)
			updateAdminWithLatest("Failed to read response body", utils.EventTypeError, map[string]any{})

			return
		}

		fmt.Printf("\n\x1b[36m[Response body for %s]\x1b[0m\n", urlMatchForLogging)
		fmt.Println(string(bodyBytes))

		if _, err := resWriter.Write(bodyBytes); err != nil {
			utils.LogError("Failed to write response body: ", err)
			updateAdminWithLatest("Failed to write response body", utils.EventTypeError, map[string]any{})
		}
	} else if _, err := io.Copy(resWriter, fresp.Body); err != nil {
		utils.LogError("Failed to copy response body: ", err)
		updateAdminWithLatest("Failed to copy response body", utils.EventTypeError, map[string]any{})
	}
}
