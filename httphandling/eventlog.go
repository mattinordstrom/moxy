package httphandling

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattinordstrom/moxy/utils"
)

// LogEntry never holds headers, bodies, payloads or query values, so the log is safe to keep on disk and hand to Claude.
type LogEntry struct {
	Seq        int64  `json:"seq"`
	Time       string `json:"time"`
	Type       string `json:"type"`
	Method     string `json:"method,omitempty"`
	URL        string `json:"url,omitempty"`
	Rule       string `json:"rule,omitempty"`
	Target     string `json:"target,omitempty"`
	Status     int    `json:"status,omitempty"`
	DurationMS *int64 `json:"durationMs,omitempty"`
	Message    string `json:"message,omitempty"`
}

const (
	defaultRouteRule = "(default route)"
	eventLogFilePerm = 0o600
)

var (
	EventLogSize            = 1000
	EventLogFileLimit int64 = 5 << 20

	eventLogMu      sync.Mutex
	eventLog        []LogEntry
	eventLogSeq     int64
	eventLogFile    *os.File
	eventLogPath    string
	eventLogWritten int64
)

func OpenEventLogFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, eventLogFilePerm)
	if err != nil {
		return fmt.Errorf("open event log: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()

		return fmt.Errorf("stat event log: %w", err)
	}

	eventLogMu.Lock()
	defer eventLogMu.Unlock()

	eventLogFile = file
	eventLogPath = path
	eventLogWritten = info.Size()

	return nil
}

func recordEvent(entry LogEntry) {
	eventLogMu.Lock()
	defer eventLogMu.Unlock()

	eventLogSeq++
	entry.Seq = eventLogSeq
	entry.Time = time.Now().Format("2006-01-02T15:04:05.000")

	eventLog = append(eventLog, entry)
	if len(eventLog) > EventLogSize {
		eventLog = append(eventLog[:0:0], eventLog[len(eventLog)-EventLogSize:]...)
	}

	writeEventLogLine(entry)
}

func writeEventLogLine(entry LogEntry) {
	if eventLogFile == nil {
		return
	}

	line, err := json.Marshal(entry)
	if err != nil {
		return
	}

	if eventLogWritten+int64(len(line))+1 > EventLogFileLimit {
		rotateEventLogFile()
	}

	written, err := eventLogFile.Write(append(line, '\n'))
	eventLogWritten += int64(written)

	if err != nil {
		utils.LogError("Writing "+eventLogPath+" failed, stopped logging to it:", err)
		eventLogFile.Close()
		eventLogFile = nil
	}
}

func rotateEventLogFile() {
	eventLogFile.Close()
	eventLogFile = nil

	err := os.Rename(eventLogPath, eventLogPath+".1")
	if err == nil {
		var file *os.File

		file, err = os.OpenFile(eventLogPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, eventLogFilePerm)
		if err == nil {
			eventLogFile = file
			eventLogWritten = 0
		}
	}

	if err != nil {
		utils.LogError("Rotating "+eventLogPath+" failed, stopped logging to it:", err)
	}
}

func eventsSince(since int64) []LogEntry {
	eventLogMu.Lock()
	defer eventLogMu.Unlock()

	entries := []LogEntry{}

	for _, entry := range eventLog {
		if entry.Seq > since {
			entries = append(entries, entry)
		}
	}

	return entries
}

func handleEventLogReq(resWriter http.ResponseWriter, req *http.Request) {
	since, err := strconv.ParseInt(req.URL.Query().Get("since"), 10, 64)
	if err != nil {
		since = 0
	}

	resWriter.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(resWriter).Encode(eventsSince(since))
	if err != nil {
		utils.LogError("Error writing response: ", err)
	}
}

func millisSince(start time.Time) *int64 {
	ms := time.Since(start).Milliseconds()

	return &ms
}

// Query values often carry tokens or personal data, so only the parameter names are kept.
func redactURL(rawURL string) string {
	path, query, hasQuery := strings.Cut(rawURL, "?")

	if parsed, err := url.Parse(path); err == nil && parsed.User != nil {
		path = parsed.Redacted()
	}

	if !hasQuery {
		return path
	}

	params := strings.Split(query, "&")
	for i, param := range params {
		if name, _, hasValue := strings.Cut(param, "="); hasValue {
			params[i] = name + "=***"
		}
	}

	return path + "?" + strings.Join(params, "&")
}
