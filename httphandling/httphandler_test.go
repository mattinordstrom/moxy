package httphandling_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mattinordstrom/moxy/httphandling"
	testhelper "github.com/mattinordstrom/moxy/testhelpers"
)

type mockTransport struct{}

var ErrConnectionRefused = errors.New("connection refused")

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// When testing all forwarded requests (proxydef_test.json) will fail
	return nil, fmt.Errorf("mock transport error for %s: %w", req.Host, ErrConnectionRefused)
}

func TestMain(m *testing.M) {
	testhelper.SetupTest(m)
}

func TestProxyResponseFail(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "/api/test123", nil)
	if err != nil {
		t.Fatal(err)
	}

	httphandling.ForwardClient = http.Client{
		Transport: &mockTransport{},
	}

	resRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(httphandling.HTTPHandler)

	handler.ServeHTTP(resRecorder, req)

	actualBody := resRecorder.Body.String()
	expectedBody := "Error: No response from http://localhost:8088/api/test123"

	if actualBody != expectedBody {
		testhelper.PrintAssertError(t, expectedBody, actualBody,
			"Handler returned wrong body")
	}
}

func TestMockedResponse(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, "/api/test/mocktesting", nil)
	if err != nil {
		t.Fatal(err)
	}

	resRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(httphandling.HTTPHandler)

	handler.ServeHTTP(resRecorder, req)

	if status := resRecorder.Code; status != http.StatusNotFound {
		testhelper.PrintAssertError(t, strconv.Itoa(http.StatusNotFound), strconv.Itoa(status),
			"Handler returned wrong status code")
	}

	expectedBody := `{"response":"mocktest"}`
	if resRecorder.Body.String() != expectedBody {
		testhelper.PrintAssertError(t, expectedBody, resRecorder.Body.String(),
			"Handler returned unexpected body")
	}
}
