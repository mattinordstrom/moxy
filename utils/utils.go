package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

var ColorRed = "\033[31m"
var ColorGreen = "\033[32m"
var ColorPurple = "\033[35m"
var ColorGray = "\033[37m"
var ColorBold = "\033[1m"
var ColorReset = "\033[0m"

var RightArrow = "  \u2794  "

func GetMockEventString(val map[string]interface{}, withColor bool, payload string) string {
	rawString := fmt.Sprint(val["method"]) +
		" " +
		fmt.Sprint(val["urlpart"]) +
		"  \u2794  [" +
		fmt.Sprint(val["freezetimems"]) +
		" ms] " +
		fmt.Sprint(val["statuscode"]) +
		" " + payload

	if withColor {
		return ColorPurple + rawString + ColorReset
	}

	return rawString
}

func GetJSONObj(filename string) []interface{} {
	var result interface{}
	byteValue := getJSONResultBytes(filename, false)

	err := json.Unmarshal(byteValue, &result)
	if err != nil {
		log.Fatalf("Error occurred during unmarshalling (obj) in file %s: %v", filename, err)
	}

	objArr, ok := result.([]interface{})
	if !ok {
		log.Fatalln("ERROR expected an array of objects from json")
	}

	return objArr
}

func GetJSONObjAsString(absoluteFilePath string) string {
	var result interface{}
	byteValue := getJSONResultBytes(absoluteFilePath, true)
	err := json.Unmarshal(byteValue, &result)
	if err != nil {
		log.Fatalf("Error occurred during unmarshalling (str) in file %s: %v", absoluteFilePath, err)
	}

	jsonStr, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	return string(jsonStr)
}

func getJSONResultBytes(filename string, absolutePath bool) []byte {
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

	byteValue, _ := readJSONFileWithCache(fullJSONPath)

	return byteValue
}
