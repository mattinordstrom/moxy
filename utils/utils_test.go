package utils_test

import (
	"encoding/json"
	"log"
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
	testData.Payload = json.RawMessage(`{"response":"test123"}`)
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
		testhelper.PrintAssertError(t, expected, testStr, "")
	}
}

func TestGetProxyEventString(t *testing.T) {
	t.Parallel()

	testStr := utils.GetProxyEventString("/api/someendpoint", "http://localhost:8088/", "", false)

	expected := utils.ColorGreen +
		"/api/someendpoint  ➔  http://localhost:8088/api/someendpoint" +
		utils.ColorReset

	if testStr != expected {
		testhelper.PrintAssertError(t, expected, testStr, "")
	}
}

func TestUsePayloadFromFile(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mock string
		want bool
	}{
		"only payloadFromFile":             {`{"payloadFromFile":"/a.json"}`, true},
		"empty payload":                    {`{"payload":"","payloadFromFile":"/a.json"}`, true},
		"null payload without toggle":      {`{"payload":null,"payloadFromFile":"/a.json"}`, true},
		"both without toggle":              {`{"payload":{"a":1},"payloadFromFile":"/a.json"}`, false},
		"no file":                          {`{"payload":null}`, false},
		"toggle on":                        {`{"payload":{"a":1},"payloadFromFile":"/a.json","usePayloadFromFile":true}`, true},
		"toggle off serves a null payload": {`{"payload":null,"payloadFromFile":"/a.json","usePayloadFromFile":false}`, false},
	}

	for name, test := range tests {
		var mock models.Mock
		if err := json.Unmarshal([]byte(test.mock), &mock); err != nil {
			t.Fatal(err)
		}

		if got := utils.UsePayloadFromFile(mock); got != test.want {
			t.Errorf("%s: got %v, want %v", name, got, test.want)
		}
	}
}
