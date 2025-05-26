package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mattinordstrom/moxy/models"
)

var (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorPurple = "\033[35m"
	ColorYellow = "\033[33m"
	ColorGray   = "\033[37m"
	ColorBold   = "\033[1m"
	ColorReset  = "\033[0m"
)

var RightArrow = "  \u2794  "

var (
	MockFile  = "mockdef.json"
	ProxyFile = "proxydef.json"
)

var (
	EventTypeMock   = "mock"
	EventTypeProxy  = "proxy"
	EventTypeWSMock = "wsmock"
	EventTypeError  = "error"
)

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

func GetProxyEventString(url string, target string, extraInfo string, defaultRoute bool) string {
	newURL := strings.TrimRight(target, "/") + "/" + strings.TrimLeft(url, "/")

	if defaultRoute {
		return ColorGray + extraInfo + url + RightArrow + newURL + ColorReset
	}

	return ColorGreen + extraInfo + url + RightArrow + newURL + ColorReset
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

func LogError(str string, err error) {
	log.Println(ColorRed + str + " " + err.Error() + ColorReset)
}

func CopyFile(src, dst string) (bool, error) {
	if _, err := os.Stat(dst); err == nil {
		// Destination file exists, do nothing
		return true, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("error checking destination file: %w", err)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return false, fmt.Errorf("error opening source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return false, fmt.Errorf("error creating destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return false, fmt.Errorf("error copying file contents: %w", err)
	}

	err = destFile.Sync()
	if err != nil {
		return false, fmt.Errorf("error syncing destination file: %w", err)
	}

	return false, nil
}
