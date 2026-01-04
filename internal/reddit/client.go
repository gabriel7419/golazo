package reddit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DebugLogger is a function type for debug logging
type DebugLogger func(message string)

// Fetcher defines the interface for fetching data from Reddit.
// This allows for easy swapping between public JSON API and OAuth API.
type Fetcher interface {
	Search(query string, limit int, matchTime time.Time) ([]SearchResult, error)
}

// PublicJSONFetcher uses Reddit's public JSON endpoints (no auth required).
// Note: Has stricter rate limits than OAuth API.
type PublicJSONFetcher struct {
	httpClient  *http.Client
	userAgent   string
	rateLimiter *rateLimiter
}

// rateLimiter implements adaptive rate limiting for Reddit API.
// Increases delays when CAPTCHA errors are detected.
type rateLimiter struct {
	mu              sync.Mutex
	lastRequest     time.Time
	minInterval     time.Duration
	captchaCount    int
	lastCaptchaTime time.Time
	userAgentIndex  int // Track which user agent to use next
}

func newRateLimiter(requestsPerMinute int) *rateLimiter {
	interval := time.Minute / time.Duration(requestsPerMinute)
	return &rateLimiter{
		minInterval: interval,
	}
}

func (r *rateLimiter) wait() {
	r.mu.Lock()
	defer r.mu.Unlock()

	elapsed := time.Since(r.lastRequest)

	// If we've had CAPTCHA errors recently, be more conservative
	currentInterval := r.minInterval
	if r.captchaCount > 0 && time.Since(r.lastCaptchaTime) < 10*time.Minute {
		// Double the interval after CAPTCHA detections
		currentInterval = r.minInterval * 2
	}

	if elapsed < currentInterval {
		time.Sleep(currentInterval - elapsed)
	}
	r.lastRequest = time.Now()
}

// recordCaptchaError increases the rate limiting after CAPTCHA detection
func (r *rateLimiter) recordCaptchaError() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.captchaCount++
	r.lastCaptchaTime = time.Now()
}

// getNextUserAgent returns the next user agent in rotation
func (r *rateLimiter) getNextUserAgent() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	agent := userAgents[r.userAgentIndex]
	r.userAgentIndex = (r.userAgentIndex + 1) % len(userAgents)
	return agent
}

// User agents to rotate through to reduce bot detection
var userAgents = []string{
	"golazo:v1.0.0 (by /u/golazo_app)", // Keep original for backwards compatibility
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
}

// NewPublicJSONFetcher creates a new fetcher using public Reddit JSON API.
func NewPublicJSONFetcher() *PublicJSONFetcher {
	return &PublicJSONFetcher{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		// Reddit requires a descriptive User-Agent
		userAgent:   userAgents[0],     // Start with first agent
		rateLimiter: newRateLimiter(5), // Reduced to 5 requests per minute to reduce CAPTCHA blocking
	}
}

// Search performs a search on r/soccer for Media posts matching the query.
// matchTime is used to filter results to posts created around the match date.
func (f *PublicJSONFetcher) Search(query string, limit int, matchTime time.Time) ([]SearchResult, error) {
	f.rateLimiter.wait()

	// Build timestamp range for filtering (match day -1 to +2 days)
	// Goals are usually posted within hours of happening, but we add buffer
	startTime := matchTime.Add(-24 * time.Hour).Unix()
	endTime := matchTime.Add(48 * time.Hour).Unix()

	// Build search URL for r/soccer with Media flair filter and timestamp
	// Reddit CloudSearch supports timestamp:START..END syntax
	searchURL := fmt.Sprintf(
		"https://www.reddit.com/r/soccer/search.json?q=%s+flair:Media+timestamp:%d..%d&restrict_sr=on&sort=relevance&limit=%d",
		url.QueryEscape(query),
		startTime,
		endTime,
		limit,
	)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Use rotating user agents to reduce bot detection
	req.Header.Set("User-Agent", f.rateLimiter.getNextUserAgent())

	// Add realistic browser headers to further reduce CAPTCHA blocking
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,*;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch from reddit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("reddit API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	// Check if Reddit is serving a CAPTCHA or bot detection page
	bodyStr := string(body)
	if isCaptchaResponse(bodyStr) {
		f.rateLimiter.recordCaptchaError()
		return nil, fmt.Errorf("reddit is blocking requests (CAPTCHA/bot detection)")
	}

	var searchResp redditSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		// Check if the response is HTML (likely a CAPTCHA or error page)
		if strings.Contains(bodyStr, "<html") || strings.Contains(bodyStr, "<!DOCTYPE html") {
			return nil, fmt.Errorf("reddit returned HTML instead of JSON (likely CAPTCHA or rate limit)")
		}
		return nil, fmt.Errorf("parse response: %w", err)
	}

	results := make([]SearchResult, 0, len(searchResp.Data.Children))
	for _, child := range searchResp.Data.Children {
		result := child.Data.toSearchResult()
		// Only include posts with Media flair
		if result.Flair == "Media" {
			results = append(results, result)
		}
	}

	return results, nil
}

