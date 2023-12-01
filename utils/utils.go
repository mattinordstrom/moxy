package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mattinordstrom/moxy/models"
)

var ColorRed = "\033[31m"
var ColorGreen = "\033[32m"
var ColorPurple = "\033[35m"
var ColorGray = "\033[37m"
var ColorBold = "\033[1m"
var ColorReset = "\033[0m"

var RightArrow = "  \u2794  "

var MockFile = "mockdef.json"
var ProxyFile = "proxydef.json"

func UsePayloadFromFile(mockEntity models.Mock) bool {
	payload := mockEntity.Payload
	payloadFromFile := mockEntity.PayloadFromFile

	usePayloadFromFile := false
	payloadFromFileSet := payloadFromFile != ""

	if (payload == nil || payload == "") && payloadFromFileSet {
		usePayloadFromFile = true
	}

	return usePayloadFromFile
}

func GetMockEventString(mockEntity models.Mock, withColor bool, payload string) string {
	rawString := mockEntity.Method +
		" " +
		mockEntity.URLPart +
		"  \u2794  [" +
		strconv.Itoa(mockEntity.FreezeTimeMS) +
		" ms] " +
		strconv.Itoa(mockEntity.StatusCode) +
		" " + payload

	if withColor {
		return ColorPurple + rawString + ColorReset
	}

	return rawString
}

func GetMockJSON() []models.Mock {
	var result []models.Mock
	byteValue, _ := getJSONResultBytes(MockFile, false)

	err := json.Unmarshal(byteValue, &result)
	if err != nil {
		log.Fatalf("Error occurred during unmarshalling (obj) in file %s: %v", MockFile, err)
	}

	return result
}

func GetProxyJSON() []models.Proxy {
	var result []models.Proxy
	byteValue, _ := getJSONResultBytes(ProxyFile, false)

	err := json.Unmarshal(byteValue, &result)
	if err != nil {
		log.Fatalf("Error occurred during unmarshalling (obj) in file %s: %v", ProxyFile, err)
	}

	return result
}

func GetJSONPayloadFromAbsolutePath(absoluteFilePath string) ([]byte, error) {
	return getJSONResultBytes(absoluteFilePath, true)
}

func getJSONResultBytes(filename string, absolutePath bool) ([]byte, error) {
	var fullJSONPath string
	var jsonPath string
	var err error

	if absolutePath {
		fullJSONPath = filename
	} else {
		jsonPath, err = os.Getwd()
		if err != nil {
			panic(err)
		}

		// When debugging tests the path could differ? Cleanup
		jsonPath = strings.ReplaceAll(jsonPath, "/httphandling", "")

		fullJSONPath = jsonPath + "/" + filename
	}

	byteValue, err := readJSONFileWithCache(fullJSONPath)

	return byteValue, err
}

func LogError(err error) {
	const red = "\033[31m"
	const reset = "\033[0m"

	fmt.Println(red, "Error:", err, reset)
}
