package testhelper

import (
	"bytes"
	"os"
	"testing"

	"github.com/mattinordstrom/moxy/utils"
)

func SetupTest(m *testing.M) {
	utils.MockFile = "mockdef_test.json"
	utils.ProxyFile = "proxydef_test.json"

	// Run the tests
	code := m.Run()

	os.Exit(code)
}

func CaptureOutput(t *testing.T, funct func()) string {
	t.Helper()

	// Save original stdout
	origStdout := os.Stdout

	// Create a pipe to capture output
	read, write, _ := os.Pipe()
	os.Stdout = write

	// Execute the function with arguments
	funct()

	// Close the pipe and restore stdout
	write.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(read); err != nil {
		t.Fatalf("Failed to read from buffer: %v", err)
	}

	return buf.String()
}

func PrintAssertError(t *testing.T, expected string, actual string, extraInfo string) {
	t.Helper()

	t.Error(utils.ColorRed + "\n" + extraInfo + "\n EXPECTED: " + expected + "\n ACTUAL: " + actual + utils.ColorReset)
}