// Client provides goal replay link fetching from Reddit r/soccer.
// Supports both OAuth API (preferred) and public JSON API (fallback).
type Client struct {
	oauthFetcher  *OAuthClient // Preferred: higher rate limits, no CAPTCHAs
	publicFetcher Fetcher      // Fallback: public API with limitations
	cache         *GoalLinkCache
	debugLogger   DebugLogger  // Optional debug logger function
}

// debugLog is a helper method to safely call the debug logger if it exists
func (c *Client) debugLog(message string) {
	if c.debugLogger != nil {
		c.debugLogger(message)
	}
}

// NewClient creates a new Reddit client with OAuth (preferred) or public API (fallback).
func NewClient() (*Client, error) {
	cache, err := NewGoalLinkCache()
	if err != nil {
		return nil, fmt.Errorf("create cache: %w", err)
	}

	// Try OAuth first (higher rate limits, no CAPTCHAs)
	oauthClient, err := NewOAuthClient()
	if err != nil {
		return nil, fmt.Errorf("OAuth client initialization failed: %w", err)
	}

	// Always initialize public API as fallback
	publicFetcher := NewPublicJSONFetcher()

	return &Client{
		oauthFetcher:  oauthClient, // May be nil if OAuth not configured
		publicFetcher: publicFetcher,
		cache:         cache,
		debugLogger:   nil,         // No debug logger by default
	}, nil
}

// NewClientWithDebug creates a new Reddit client with debug logging enabled.
func NewClientWithDebug(debugLogger DebugLogger) (*Client, error) {
	cache, err := NewGoalLinkCache()
	if err != nil {
		return nil, fmt.Errorf("create cache: %w", err)
	}

	// Try OAuth first (higher rate limits, no CAPTCHAs)
	oauthClient, err := NewOAuthClientWithDebug(debugLogger)
	if err != nil {
		return nil, fmt.Errorf("OAuth client initialization failed: %w", err)
	}

	// Always initialize public API as fallback
	publicFetcher := NewPublicJSONFetcher()

	return &Client{
		oauthFetcher:  oauthClient, // May be nil if OAuth not configured
		publicFetcher: publicFetcher,
		cache:         cache,
		debugLogger:   debugLogger,
	}, nil
}

// NewClientWithFetcher creates a new Reddit client with a custom fetcher.
// Use this for testing or when switching to OAuth API.
func NewClientWithFetcher(fetcher Fetcher, cache *GoalLinkCache) *Client {
	return &Client{
		oauthFetcher:  nil, // No OAuth in test mode
		publicFetcher: fetcher,
		cache:         cache,
		debugLogger:   nil,
	}
}

