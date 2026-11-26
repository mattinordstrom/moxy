package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mattinordstrom/moxy/utils"
)

func TestInvalidJSONIsNotCached(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "payload.json")

	if err := os.WriteFile(path, []byte(`{"half":`), 0o600); err != nil {
		t.Fatal(err)
	}

	if _, err := utils.GetJSONPayloadFromAbsolutePath(path); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(`{"whole":true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := utils.GetJSONPayloadFromAbsolutePath(path)
	if err != nil || string(got) != `{"whole":true}` {
		t.Errorf("expected the fixed file to be read right away, got %q %v", got, err)
	}
}

func TestWriteFileAtomicKeepsSymlinks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "real.json")
	link := filepath.Join(dir, "link.json")

	if err := os.WriteFile(target, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if err := utils.WriteFileAtomic(link, []byte(`[1]`), 0o600); err != nil {
		t.Fatal(err)
	}

	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("expected link.json to still be a symlink, err %v", err)
	}

	if got, _ := os.ReadFile(target); string(got) != `[1]` {
		t.Errorf("expected the target to be written, got %q", got)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("expected no temp files left behind, got %d entries", len(entries))
	}
}
