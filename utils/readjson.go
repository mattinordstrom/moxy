package utils

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type CacheItem struct {
	Data      []byte
	Timestamp time.Time
}

var cache = make(map[string]CacheItem)
var rwMu sync.RWMutex
var cacheDuration = 2 * time.Second

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

	rwMu.Lock()
	cache[filePath] = CacheItem{Data: bytes, Timestamp: time.Now()}
	rwMu.Unlock()

	return bytes, nil
}
