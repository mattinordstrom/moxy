package utils

import (
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

	testStr := GetMockEventString(testData, true)

	expected := ColorPurple + "GET /api/test/someendpoint  \u2794  [500 ms] 404 {\"response\":\"test123\"}" + ColorReset
	if testStr != expected {
		t.Error("\n EXPECTED: " + expected + "\n ACTUAL: " + testStr)
	}
}
