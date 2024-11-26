package httphandling

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattinordstrom/moxy/models"
	"github.com/mattinordstrom/moxy/utils"
)

var (
	Port         = 9097
	DefaultRoute = ""
)

var (
	ServerReadTimeout  = 60
	ServerWriteTimeout = 60
	ServerIdleTimeout  = 45
)

var htmlBreak = "<br />"

func CreateHTTPListener() {
	// Start listening
	fmt.Printf("Now listening on port %s...\n", strconv.Itoa(Port))

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
	http.Handle("/ui/static/", http.StripPrefix("/ui/static/", fs))
	http.HandleFunc("/moxyws", handleWebSocket)
	http.HandleFunc("/", HTTPHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", Port),
		ReadTimeout:  time.Duration(ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(ServerWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(ServerIdleTimeout) * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func HTTPHandler(resWriter http.ResponseWriter, req *http.Request) {
	mockObjArr := utils.GetMockJSON()
	proxyObjArr := utils.GetProxyJSON()
	reqURL := fmt.Sprint(req.URL)

	if strings.HasPrefix(reqURL, "/moxyadminui") {
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

				updateAdminWithLatest(utils.GetMockEventString(mockEntity, false, string(jsonPayload)), utils.EventTypeMock, map[string]interface{}{})
				fmt.Println(utils.GetMockEventString(mockEntity, true, string(jsonPayload)))
			} else {
				// Payload is a json from separate file
				payloadPath := mockEntity.PayloadFromFile

				payloadFromFile, err := utils.GetJSONPayloadFromAbsolutePath(payloadPath)
				if err != nil {
					utils.LogError("Error: ", err)
					updateAdminWithLatest(err.Error(), utils.EventTypeError, map[string]interface{}{})

					return
				}

				_, wErr := resWriter.Write(payloadFromFile)
				if wErr != nil {
					utils.LogError("Error occurred during write (payloadFromFile): ", wErr)

					return
				}

				updateAdminWithLatest(utils.GetMockEventString(mockEntity, false, payloadPath), utils.EventTypeMock, map[string]interface{}{})
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

	for _, proxyEntity := range objArr {
		trimmedURL := strings.TrimRight(strings.TrimLeft(proxyEntity.URLPart, "/"), "/")
		isMatch := proxyEntity.Active && strings.Contains(reqURL, trimmedURL)

		if strings.Contains(trimmedURL, ".*") && !isMatch {
			regex, err := regexp.Compile(trimmedURL)
			if err != nil {
				utils.LogError("Error compiling regex (proxy): ", err)

				return
			}

			isMatch = proxyEntity.Active && regex.MatchString(reqURL)
		}

		if isMatch {
			newURL = strings.TrimRight(proxyEntity.Target, "/") + "/" + strings.TrimLeft(reqURL, "/")

			if proxyEntity.Verbose {
				fmt.Println(" ")
				fmt.Println(utils.GetProxyEventString(reqURL, proxyEntity.Target, req.Method+" ", false))
				// headers
				var sBuilder strings.Builder
				for name, values := range req.Header {
					hValues := strings.Join(values, ", ")
					sBuilder.WriteString(fmt.Sprintf("%s: %s | ", name, hValues))
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

				extras := map[string]interface{}{"body": bodyStr, "headers": headerString, "httpMethod": req.Method}

				updateAdminWithLatest(reqURL+utils.RightArrow+newURL+htmlBreak, utils.EventTypeProxy, extras)

				// Restore the body for further processing
				req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

				fmt.Println(" ")
			} else {
				updateAdminWithLatest(reqURL+utils.RightArrow+newURL, utils.EventTypeProxy, map[string]interface{}{})
				fmt.Println(utils.GetProxyEventString(reqURL, proxyEntity.Target, "", false))
			}

			break
		}
	}

	if newURL == "" {
		newURL = DefaultRoute + reqURL
		updateAdminWithLatest(reqURL+utils.RightArrow+newURL, utils.EventTypeProxy, map[string]interface{}{})
		fmt.Println(utils.GetProxyEventString(reqURL, DefaultRoute, "", true))
	}

	forwardReq(resWriter, req, newURL)
}

func forwardReq(resWriter http.ResponseWriter, req *http.Request, newURL string) {
	freq := createReqFromReq(req, newURL)

	fresp, resperr := ForwardClient.Do(freq)
	if resperr != nil {
		utils.LogError("", resperr)
		fmt.Fprintf(resWriter, "Error: No response from %s", newURL)

		updateAdminWithLatest("Error: No response from "+newURL, utils.EventTypeError, map[string]interface{}{})

		return
	}

	defer fresp.Body.Close()

	for key, values := range fresp.Header {
		for _, value := range values {
			resWriter.Header().Add(key, value)
		}
	}

	resWriter.WriteHeader(fresp.StatusCode)

	if _, err := io.Copy(resWriter, fresp.Body); err != nil {
		utils.LogError("Failed to copy response body: ", err)
		updateAdminWithLatest("Failed to copy response body", utils.EventTypeError, map[string]interface{}{})
	}
}
