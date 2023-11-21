package httphandling

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mattinordstrom/moxy/utils"
)

var MockFile = "mockdef.json"
var ProxyFile = "proxydef.json"
var mock = "mock"
var proxy = "proxy"

var Port = 9097
var DefaultRoute = ""

var htmlBreak = "<br />"

func CreateHTTPListener() {
	// Start listening
	fmt.Printf("Now listening on port %s...\n", strconv.Itoa(Port))

	fs := http.FileServer(http.Dir("ui/static"))
	http.Handle("/ui/static/", http.StripPrefix("/ui/static/", fs))
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/", httpHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprint(":", Port), nil))
}

func httpHandler(resWriter http.ResponseWriter, req *http.Request) {
	objArr := utils.GetJSONObj(ProxyFile)
	mockObjArr := utils.GetJSONObj(MockFile)
	reqURL := fmt.Sprint(req.URL)

	if strings.HasPrefix(reqURL, "/moxyadminui") {
		handleAdminReq(req, resWriter)

		return
	}

	////////////////////////////////////////
	// Loop mockdef json obj
	for _, val := range mockObjArr {
		val, ok := val.(map[string]interface{})
		if !ok {
			log.Println("mockdef. expected type map[string]interface{}")

			return
		}

		urlpart := fmt.Sprint(val["urlpart"])
		active, parseError := strconv.ParseBool(fmt.Sprint(val["active"]))
		if parseError != nil {
			log.Println("parseError active flag")

			return
		}

		isMatch := active && (req.Method == fmt.Sprint(val["method"]))
		isMatch = isMatch && strings.Contains(reqURL, urlpart)

		if strings.Contains(urlpart, ".*") && !isMatch {
			regex, err := regexp.Compile(urlpart)
			if err != nil {
				fmt.Println("Error compiling regex:", err)

				return
			}

			isMatch = regex.MatchString(reqURL)
		}

		if isMatch {
			freezetime := fmt.Sprint(val["freezetimems"])
			if freezetime != "0" {
				dur, _ := time.ParseDuration(freezetime + "ms")
				time.Sleep(dur)
			}
			statuscode, _ := strconv.Atoi(fmt.Sprint(val["statuscode"]))
			resWriter.Header().Set("Content-Type", "application/json")
			resWriter.WriteHeader(statuscode)

			payload := val["payload"]
			payloadFromFile := val["payloadFromFile"]

			usePayloadFromFile := false
			payloadFromFileSet := payloadFromFile != nil && payloadFromFile != ""

			if (payload == nil || payload == "") && payloadFromFileSet {
				usePayloadFromFile = true
			}

			if !usePayloadFromFile {
				if payload, ok := val["payload"].(map[string]interface{}); ok {
					// Payload is a json
					if jsonPayload, err := json.Marshal(payload); err == nil {
						_, wErr := resWriter.Write(jsonPayload)
						if wErr != nil {
							log.Fatalf("Error occurred during write: %v", wErr)
						}
					} else {
						log.Fatalf("Error couldnt marshal payload to JSON: %v", err)
					}
				} else {
					// Payload is not a json
					payload := val["payload"]
					_, err := resWriter.Write([]byte(fmt.Sprint(payload)))
					if err != nil {
						log.Fatalf("Error occurred during write (non json payload): %v", err)
					}
				}

				payloadM, err := json.Marshal(val["payload"])
				if err != nil {
					log.Fatalf("Error occurred during marshalling: %v", err)
				}

				updateAdminWithLatest(utils.GetMockEventString(val, false, string(payloadM)), mock)
				fmt.Println(utils.GetMockEventString(val, true, string(payloadM)))
			} else {
				// Payload is a json from separate file
				payloadPath := val["payloadFromFile"]

				payloadFromFile := utils.GetJSONObjAsString(fmt.Sprint(payloadPath))
				_, err := resWriter.Write([]byte(payloadFromFile))
				if err != nil {
					log.Fatalf("Error occurred during write: %v", err)
				}

				updateAdminWithLatest(utils.GetMockEventString(val, false, fmt.Sprint(payloadPath)), mock)
				fmt.Println(utils.GetMockEventString(val, true, fmt.Sprint(payloadPath)))
			}

			return
		}
	}

	////////////////////////////////////////
	// Loop proxydef json obj
	useProxyForReq(resWriter, req, objArr, reqURL)
}

