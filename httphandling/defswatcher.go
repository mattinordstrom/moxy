package httphandling

import (
	"os"
	"strconv"
	"time"

	"github.com/mattinordstrom/moxy/utils"
)

var DefsWatchInterval = 1000 * time.Millisecond

func watchDefFiles() {
	watched := []string{utils.MockFile, utils.ProxyFile}

	known := make(map[string]string, len(watched))
	for _, file := range watched {
		known[file] = DefFileFingerprint(file)
	}

	ticker := time.NewTicker(DefsWatchInterval)
	defer ticker.Stop()

	for range ticker.C {
		for _, file := range watched {
			fingerprint := DefFileFingerprint(file)
			if fingerprint == known[file] {
				continue
			}

			known[file] = fingerprint

			recordEvent(LogEntry{Type: utils.EventTypeDefsChanged, Message: file + " changed"})
			updateAdminWithLatest(file+" changed", utils.EventTypeDefsChanged,
				map[string]any{"file": file})
		}
	}
}

func DefFileFingerprint(path string) string {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return ""
	}

	return strconv.FormatInt(fileInfo.ModTime().UnixNano(), 10) +
		":" + strconv.FormatInt(fileInfo.Size(), 10)
}
