package httphandling

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattinordstrom/moxy/utils"
)

func TestMockedResponse(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "/api/test/mocktesting", nil)
	if err != nil {
		t.Fatal(err)
	}

	utils.MockFile = "mockdef_test.json"

	resRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(httpHandler)

	handler.ServeHTTP(resRecorder, req)

	if status := resRecorder.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}

	expected := `{"response":"mocktest"}`
	if resRecorder.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", resRecorder.Body.String(), expected)
	}
}

func TestCreateReqFromReq(t *testing.T) {
	t.Parallel()

	expectedReqBody := `{ "id": "123" }`
	expectedReqHeaderUA := "PostmanRuntime/7.30.0"

	req, err := http.NewRequest(http.MethodGet, "/api/test/mocktesting", bytes.NewReader([]byte(expectedReqBody)))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("User-Agent", expectedReqHeaderUA)
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	newURL := "http://localhost:9098/api/test/mocktesting"
	freq, _ := createReqFromReq(req, newURL)

	// Test Accept-Encoding header
	if acceptEncoding := freq.Header.Get("Accept-Encoding"); acceptEncoding != "" {
		t.Errorf("handler returned wrong header Accept-Encoding: got %v want EMPTY", acceptEncoding)
	}

	// Test user-agent header
	if ua := freq.Header.Get("User-Agent"); ua != expectedReqHeaderUA {
		t.Errorf("handler returned wrong header User-Agent: got %v want %v", ua, expectedReqHeaderUA)
	}

	// Test body
	reqBody, ioerr := io.ReadAll(freq.Body)
	if ioerr != nil {
		t.Fatal(ioerr)
	}
	if reqBodyString := string(reqBody); reqBodyString != expectedReqBody {
		t.Errorf("handler returned wrong body: got %v want %v", reqBodyString, expectedReqBody)
	}
}
