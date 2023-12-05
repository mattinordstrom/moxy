package utils_test

import (
	"encoding/json"
	"log"
	"strconv"
	"testing"

	"github.com/mattinordstrom/moxy/models"
	testhelper "github.com/mattinordstrom/moxy/testhelpers"
	"github.com/mattinordstrom/moxy/utils"
)

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
		"GET /api/test/someendpoint  ➔  [500 ms] 404 {\"response\":\"test123\"}" +
		utils.ColorReset

	if testStr != expected {
		testhelper.PrintAssertError(t, expected, testStr)
	}
}

func TestGetProxyEventString(t *testing.T) {
	t.Parallel()

	testStr := utils.GetProxyEventString("/api/someendpoint", "http://localhost:8088/", "", false)

	expected := utils.ColorGreen +
		"/api/someendpoint  ➔  http://localhost:8088/api/someendpoint" +
		utils.ColorReset

	if testStr != expected {
		testhelper.PrintAssertError(t, expected, testStr)
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
		testhelper.PrintAssertError(t, strconv.FormatBool(expected), strconv.FormatBool(res))
	}

	// Test with both
	testData.Payload = map[string]interface{}{
		"response": "test123",
	}
	res = utils.UsePayloadFromFile(testData)
	expected = false

	if res != expected {
		testhelper.PrintAssertError(t, strconv.FormatBool(expected), strconv.FormatBool(res))
	}
}
