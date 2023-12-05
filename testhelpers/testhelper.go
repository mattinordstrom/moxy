package testhelper

import (
	"bytes"
	"os"
	"testing"
)

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

func PrintAssertError(t *testing.T, expected string, actual string) {
	t.Helper()
	t.Error("\n EXPECTED: " + expected + "\n ACTUAL: " + actual)
}
