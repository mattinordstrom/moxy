package utils_test

import (
	"encoding/json"
	"log"
	"strconv"
	"testing"

	"github.com/mattinordstrom/moxy/models"
	"github.com/mattinordstrom/moxy/utils"
)

func printAssertError(t *testing.T, expected string, actual string) {
	t.Helper()
	t.Error("\n EXPECTED: " + expected + "\n ACTUAL: " + actual)
}

func TestGetMockEventString(t *testing.T) {
	t.Parallel()

	var testData models.Mock
	testData.Method = "GET"
	testData.URLPart = "/api/test/someendpoint"
	testData.StatusCode = 404
	testData.Payload = map[string]interface{}{
		"response": "test123",
	}
	testData.FreezeTimeMS = 500

	payloadM, err := json.Marshal(testData.Payload)
	if err != nil {
		log.Fatalf("Error occurred during marshalling in test: %v", err)
	}

	testStr := utils.GetMockEventString(testData, true, string(payloadM))

	expected := utils.ColorPurple +
		"GET /api/test/someendpoint  \u2794  [500 ms] 404 {\"response\":\"test123\"}" +
		utils.ColorReset

	if testStr != expected {
		printAssertError(t, expected, testStr)
	}
}

func TestUsePayloadFromFile(t *testing.T) {
	t.Parallel()

	var testData models.Mock
	testData.Method = "GET"
	testData.URLPart = "/api/test/someendpoint"
	testData.StatusCode = 200
	testData.FreezeTimeMS = 0

	// Test with only payloadFromFile
	testData.PayloadFromFile = "/home/test/archive/somefile.json"
	res := utils.UsePayloadFromFile(testData)
	expected := true

	if res != expected {
		printAssertError(t, strconv.FormatBool(expected), strconv.FormatBool(res))
	}

	// Test with both
	testData.Payload = map[string]interface{}{
		"response": "test123",
	}
	res = utils.UsePayloadFromFile(testData)
	expected = false

	if res != expected {
		printAssertError(t, strconv.FormatBool(expected), strconv.FormatBool(res))
	}
}
