package httphandling_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	testhelper "github.com/mattinordstrom/moxy/testhelpers"
	"github.com/mattinordstrom/moxy/utils"
)

func getMock(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)

	return testhelper.GetRecorder(req)
}

func TestInvalidRegexOnlyDisablesItsOwnMock(t *testing.T) {
	t.Parallel()

	res := getMock(t, "/api/test/mocktesting")
	if res.Code != http.StatusNotFound || res.Body.String() != `{"response":"mocktest"}` {
		t.Errorf("expected the valid mock to still be served, got %d %q", res.Code, res.Body.String())
	}
}

func TestMockWithoutStatusCodeReturns200(t *testing.T) {
	t.Parallel()

	res := getMock(t, "/api/edge/nostatus")
	if res.Code != http.StatusOK || res.Body.String() != `{"response":"edge case"}` {
		t.Errorf("expected 200 with the payload, got %d %q", res.Code, res.Body.String())
	}
}

func TestMockWithInvalidStatusCodeReturns500(t *testing.T) {
	t.Parallel()

	res := getMock(t, "/api/edge/badstatus")
	if res.Code != http.StatusInternalServerError || !strings.Contains(res.Body.String(), "invalid statuscode 42") {
		t.Errorf("expected 500 naming the statuscode, got %d %q", res.Code, res.Body.String())
	}
}

func TestMissingPayloadFromFileReturns500(t *testing.T) {
	t.Parallel()

	res := getMock(t, "/api/edge/missingfile")
	if res.Code != http.StatusInternalServerError || !strings.Contains(res.Body.String(), "payloadFromFile") {
		t.Errorf("expected 500 naming payloadFromFile, got %d %q", res.Code, res.Body.String())
	}
}

//nolint:paralleltest // swaps the global mock file
func TestAdminWorksWithBrokenMockdef(t *testing.T) {
	setupArchive(t)

	broken := filepath.Join(t.TempDir(), "broken.json")
	writeFile(t, broken, `[{"active": tru`)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	rel, err := filepath.Rel(filepath.Dir(wd), broken)
	if err != nil {
		t.Fatal(err)
	}

	orig := utils.MockFile
	utils.MockFile = rel

	t.Cleanup(func() { utils.MockFile = orig })

	if res := getMock(t, "/api/test/mocktesting"); res.Code != http.StatusInternalServerError {
		t.Errorf("expected mock traffic to fail on a broken mockdef, got %d", res.Code)
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/moxyadminui/settings", nil)
	req.Host = adminHost

	if res := testhelper.GetRecorder(req); res.Code != http.StatusOK {
		t.Errorf("expected the admin API to keep working, got %d %q", res.Code, res.Body.String())
	}
}
