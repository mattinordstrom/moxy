package utils

import (
	"encoding/json"
	"log"
	"testing"
)

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
		t.Error("\n EXPECTED: " + expected + "\n ACTUAL: " + testStr)
	}
}