// GoalLink retrieves a cached goal link or fetches from Reddit if not cached.
// Returns nil if the goal link was previously searched but not found.
func (c *Client) GoalLink(goal GoalInfo) (*GoalLink, error) {
	key := GoalLinkKey{MatchID: goal.MatchID, Minute: goal.Minute}

	// Check cache first (includes "not found" markers)
	if link := c.cache.Get(key); link != nil {
		// If this is a "not found" marker, return nil (don't re-search)
		if IsNotFound(link) {
			return nil, nil
		}
		return link, nil
	}

	// Search Reddit for the goal
	link, err := c.searchForGoal(goal)
	if err != nil {
		// Don't cache errors - allow retry
		return nil, err
	}

	if link != nil {
		// Cache the result (silently ignore cache errors - best-effort)
		_ = c.cache.Set(*link)
	} else {
		// Cache "not found" to avoid re-searching
		_ = c.cache.SetNotFound(goal.MatchID, goal.Minute)
	}

	return link, nil
}

// BatchSize is the maximum number of goals to fetch per batch.
const BatchSize = 5

// BatchDelay is the delay between batches to avoid rate limiting.
const BatchDelay = 2 * time.Second

// GoalLinks retrieves links for multiple goals, using cache where available.
// Goals are de-duplicated and batched to avoid rate limiting.
func (c *Client) GoalLinks(goals []GoalInfo) map[GoalLinkKey]*GoalLink {
	results := make(map[GoalLinkKey]*GoalLink)

	// De-duplicate goals by key and filter out already-cached goals
	seen := make(map[GoalLinkKey]bool)
	var uncachedGoals []GoalInfo

	for _, goal := range goals {
		key := GoalLinkKey{MatchID: goal.MatchID, Minute: goal.Minute}

		// Skip duplicates
		if seen[key] {
			continue
		}
		seen[key] = true

		// Check cache first
		if link := c.cache.Get(key); link != nil {
			if !IsNotFound(link) {
				results[key] = link
			}
			// Skip - already cached (found or not found)
			continue
		}

		uncachedGoals = append(uncachedGoals, goal)
	}

	// Fetch uncached goals in batches
	for i := 0; i < len(uncachedGoals); i += BatchSize {
		// Add delay between batches (not before first batch)
		if i > 0 {
			time.Sleep(BatchDelay)
		}

		// Process batch
		end := i + BatchSize
		if end > len(uncachedGoals) {
			end = len(uncachedGoals)
		}

		for _, goal := range uncachedGoals[i:end] {
			key := GoalLinkKey{MatchID: goal.MatchID, Minute: goal.Minute}
			link, err := c.GoalLink(goal)
			if err == nil && link != nil {
				results[key] = link
			}
		}
	}

	return results
}

// searchForGoal searches Reddit for a specific goal.
// Uses OAuth API if available (preferred), falls back to public API with retry logic.
func (c *Client) searchForGoal(goal GoalInfo) (*GoalLink, error) {
	// Try OAuth API first if available (no CAPTCHA issues, higher rate limits)
	if c.oauthFetcher != nil && c.oauthFetcher.IsAvailable() {
		c.debugLog(fmt.Sprintf("Using OAuth API for goal %d:%d", goal.MatchID, goal.Minute))
		result, err := c.searchForGoalOnce(goal, c.oauthFetcher)
		if err != nil {
			c.debugLog(fmt.Sprintf("OAuth API failed for goal %d:%d: %v", goal.MatchID, goal.Minute, err))
			return nil, err
		}
		return result, nil
	}

	// Fall back to public API with retry logic for CAPTCHA handling
	c.debugLog(fmt.Sprintf("Using public API for goal %d:%d (OAuth not available)", goal.MatchID, goal.Minute))
	maxRetries := 3
	baseDelay := 30 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * baseDelay
			c.debugLog(fmt.Sprintf("Retrying goal %d:%d in %v (attempt %d/%d)", goal.MatchID, goal.Minute, delay, attempt+1, maxRetries))
			time.Sleep(delay)
		}

		result, err := c.searchForGoalOnce(goal, c.publicFetcher)
		if err == nil {
			return result, nil
		}

		// Check if this is a CAPTCHA/rate limit error that we should retry
		if strings.Contains(err.Error(), "CAPTCHA") ||
			strings.Contains(err.Error(), "blocking requests") ||
			strings.Contains(err.Error(), "rate limit") ||
			strings.Contains(err.Error(), "HTML instead of JSON") {
			c.debugLog(fmt.Sprintf("Reddit API error for goal %d:%d (attempt %d/%d): %v", goal.MatchID, goal.Minute, attempt+1, maxRetries, err))
			if attempt < maxRetries-1 {
				continue
			}
			c.debugLog(fmt.Sprintf("Max retries exceeded for goal %d:%d: %v", goal.MatchID, goal.Minute, err))
		}

		// For other errors or if we've exhausted retries, return the error
		c.debugLog(fmt.Sprintf("Non-retryable error for goal %d:%d: %v", goal.MatchID, goal.Minute, err))
		return nil, err
	}

	return nil, nil // No match found after all retries
}

