package httphandling_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/mattinordstrom/moxy/config"
	testhelper "github.com/mattinordstrom/moxy/testhelpers"
)

const (
	payloadFileURL = "/moxyadminui/editpayloadfile"
	adminHost      = "localhost:9097"
	ipv4Host       = "127.0.0.1:9097"
	ipv6Host       = "[::1]:9097"
	headerOrigin   = "Origin"
	headerFetch    = "Sec-Fetch-Site"
	appOrigin      = "http://localhost:3000"
	localRemote    = "127.0.0.1:50000"
)

func setupArchive(t *testing.T) (string, string) {
	t.Helper()

	dir := t.TempDir()
	archive := filepath.Join(dir, "archive")
	sibling := filepath.Join(dir, "archive2")

	for _, d := range []string{archive, sibling} {
		if err := os.Mkdir(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	writeFile(t, filepath.Join(archive, "ok.json"), `{"ok":true}`)
	writeFile(t, filepath.Join(sibling, "secret.json"), `{"secret":true}`)

	if err := os.Symlink(sibling, filepath.Join(archive, "link")); err != nil {
		t.Fatal(err)
	}

	orig := config.AppConfig.Defaults.PayloadArchivePath
	config.AppConfig.Defaults.PayloadArchivePath = archive

	t.Cleanup(func() { config.AppConfig.Defaults.PayloadArchivePath = orig })

	return archive, sibling
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func getPayloadFile(t *testing.T, file string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, payloadFileURL+"?file="+url.QueryEscape(file), nil)
	req.RemoteAddr = localRemote
	req.Host = adminHost

	return testhelper.GetRecorder(req)
}

func postPayloadFile(t *testing.T, file string) *httptest.ResponseRecorder {
	t.Helper()

	return postPayloadFileContent(t, file, `{"written":true}`)
}

func postPayloadFileContent(t *testing.T, file, payload string) *httptest.ResponseRecorder {
	t.Helper()

	body := `{"file":` + strconv.Quote(file) + `,"payload":` + payload + `}`
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, payloadFileURL, strings.NewReader(body))
	req.RemoteAddr = localRemote
	req.Host = adminHost
	req.Header.Set(headerOrigin, "http://"+adminHost)
	req.Header.Set("Content-Type", "application/json")

	return testhelper.GetRecorder(req)
}

//nolint:paralleltest // mutates the global archive path
func TestGetPayloadFileConfinedToArchive(t *testing.T) {
	setupArchive(t)

	if res := getPayloadFile(t, "ok.json"); res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"ok"`) {
		t.Errorf("expected ok.json to be served, got %d %q", res.Code, res.Body.String())
	}

	for _, file := range []string{"../archive2/secret.json", "../../../etc/passwd", "link/secret.json"} {
		res := getPayloadFile(t, file)
		if res.Code == http.StatusOK || strings.Contains(res.Body.String(), "secret") {
			t.Errorf("%s: expected access to be denied, got %d %q", file, res.Code, res.Body.String())
		}
	}
}

//nolint:paralleltest // mutates the global archive path
func TestPostPayloadFileConfinedToArchive(t *testing.T) {
	archive, sibling := setupArchive(t)

	if res := postPayloadFile(t, filepath.Join(archive, "new.json")); res.Code != http.StatusOK {
		t.Errorf("expected write inside archive to succeed, got %d %q", res.Code, res.Body.String())
	}

	if _, err := os.Stat(filepath.Join(archive, "new.json")); err != nil {
		t.Errorf("expected new.json to be written: %v", err)
	}

	for _, file := range []string{filepath.Join(sibling, "evil.json"), filepath.Join(archive, "link", "evil.json")} {
		if res := postPayloadFile(t, file); res.Code == http.StatusOK {
			t.Errorf("%s: expected write to be denied, got %d", file, res.Code)
		}
	}

	if _, err := os.Stat(filepath.Join(sibling, "evil.json")); !os.IsNotExist(err) {
		t.Errorf("expected nothing written outside the archive, stat err: %v", err)
	}
}

//nolint:paralleltest // mutates the global archive path and config
func TestPostPayloadFileKeepsNumbers(t *testing.T) {
	archive, _ := setupArchive(t)

	for _, autoSort := range []bool{false, true} {
		orig := config.AppConfig.Admin.AutoSortJSONKeys
		config.AppConfig.Admin.AutoSortJSONKeys = autoSort

		file := filepath.Join(archive, "numbers.json")
		res := postPayloadFileContent(t, file, `{"b":1.10,"a":12345678901234567890}`)

		config.AppConfig.Admin.AutoSortJSONKeys = orig

		if res.Code != http.StatusOK {
			t.Fatalf("autoSort=%v: expected save to succeed, got %d %q", autoSort, res.Code, res.Body.String())
		}

		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		want := "{\n  \"b\": 1.10,\n  \"a\": 12345678901234567890\n}"
		if autoSort {
			want = "{\n  \"a\": 12345678901234567890,\n  \"b\": 1.10\n}"
		}

		if string(content) != want {
			t.Errorf("autoSort=%v: expected %q, got %q", autoSort, want, content)
		}
	}
}
