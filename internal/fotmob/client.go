package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

const (
	baseURL = "https://www.fotmob.com/api"
)

// Supported league IDs for match fetching
var (
	// SupportedLeagues contains the league IDs that will be queried for matches.
	// FotMob league IDs:
	//
	//   Top 5 European Leagues:
	//   - Premier League: 47
	//   - La Liga: 87
	//   - Bundesliga: 54
	//   - Serie A (Italy): 55
	//   - Ligue 1: 53
	//
	//   European Competitions:
	//   - UEFA Champions League: 42
	//   - UEFA Europa League: 73
	//   - UEFA Euro: 50
	//
	//   South America:
	//   - Brasileirão Série A: 268
	//   - Liga Profesional Argentina: 112
	//   - Copa Libertadores: 14
	//   - Copa America: 44
	//
	//   Other:
	//   - MLS (USA): 130
	//   - FIFA World Cup: 77
	//
	SupportedLeagues = []int{
		// Top 5 European Leagues
		47, // Premier League
		87, // La Liga
		54, // Bundesliga
		55, // Serie A (Italy)
		53, // Ligue 1
		// European Competitions
		42, // UEFA Champions League
		73, // UEFA Europa League
		50, // UEFA Euro
		// South America
		268, // Brasileirão Série A
		112, // Liga Profesional Argentina
		14,  // Copa Libertadores
		44,  // Copa America
		// Other
		130, // MLS
		77,  // FIFA World Cup
	}
)

// Client implements the api.Client interface for FotMob API
type Client struct {
	httpClient  *http.Client
	baseURL     string
	rateLimiter *RateLimiter
	cache       *ResponseCache
}

// NewClient creates a new FotMob API client with default configuration.
// Includes minimal rate limiting (200ms between requests) for fast concurrent requests.
// Uses default caching configuration for improved performance.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseURL:     baseURL,
		rateLimiter: NewRateLimiter(200 * time.Millisecond), // Minimal delay for concurrent requests
		cache:       NewResponseCache(DefaultCacheConfig()),
	}
}

// GetCache returns the response cache for external access (e.g., pre-fetching).
func (c *Client) GetCache() *ResponseCache {
	return c.cache
}

// MatchesByDate retrieves all matches for a specific date.
// Since FotMob doesn't have a single endpoint for all matches by date,
// we query each supported league separately and filter by date client-side.
// We query both "fixtures" (upcoming) and "results" (finished) tabs concurrently.
// All requests are made concurrently with minimal rate limiting for maximum speed.
// Results are cached to avoid redundant API calls.
func (c *Client) MatchesByDate(ctx context.Context, date time.Time) ([]api.Match, error) {
	// Normalize date to UTC for consistent comparison
	requestDateStr := date.UTC().Format("2006-01-02")

	// Check cache first
	if cached := c.cache.GetMatches(requestDateStr); cached != nil {
		return cached, nil
	}

	// Use a mutex to protect the shared slice
	var mu sync.Mutex
	var allMatches []api.Match

	// Query leagues concurrently - no stagger delays, just rate limiting
	// Best-effort aggregation: if a league query fails, we skip it and continue with others
	// This allows partial results even if some leagues are unavailable
	var wg sync.WaitGroup

	// Query both fixtures (upcoming) and results (finished) tabs
	tabs := []string{"fixtures", "results"}
	for _, tab := range tabs {
		for _, leagueID := range SupportedLeagues {
			wg.Add(1)
			go func(id int, tabName string) {
				defer wg.Done()

				// Apply rate limiting (minimal delay for concurrent requests)
				c.rateLimiter.Wait()

				url := fmt.Sprintf("%s/leagues?id=%d&tab=%s", c.baseURL, id, tabName)

				req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				if err != nil {
					// Skip this league on error - best effort aggregation
					return
				}

				req.Header.Set("User-Agent", "Mozilla/5.0")

				resp, err := c.httpClient.Do(req)
				if err != nil {
					// Skip this league on request error - best effort aggregation
					return
				}
				defer resp.Body.Close()

				var leagueResponse struct {
					Details struct {
						ID          int    `json:"id"`
						Name        string `json:"name"`
						Country     string `json:"country"`
						CountryCode string `json:"countryCode,omitempty"`
					} `json:"details"`
					Fixtures struct {
						AllMatches []fotmobMatch `json:"allMatches"`
					} `json:"fixtures"`
				}

				if err := json.NewDecoder(resp.Body).Decode(&leagueResponse); err != nil {
					// Skip this league on parse error - best effort aggregation
					return
				}

				// Filter matches for the requested date and add league info
				// Note: Matches are sorted chronologically, so we need to check all matches
				var leagueMatches []api.Match
				for _, m := range leagueResponse.Fixtures.AllMatches {
					// Check if match is on the requested date
					if m.Status.UTCTime != "" {
						// Parse the UTC time - FotMob sometimes uses .000Z format
						var matchTime time.Time
						var err error
						matchTime, err = time.Parse(time.RFC3339, m.Status.UTCTime)
						if err != nil {
							// Try alternative format with milliseconds (.000Z)
							matchTime, err = time.Parse("2006-01-02T15:04:05.000Z", m.Status.UTCTime)
						}
						if err == nil {
							// Compare dates in UTC to avoid timezone issues
							matchDateStr := matchTime.UTC().Format("2006-01-02")
							if matchDateStr == requestDateStr {
								// Set league info from the response details
								if m.League.ID == 0 {
									m.League = league{
										ID:          leagueResponse.Details.ID,
										Name:        leagueResponse.Details.Name,
										Country:     leagueResponse.Details.Country,
										CountryCode: leagueResponse.Details.CountryCode,
									}
								}
								leagueMatches = append(leagueMatches, m.toAPIMatch())
							}
						}
					}
				}

				// Append to shared slice with mutex protection
				mu.Lock()
				allMatches = append(allMatches, leagueMatches...)
				mu.Unlock()
			}(leagueID, tab)
		}
	}

	wg.Wait()

	// Cache the results before returning
	c.cache.SetMatches(requestDateStr, allMatches)

	return allMatches, nil
}

