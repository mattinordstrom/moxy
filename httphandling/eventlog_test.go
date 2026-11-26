package httphandling_test

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattinordstrom/moxy/httphandling"
)

func TestRedactURL(t *testing.T) {
	t.Parallel()

	tests := map[string]string{ //nolint:gosec // fake credentials to check they are redacted
		"/api/users":                              "/api/users",
		"/api/users?token=abc&page=2":             "/api/users?token=***&page=***",
		"/api/search?q=a=b&flag":                  "/api/search?q=***&flag",
		"http://host:8080/api?apikey=secret":      "http://host:8080/api?apikey=***",
		"https://user:pass@host/api?email=a@b.se": "https://user:xxxxx@host/api?email=***",
	}

	for in, want := range tests {
		if got := httphandling.RedactURL(in); got != want {
			t.Errorf("RedactURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func readEventLog(t *testing.T, since string) []httphandling.LogEntry {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/moxyadminui/log?since="+since, nil)
	req.Host = adminHost
	req.RemoteAddr = localRemote

	res := httptest.NewRecorder()
	httphandling.RootHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("log endpoint returned %d: %s", res.Code, res.Body.String())
	}

	var entries []httphandling.LogEntry
	if err := json.Unmarshal(res.Body.Bytes(), &entries); err != nil {
		t.Fatal(err)
	}

	return entries
}

func serveRequest(t *testing.T, method string, target string) {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), method, target, strings.NewReader(`{"secret":"body"}`))
	req.Header.Set("Authorization", "Bearer secret-token")
	httphandling.RootHandler(httptest.NewRecorder(), req)
}

//nolint:paralleltest // uses the global event log
func TestEventLogRecordsMockWithoutPayload(t *testing.T) {
	httphandling.ResetEventLog(t)

	serveRequest(t, http.MethodGet, "/api/test/mocktesting?token=abc")

	entries := entriesOfType(readEventLog(t, "0"), "mock")
	if len(entries) != 1 {
		t.Fatalf("expected 1 mock entry, got %+v", entries)
	}

	entry := entries[0]
	if entry.Type != "mock" || entry.Method != http.MethodGet || entry.URL != "/api/test/mocktesting?token=***" ||
		entry.Rule != "/api/test/mocktesting" || entry.Status != http.StatusNotFound || entry.DurationMS == nil {
		t.Errorf("unexpected entry %+v", entry)
	}

	assertNoSecrets(t, entries)
}

//nolint:paralleltest // uses the global event log and default route
func TestEventLogRecordsProxyStatusAndFailures(t *testing.T) {
	httphandling.ResetEventLog(t)
	startUpstream(t)

	serveRequest(t, http.MethodPost, "/upstream/echo?session=abc")

	httphandling.DefaultRoute = "http://127.0.0.1:1"
	serveRequest(t, http.MethodGet, "/upstream/down")

	entries := readEventLog(t, "0")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %+v", entries)
	}

	proxied := entries[0]
	if proxied.Type != "proxy" || proxied.Method != http.MethodPost || proxied.URL != "/upstream/echo?session=***" ||
		proxied.Rule != "(default route)" || !strings.HasSuffix(proxied.Target, "/upstream/echo?session=***") ||
		proxied.Status != http.StatusOK || proxied.DurationMS == nil {
		t.Errorf("unexpected proxy entry %+v", proxied)
	}

	failed := entries[1]
	if failed.Type != "error" || failed.Status != http.StatusBadGateway || !strings.HasPrefix(failed.Message, "No response") {
		t.Errorf("unexpected failure entry %+v", failed)
	}

	assertNoSecrets(t, entries)

	if got := readEventLog(t, "1"); len(got) != 1 || got[0].Seq != 2 {
		t.Errorf("since=1 returned %+v", got)
	}
}

//nolint:paralleltest // uses the global event log
func TestEventLogKeepsOnlyTheLatestEntries(t *testing.T) {
	httphandling.ResetEventLog(t)

	orig := httphandling.EventLogSize
	httphandling.EventLogSize = 3

	t.Cleanup(func() { httphandling.EventLogSize = orig })

	for range 5 {
		serveRequest(t, http.MethodGet, "/api/test/mocktesting")
	}

	entries := readEventLog(t, "0")
	if len(entries) != 3 || entries[0].Seq != 3 || entries[2].Seq != 5 {
		t.Errorf("expected entries 3-5, got %+v", entries)
	}
}

//nolint:paralleltest // uses the global event log
func TestEventLogFileRotates(t *testing.T) {
	httphandling.ResetEventLog(t)

	path := filepath.Join(t.TempDir(), "moxy.log")
	if err := httphandling.OpenEventLogFile(path); err != nil {
		t.Fatal(err)
	}

	orig := httphandling.EventLogFileLimit
	httphandling.EventLogFileLimit = 400

	t.Cleanup(func() { httphandling.EventLogFileLimit = orig })

	for range 6 {
		serveRequest(t, http.MethodGet, "/api/test/mocktesting?token=abc")
	}

	current := readLogFile(t, path)
	rotated := readLogFile(t, path+".1")

	if len(current) == 0 || len(rotated) == 0 || len(current)+len(rotated) > 6 {
		t.Fatalf("expected entries split over both files, got %d and %d", len(current), len(rotated))
	}

	if last := current[len(current)-1]; last.Seq != 6 || last.URL != "/api/test/mocktesting?token=***" {
		t.Errorf("unexpected last entry %+v", last)
	}

	for _, file := range []string{path, path + ".1"} {
		info, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}

		if info.Size() > httphandling.EventLogFileLimit {
			t.Errorf("%s is %d bytes, over the limit", file, info.Size())
		}

		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("%s has permissions %o, want 600", file, perm)
		}
	}
}

func entriesOfType(entries []httphandling.LogEntry, entryType string) []httphandling.LogEntry {
	var matching []httphandling.LogEntry

	for _, entry := range entries {
		if entry.Type == entryType {
			matching = append(matching, entry)
		}
	}

	return matching
}

func readLogFile(t *testing.T, path string) []httphandling.LogEntry {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var entries []httphandling.LogEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry httphandling.LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			t.Fatalf("bad line %q: %v", scanner.Text(), err)
		}

		entries = append(entries, entry)
	}

	return entries
}

func assertNoSecrets(t *testing.T, entries []httphandling.LogEntry) {
	t.Helper()

	raw, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}

	for _, secret := range []string{"abc", "secret", "Bearer", "body", "statuscode"} {
		if strings.Contains(string(raw), secret) {
			t.Errorf("log contains %q: %s", secret, raw)
		}
	}
}
