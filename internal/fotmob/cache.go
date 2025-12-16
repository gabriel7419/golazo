package fotmob

import (
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// CacheConfig holds configuration for API response caching.
type CacheConfig struct {
	MatchesTTL       time.Duration // How long to cache match list results
	MatchDetailsTTL  time.Duration // How long to cache match details
	MaxMatchesCache  int           // Maximum number of date entries to cache
	MaxDetailsCache  int           // Maximum number of match details to cache
}

// DefaultCacheConfig returns sensible defaults for caching.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MatchesTTL:       5 * time.Minute,  // Matches can change (scores update)
		MatchDetailsTTL:  2 * time.Minute,  // Details for live matches need fresher data
		MaxMatchesCache:  10,               // Cache up to 10 date queries
		MaxDetailsCache:  100,              // Cache up to 100 match details
	}
}

// cachedMatches holds cached match data with expiration.
type cachedMatches struct {
	matches   []api.Match
	expiresAt time.Time
}

// cachedDetails holds cached match details with expiration.
type cachedDetails struct {
	details   *api.MatchDetails
	expiresAt time.Time
}

// ResponseCache provides thread-safe caching for API responses.
type ResponseCache struct {
	config         CacheConfig
	matchesMu      sync.RWMutex
	matchesCache   map[string]cachedMatches // key: "YYYY-MM-DD"
	detailsMu      sync.RWMutex
	detailsCache   map[int]cachedDetails    // key: matchID
}

// NewResponseCache creates a new cache with the given configuration.
func NewResponseCache(config CacheConfig) *ResponseCache {
	return &ResponseCache{
		config:       config,
		matchesCache: make(map[string]cachedMatches),
		detailsCache: make(map[int]cachedDetails),
	}
}

// GetMatches retrieves cached matches for a date, returns nil if not cached or expired.
func (c *ResponseCache) GetMatches(dateKey string) []api.Match {
	c.matchesMu.RLock()
	defer c.matchesMu.RUnlock()

	cached, ok := c.matchesCache[dateKey]
	if !ok || time.Now().After(cached.expiresAt) {
		return nil
	}
	return cached.matches
}

// SetMatches stores matches in cache with TTL.
func (c *ResponseCache) SetMatches(dateKey string, matches []api.Match) {
	c.matchesMu.Lock()
	defer c.matchesMu.Unlock()

	// Evict oldest entries if cache is full
	if len(c.matchesCache) >= c.config.MaxMatchesCache {
		c.evictOldestMatches()
	}

	c.matchesCache[dateKey] = cachedMatches{
		matches:   matches,
		expiresAt: time.Now().Add(c.config.MatchesTTL),
	}
}

// GetDetails retrieves cached match details, returns nil if not cached or expired.
func (c *ResponseCache) GetDetails(matchID int) *api.MatchDetails {
	c.detailsMu.RLock()
	defer c.detailsMu.RUnlock()

	cached, ok := c.detailsCache[matchID]
	if !ok || time.Now().After(cached.expiresAt) {
		return nil
	}
	return cached.details
}

// SetDetails stores match details in cache with TTL.
// For finished matches, uses a longer TTL since the data won't change.
func (c *ResponseCache) SetDetails(matchID int, details *api.MatchDetails) {
	c.detailsMu.Lock()
	defer c.detailsMu.Unlock()

	// Evict oldest entries if cache is full
	if len(c.detailsCache) >= c.config.MaxDetailsCache {
		c.evictOldestDetails()
	}

	ttl := c.config.MatchDetailsTTL
	// Use longer TTL for finished matches since they won't change
	if details != nil && details.Status == api.MatchStatusFinished {
		ttl = 30 * time.Minute
	}

	c.detailsCache[matchID] = cachedDetails{
		details:   details,
		expiresAt: time.Now().Add(ttl),
	}
}

// GetCachedMatchIDs returns all match IDs currently in the details cache.
func (c *ResponseCache) GetCachedMatchIDs() []int {
	c.detailsMu.RLock()
	defer c.detailsMu.RUnlock()

	ids := make([]int, 0, len(c.detailsCache))
	for id := range c.detailsCache {
		ids = append(ids, id)
	}
	return ids
}

// ClearDetailsCache clears all cached match details.
func (c *ResponseCache) ClearDetailsCache() {
	c.detailsMu.Lock()
	defer c.detailsMu.Unlock()
	c.detailsCache = make(map[int]cachedDetails)
}

// evictOldestMatches removes expired or oldest entries (must hold write lock).
func (c *ResponseCache) evictOldestMatches() {
	now := time.Now()
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, cached := range c.matchesCache {
		// Remove expired entries
		if now.After(cached.expiresAt) {
			delete(c.matchesCache, key)
			continue
		}
		// Track oldest non-expired entry
		if first || cached.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.expiresAt
			first = false
		}
	}

	// If still at capacity after removing expired, remove oldest
	if len(c.matchesCache) >= c.config.MaxMatchesCache && oldestKey != "" {
		delete(c.matchesCache, oldestKey)
	}
}

// evictOldestDetails removes expired or oldest entries (must hold write lock).
func (c *ResponseCache) evictOldestDetails() {
	now := time.Now()
	var oldestKey int
	var oldestTime time.Time
	first := true

	for key, cached := range c.detailsCache {
		// Remove expired entries
		if now.After(cached.expiresAt) {
			delete(c.detailsCache, key)
			continue
		}
		// Track oldest non-expired entry
		if first || cached.expiresAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.expiresAt
			first = false
		}
	}

	// If still at capacity after removing expired, remove oldest
	if len(c.detailsCache) >= c.config.MaxDetailsCache && oldestKey != 0 {
		delete(c.detailsCache, oldestKey)
	}
}

