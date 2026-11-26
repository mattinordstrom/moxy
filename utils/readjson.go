package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CacheItem struct {
	Data      []byte
	Timestamp time.Time
}

var (
	cache         = make(map[string]CacheItem)
	rwMu          sync.RWMutex
	cacheDuration = 1500 * time.Millisecond
)

func readJSONFileWithCache(filePath string) ([]byte, error) {
	rwMu.RLock()

	item, exists := cache[filePath]

	rwMu.RUnlock()

	if exists && time.Since(item.Timestamp) < cacheDuration {
		// fmt.Println("READJSON: Using cached data for", filePath)
		return item.Data, nil
	}

	// fmt.Println("READJSON: Reading file from disk:", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// A file caught mid-edit would otherwise keep failing every request until the cache expires
	if json.Valid(bytes) {
		rwMu.Lock()
		cache[filePath] = CacheItem{Data: bytes, Timestamp: time.Now()}
		rwMu.Unlock()
	}

	return bytes, nil
}

// ForgetCache makes the next read hit the disk, so a save from the admin UI applies right away.
func ForgetCache() {
	rwMu.Lock()
	clear(cache)
	rwMu.Unlock()
}

// WriteFileAtomic replaces the file in one rename, so moxy and editors never read it half written.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	// Write next to the real file so a symlinked def file stays a symlink
	target, err := filepath.EvalSymlinks(path)
	if err != nil {
		target = path
	}

	tmp, err := os.CreateTemp(filepath.Dir(target), "."+filepath.Base(target)+".tmp*")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", path, err)
	}
	defer os.Remove(tmp.Name())

	_, err = tmp.Write(data)
	if err == nil {
		err = tmp.Chmod(perm)
	}

	closeErr := tmp.Close()
	if err == nil {
		err = closeErr
	}

	if err != nil {
		return fmt.Errorf("write temp file for %s: %w", path, err)
	}

	err = os.Rename(tmp.Name(), target)
	if err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}

	ForgetCache()

	return nil
}