// MatchDetails retrieves detailed information about a specific match.
// Results are cached to avoid redundant API calls.
func (c *Client) MatchDetails(ctx context.Context, matchID int) (*api.MatchDetails, error) {
	// Check cache first
	if cached := c.cache.GetDetails(matchID); cached != nil {
		return cached, nil
	}

	// Apply rate limiting
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/matchDetails?matchId=%d", c.baseURL, matchID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for match %d: %w", matchID, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch match details for match %d: %w", matchID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for match %d", resp.StatusCode, matchID)
	}

	var response fotmobMatchDetails

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode match details response for match %d: %w", matchID, err)
	}

	details := response.toAPIMatchDetails()

	// Cache the result
	c.cache.SetDetails(matchID, details)

	return details, nil
}

// BatchMatchDetails retrieves details for multiple matches concurrently.
// Uses caching and rate limiting to balance speed with API limits.
// Returns a map of matchID -> details (nil if fetch failed).
func (c *Client) BatchMatchDetails(ctx context.Context, matchIDs []int) map[int]*api.MatchDetails {
	results := make(map[int]*api.MatchDetails)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range matchIDs {
		wg.Add(1)
		go func(matchID int) {
			defer wg.Done()

			details, err := c.MatchDetails(ctx, matchID)
			if err != nil {
				// Store nil for failed fetches
				mu.Lock()
				results[matchID] = nil
				mu.Unlock()
				return
			}

			mu.Lock()
			results[matchID] = details
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	return results
}

// PreFetchMatchDetails fetches details for the first N matches in the background.
// This improves perceived performance by pre-loading details before user selection.
// maxConcurrent limits how many concurrent requests to make.
func (c *Client) PreFetchMatchDetails(ctx context.Context, matchIDs []int, maxPrefetch int) {
	if len(matchIDs) == 0 {
		return
	}

	// Limit the number of matches to prefetch
	if maxPrefetch > 0 && len(matchIDs) > maxPrefetch {
		matchIDs = matchIDs[:maxPrefetch]
	}

	// Filter out already cached matches
	var uncachedIDs []int
	for _, id := range matchIDs {
		if c.cache.GetDetails(id) == nil {
			uncachedIDs = append(uncachedIDs, id)
		}
	}

	if len(uncachedIDs) == 0 {
		return
	}

	// Fetch uncached matches in the background (fire and forget)
	go func() {
		c.BatchMatchDetails(ctx, uncachedIDs)
	}()
}

// Leagues retrieves available leagues.
func (c *Client) Leagues(ctx context.Context) ([]api.League, error) {
	// FotMob doesn't have a direct leagues endpoint, so we'll return an empty slice
	// In a real implementation, you might need to maintain a list of known leagues
	// or fetch them from a different endpoint
	return []api.League{}, nil
}

// LeagueMatches retrieves matches for a specific league.
func (c *Client) LeagueMatches(ctx context.Context, leagueID int) ([]api.Match, error) {
	// This would require a different endpoint structure
	// For now, we'll return an empty slice
	// In a real implementation, you'd use: /api/leagues?id={leagueID}
	return []api.Match{}, nil
}

// LeagueTable retrieves the league table/standings for a specific league.
func (c *Client) LeagueTable(ctx context.Context, leagueID int) ([]api.LeagueTableEntry, error) {
	// Apply rate limiting
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/leagues?id=%d", c.baseURL, leagueID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request for league %d table: %w", leagueID, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch league table for league %d: %w", leagueID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for league %d table", resp.StatusCode, leagueID)
	}

	var response struct {
		Data struct {
			Table []fotmobTableRow `json:"table"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode league table response for league %d: %w", leagueID, err)
	}

	entries := make([]api.LeagueTableEntry, 0, len(response.Data.Table))
	for _, row := range response.Data.Table {
		entries = append(entries, row.toAPITableEntry())
	}

	return entries, nil
}
