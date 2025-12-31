// Package fotmob provides a client for the FotMob API.
package fotmob

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/data"
)

const (
	// EmptyCacheFileName is the name of the cache file for empty results.
	EmptyCacheFileName = "empty-results.json"
	// EmptyCacheExpiry is the duration after which empty results expire (7 days).
	EmptyCacheExpiry = 7 * 24 * time.Hour
)

// EmptyResultsCache stores date+league combinations that returned 0 matches.
// This avoids unnecessary API calls for leagues with no matches on specific dates.
type EmptyResultsCache struct {
	mu       sync.RWMutex
	filePath string
	data     EmptyCacheData
}

// EmptyCacheData is the JSON structure stored on disk.
type EmptyCacheData struct {
	Version      int                        `json:"version"`
	EmptyResults map[string]EmptyCacheEntry `json:"empty_results"` // key: "YYYY-MM-DD:leagueID"
}

// EmptyCacheEntry represents a cached empty result with expiration.
type EmptyCacheEntry struct {
	Expires time.Time `json:"expires"`
}

// NewEmptyResultsCache creates a new cache instance.
// It loads existing data from the config directory if available.
// On Linux, uses XDG spec (~/.config/golazo). On other systems, uses ~/.golazo.
func NewEmptyResultsCache() (*EmptyResultsCache, error) {
	configDir, err := data.ConfigDir()
	if err != nil {
		return nil, err
	}

	cache := &EmptyResultsCache{
		filePath: filepath.Join(configDir, EmptyCacheFileName),
		data: EmptyCacheData{
			Version:      1,
			EmptyResults: make(map[string]EmptyCacheEntry),
		},
	}

	// Load existing cache file if it exists
	if err := cache.load(); err != nil {
		// If file doesn't exist or is corrupted, start fresh
		cache.data = EmptyCacheData{
			Version:      1,
			EmptyResults: make(map[string]EmptyCacheEntry),
		}
	}

	// Clean up expired entries on startup
	cache.cleanExpired()

	return cache, nil
}

// IsEmpty checks if a league+date combination is cached as empty.
func (c *EmptyResultsCache) IsEmpty(date string, leagueID int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.makeKey(date, leagueID)
	entry, exists := c.data.EmptyResults[key]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(entry.Expires) {
		return false
	}

	return true
}

// MarkEmpty marks a league+date combination as having no matches.
func (c *EmptyResultsCache) MarkEmpty(date string, leagueID int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.makeKey(date, leagueID)
	c.data.EmptyResults[key] = EmptyCacheEntry{
		Expires: time.Now().Add(EmptyCacheExpiry),
	}
}

// Save persists the cache to disk.
func (c *EmptyResultsCache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.filePath, data, 0644)
}

// load reads the cache from disk.
func (c *EmptyResultsCache) load() error {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &c.data)
}

// cleanExpired removes expired entries from the cache.
func (c *EmptyResultsCache) cleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data.EmptyResults {
		if now.After(entry.Expires) {
			delete(c.data.EmptyResults, key)
		}
	}
}

// makeKey creates a cache key from date and league ID.
func (c *EmptyResultsCache) makeKey(date string, leagueID int) string {
	return date + ":" + itoa(leagueID)
}

// itoa converts an int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var digits []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

// Stats returns statistics about the cache.
func (c *EmptyResultsCache) Stats() (total int, expired int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	for _, entry := range c.data.EmptyResults {
		total++
		if now.After(entry.Expires) {
			expired++
		}
	}
	return
}
