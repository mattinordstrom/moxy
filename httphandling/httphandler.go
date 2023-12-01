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

	"github.com/mattinordstrom/moxy/models"
	"github.com/mattinordstrom/moxy/utils"
)

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
	mockObjArr := utils.GetMockJSON()
	proxyObjArr := utils.GetProxyJSON()
	reqURL := fmt.Sprint(req.URL)

	if strings.HasPrefix(reqURL, "/moxyadminui") {
		handleAdminReq(req, resWriter)

		return
	}

	////////////////////////////////////////
	// Loop mockdef json obj
	for _, val := range mockObjArr {
		mockEntity := val

		trimmedURL := strings.TrimRight(strings.TrimLeft(mockEntity.URLPart, "/"), "/")
		isMatch := mockEntity.Active && (req.Method == mockEntity.Method)
		isMatch = isMatch && strings.Contains(reqURL, trimmedURL)

		if strings.Contains(trimmedURL, ".*") && !isMatch {
			regex, err := regexp.Compile(trimmedURL)
			if err != nil {
				fmt.Println("Error compiling regex:", err)

				return
			}

			isMatch = regex.MatchString(reqURL)
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
					fmt.Println("Error occurred during marshalling: ", err)
				} else {
					_, wErr := resWriter.Write(jsonPayload)
					if wErr != nil {
						fmt.Println("Error occurred during write: ", wErr)
					}
				}

				updateAdminWithLatest(utils.GetMockEventString(mockEntity, false, string(jsonPayload)), mock)
				fmt.Println(utils.GetMockEventString(mockEntity, true, string(jsonPayload)))
			} else {
				// Payload is a json from separate file
				payloadPath := mockEntity.PayloadFromFile

				payloadFromFile, err := utils.GetJSONPayloadFromAbsolutePath(payloadPath)
				if err != nil {
					utils.LogError(err)
					updateAdminWithLatest(err.Error(), "error")

					return
				}

				_, wErr := resWriter.Write(payloadFromFile)
				if wErr != nil {
					log.Fatalf("Error occurred during write: %v", wErr)
				}

				updateAdminWithLatest(utils.GetMockEventString(mockEntity, false, payloadPath), mock)
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
	for _, val := range objArr {
		proxyEntity := val

		trimmedURL := strings.TrimRight(strings.TrimLeft(proxyEntity.URLPart, "/"), "/")
		isMatch := proxyEntity.Active && strings.Contains(reqURL, trimmedURL)

		if strings.Contains(trimmedURL, ".*") && !isMatch {
			regex, err := regexp.Compile(trimmedURL)
			if err != nil {
				fmt.Println("Error compiling regex (proxy):", err)

				return
			}

			isMatch = regex.MatchString(reqURL)
		}

		if isMatch {
			newURL = proxyEntity.Target + reqURL

			if proxyEntity.Verbose {
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
