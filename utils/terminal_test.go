package utils_test

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/mattinordstrom/moxy/utils"
)

var errBadInput = errors.New("bad \x1b[2J input")

func TestSanitizeTerminal(t *testing.T) {
	t.Parallel()

	tests := map[string]string{ //nolint:gosec // not credentials
		"plain text, åäö 🙂":         "plain text, åäö 🙂",
		"keeps\nnewlines\tand tabs": "keeps\nnewlines\tand tabs",
		"\x1b[2J\x1b[Hfake prompt":  `\x1b[2J\x1b[Hfake prompt`,
		"title\x1b]0;pwned\x07":     `title\x1b]0;pwned\x07`,
		"c1 csi \u009b31m":          `c1 csi \x9b31m`,
		"carriage\roverwrite\x7f":   `carriage\x0doverwrite\x7f`,
	}

	for in, want := range tests {
		if got := utils.SanitizeTerminal(in); got != want {
			t.Errorf("SanitizeTerminal(%q) = %q, want %q", in, got, want)
		}
	}
}

//nolint:paralleltest // swaps the global logger output
func TestLogErrorHandlesNilAndControlCharacters(t *testing.T) {
	var buf bytes.Buffer

	orig := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(orig) })

	utils.LogError("Invalid proxy target:", nil)
	utils.LogError("Error:", errBadInput)

	out := buf.String()
	if !strings.Contains(out, "Invalid proxy target:") || !strings.Contains(out, `bad \x1b[2J input`) {
		t.Errorf("unexpected log output %q", out)
	}
}
