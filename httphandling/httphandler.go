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
	mockObjArr := utils.GetJSONObj(MockFile)
	proxyObjArr := utils.GetJSONObj(ProxyFile)
	reqURL := fmt.Sprint(req.URL)

	if strings.HasPrefix(reqURL, "/moxyadminui") {
		handleAdminReq(req, resWriter)

		return
	}

	////////////////////////////////////////
	// Loop mockdef json obj
	for _, val := range mockObjArr {
		// Create struct
		jsonData, err := json.Marshal(val)
		if err != nil {
			fmt.Println("error marshalling map:", err)

			return
		}
		var mockEntity models.Mock
		err = json.Unmarshal(jsonData, &mockEntity)
		if err != nil {
			fmt.Println("error unmarshalling JSON:", err)

			return
		}
		///////////

		isMatch := mockEntity.Active && (req.Method == mockEntity.Method)
		isMatch = isMatch && strings.Contains(reqURL, mockEntity.URLPart)

		if strings.Contains(mockEntity.URLPart, ".*") && !isMatch {
			regex, err := regexp.Compile(mockEntity.URLPart)
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

				payloadFromFile := utils.GetJSONObjAsString(payloadPath)
				_, err := resWriter.Write([]byte(payloadFromFile))
				if err != nil {
					log.Fatalf("Error occurred during write: %v", err)
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