func useProxyForReq(resWriter http.ResponseWriter, req *http.Request, objArr []interface{}, reqURL string) {
	newURL := ""
	for _, val := range objArr {
		val, ok := val.(map[string]interface{})
		if !ok {
			log.Fatalln("proxydef. expected type map[string]interface{}")
		}

		urlpart := fmt.Sprint(val["urlpart"])
		active, parseError := strconv.ParseBool(fmt.Sprint(val["active"]))
		if parseError != nil {
			log.Fatalln("parseError active flag")
		}
		verbose, parseError := strconv.ParseBool(fmt.Sprint(val["verbose"]))
		if parseError != nil {
			log.Fatalln("parseError verbose flag")
		}

		isMatch := active && strings.Contains(reqURL, urlpart)
		if strings.Contains(urlpart, ".*") && !isMatch {
			regex, err := regexp.Compile(urlpart)
			if err != nil {
				fmt.Println("Error compiling regex (proxy):", err)

				return
			}

			isMatch = regex.MatchString(reqURL)
		}

		if isMatch {
			newURL = fmt.Sprint(val["target"]) + reqURL

			if verbose {
				fmt.Println(" ")
				fmt.Println(utils.ColorGreen + req.Method + " " + reqURL + utils.RightArrow + newURL + utils.ColorReset)

				// headers
				var sb strings.Builder
				for name, values := range req.Header {
					sb.WriteString(fmt.Sprintf("%s: %s, ", name, strings.Join(values, ", ")))
				}
				headerString := sb.String()
				if sb.Len() > 0 {
					headerString = headerString[:len(headerString)-2]
				}

				fmt.Printf("\x1B[3m%s\x1B[23m\n", headerString)

				// body
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					fmt.Println("Error reading body for logging:", err)

					return
				}
				if bodyStr := string(bodyBytes); bodyStr != "" {
					fmt.Println(bodyStr)

					updateAdminWithLatest(htmlBreak+
						req.Method+" "+reqURL+utils.RightArrow+newURL+
						htmlBreak+headerString+
						htmlBreak+bodyStr+htmlBreak, proxy)
				} else {
					updateAdminWithLatest(htmlBreak+
						req.Method+" "+reqURL+utils.RightArrow+newURL+
						htmlBreak+headerString+htmlBreak, proxy)
				}

				// Restore the body for further processing
				req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

				fmt.Println(" ")
			} else {
				updateAdminWithLatest(reqURL+utils.RightArrow+newURL, proxy)
				fmt.Println(utils.ColorGreen + reqURL + utils.RightArrow + newURL + utils.ColorReset)
			}

			break
		}
	}

	if newURL == "" {
		newURL = DefaultRoute + reqURL
		updateAdminWithLatest(reqURL+utils.RightArrow+newURL, proxy)
		fmt.Println(utils.ColorGray + reqURL + utils.RightArrow + newURL + utils.ColorReset)
	}

	forwardReq(resWriter, req, newURL)
}

func forwardReq(resWriter http.ResponseWriter, req *http.Request, newURL string) {
	freq, client := createReqFromReq(req, newURL)

	fresp, resperr := client.Do(freq)
	if resperr != nil {
		log.Println(utils.ColorRed, resperr, utils.ColorReset)
		fmt.Fprintf(resWriter, "Error. No response")

		return
	}

	defer fresp.Body.Close()

	// Check if it's a 302 redirect
	if fresp.StatusCode == http.StatusFound {
		resWriter.Header().Set("Location", fresp.Header.Get("Location"))
	}

	body, ioerr := io.ReadAll(fresp.Body)
	if ioerr != nil {
		log.Println(utils.ColorRed, ioerr, utils.ColorReset)
		fmt.Fprintf(resWriter, "IO Error (Response body)")

		return
	}

	resWriter.Header().Set("Content-Type", fresp.Header.Get("Content-Type"))
	resWriter.WriteHeader(fresp.StatusCode)

	if _, err := resWriter.Write(body); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
