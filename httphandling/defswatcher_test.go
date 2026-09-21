package httphandling_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mattinordstrom/moxy/httphandling"
)

const testFilePermission = 0o600

func TestDefFileFingerprintIsStableWithoutChanges(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "mockdef.json")

	if err := os.WriteFile(path, []byte("[]"), testFilePermission); err != nil {
		t.Fatal(err)
	}

	first := httphandling.DefFileFingerprint(path)
	if first == "" {
		t.Fatal("expected a fingerprint for an existing file")
	}

	if second := httphandling.DefFileFingerprint(path); second != first {
		t.Fatalf("fingerprint changed without a write: %s vs %s", first, second)
	}
}

func TestDefFileFingerprintDetectsSameSizeChange(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "mockdef.json")

	if err := os.WriteFile(path, []byte(`[{"statuscode":200}]`), testFilePermission); err != nil {
		t.Fatal(err)
	}

	before := httphandling.DefFileFingerprint(path)

	time.Sleep(10 * time.Millisecond)

	if err := os.WriteFile(path, []byte(`[{"statuscode":404}]`), testFilePermission); err != nil {
		t.Fatal(err)
	}

	if after := httphandling.DefFileFingerprint(path); after == before {
		t.Fatalf("fingerprint did not change after a same-size rewrite: %s", after)
	}
}

func TestDefFileFingerprintMissingFile(t *testing.T) {
	t.Parallel()

	if fingerprint := httphandling.DefFileFingerprint(filepath.Join(t.TempDir(), "gone.json")); fingerprint != "" {
		t.Fatalf("expected an empty fingerprint for a missing file, got %s", fingerprint)
	}
}
