package utils

import (
	"encoding/json"
	"log"
	"strconv"
	"testing"
)

func printAssertError(t *testing.T, expected string, actual string) {
	t.Helper()
	t.Error("\n EXPECTED: " + expected + "\n ACTUAL: " + actual)
}

func TestGetMockEventString(t *testing.T) {
	testData := make(map[string]interface{})
	testData["method"] = "GET"
	testData["urlpart"] = "/api/test/someendpoint"
	testData["statuscode"] = 404
	testData["payload"] = map[string]interface{}{
		"response": "test123",
	}
	testData["freezetimems"] = 500

	payloadM, err := json.Marshal(testData["payload"])
	if err != nil {
		log.Fatalf("Error occurred during marshalling in test: %v", err)
	}

	testStr := GetMockEventString(testData, true, string(payloadM))

	expected := ColorPurple + "GET /api/test/someendpoint  \u2794  [500 ms] 404 {\"response\":\"test123\"}" + ColorReset
	if testStr != expected {
		printAssertError(t, expected, testStr)
	}
}

func TestUsePayloadFromFile(t *testing.T) {
	testData := make(map[string]interface{})
	testData["method"] = "GET"
	testData["urlpart"] = "/api/test/someendpoint"
	testData["statuscode"] = 200
	testData["freezetimems"] = 0

	// Test with only payloadFromFile
	testData["payloadFromFile"] = "/home/test/archive/somefile.json"
	res := UsePayloadFromFile(testData)
	expected := true
	if res != expected {
		printAssertError(t, strconv.FormatBool(expected), strconv.FormatBool(res))
	}

	// Test with both
	testData["payload"] = map[string]interface{}{
		"response": "test123",
	}
	res = UsePayloadFromFile(testData)
	expected = false
	if res != expected {
		printAssertError(t, strconv.FormatBool(expected), strconv.FormatBool(res))
	}
}