// searchForGoalOnce performs a single search attempt for a goal using the specified fetcher.
func (c *Client) searchForGoalOnce(goal GoalInfo, fetcher Fetcher) (*GoalLink, error) {
	// Strategy 1: Both teams + minute (most specific, try first)
	query1 := fmt.Sprintf("%s %s %d'", goal.HomeTeam, goal.AwayTeam, goal.Minute)
	results1, err := fetcher.Search(query1, 15, goal.MatchTime)
	if err == nil {
		// Check if we found a good match with the first strategy
		match := findBestMatch(results1, goal)
		if match != nil {
			// Found a match, return it immediately to avoid additional API calls
			return &GoalLink{
				MatchID:   goal.MatchID,
				Minute:    goal.Minute,
				URL:       match.URL,
				Title:     match.Title,
				PostURL:   match.PostURL,
				FetchedAt: time.Now(),
			}, nil
		}
	}

	// Strategy 1 didn't find a match, try broader searches
	// Only try one additional strategy to balance coverage vs rate limiting
	var allResults []SearchResult
	if err == nil {
		allResults = append(allResults, results1...)
	}

	// Strategy 2: Try with just the scoring team + minute
	// Determine which team scored
	scoringTeam := goal.AwayTeam
	if goal.IsHomeTeam {
		scoringTeam = goal.HomeTeam
	}
	query2 := fmt.Sprintf("%s %d'", scoringTeam, goal.Minute)
	results2, err := fetcher.Search(query2, 15, goal.MatchTime)
	if err == nil {
		allResults = append(allResults, results2...)
	}

	// Remove duplicates based on URL
	seen := make(map[string]bool)
	uniqueResults := make([]SearchResult, 0, len(allResults))
	for _, result := range allResults {
		if !seen[result.URL] {
			seen[result.URL] = true
			uniqueResults = append(uniqueResults, result)
		}
	}

	// Find the best matching result
	match := findBestMatch(uniqueResults, goal)
	if match == nil {
		return nil, nil // No match found, but not an error
	}

	return &GoalLink{
		MatchID:   goal.MatchID,
		Minute:    goal.Minute,
		URL:       match.URL,
		Title:     match.Title,
		PostURL:   match.PostURL,
		FetchedAt: time.Now(),
	}, nil
}

// ClearCache clears the goal link cache.
func (c *Client) ClearCache() error {
	return c.cache.Clear()
}

// Cache returns the underlying cache for direct access if needed.
func (c *Client) Cache() *GoalLinkCache {
	return c.cache
}

// isCaptchaResponse detects if Reddit is serving a CAPTCHA or bot detection page.
// This happens when Reddit blocks automated requests.
func isCaptchaResponse(body string) bool {
	captchaIndicators := []string{
		"prove your humanity",
		"captcha",
		"robot",
		"automated",
		"blocked",
		"rate limit",
		"too many requests",
	}

	bodyLower := strings.ToLower(body)
	for _, indicator := range captchaIndicators {
		if strings.Contains(bodyLower, indicator) {
			return true
		}
	}

	return false
}
