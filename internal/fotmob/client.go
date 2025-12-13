package fotmob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

const (
	baseURL = "https://www.fotmob.com/api"
)

// Supported league IDs for match fetching
// These can be easily extended by adding more league IDs to the slice
var (
	// SupportedLeagues contains the league IDs that will be queried for matches.
	// To add more leagues, simply append their IDs to this slice.
	// Known league IDs:
	//   - Premier League: 47
	//   - La Liga: 87
	//   - Bundesliga: 54
	//   - Champions League: 42
	//   - Serie A: 55
	//   - MLS: 130
	//   - Ligue 1: 53
	SupportedLeagues = []int{
		47,  // Premier League
		87,  // La Liga
		54,  // Bundesliga
		42,  // Champions League
		55,  // Serie A
		130, // MLS
	}
)

// Client implements the api.Client interface for FotMob API
type Client struct {
	httpClient  *http.Client
	baseURL     string
	rateLimiter *RateLimiter
}

// NewClient creates a new FotMob API client with default configuration.
// Includes conservative rate limiting (2 seconds between requests).
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL:     baseURL,
		rateLimiter: NewRateLimiter(2 * time.Second), // Conservative: 2 seconds between requests
	}
}

// MatchesByDate retrieves all matches for a specific date.
// Since the /api/matches endpoint returns 404, we query the supported leagues
// and aggregate their fixtures for the given date.
func (c *Client) MatchesByDate(ctx context.Context, date time.Time) ([]api.Match, error) {
	dateStr := date.Format("2006-01-02")
	allMatches := make([]api.Match, 0)

	// Query each supported league and aggregate matches
	for _, leagueID := range SupportedLeagues {
		// Apply rate limiting between league queries
		c.rateLimiter.Wait()

		url := fmt.Sprintf("%s/leagues?id=%d&tab=fixtures", c.baseURL, leagueID)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			continue // Skip this league on error
		}

		req.Header.Set("User-Agent", "Mozilla/5.0")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue // Skip this league on error
		}

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
			resp.Body.Close()
			continue // Skip this league on parse error
		}
		resp.Body.Close()

		// Filter matches for the requested date and add league info
		for _, m := range leagueResponse.Fixtures.AllMatches {
			// Check if match is on the requested date
			if m.Status.UTCTime != "" {
				matchTime, err := time.Parse(time.RFC3339, m.Status.UTCTime)
				if err == nil {
					matchDateStr := matchTime.Format("2006-01-02")
					if matchDateStr == dateStr {
						// Set league info from the response details
						if m.League.ID == 0 {
							m.League = league{
								ID:          leagueResponse.Details.ID,
								Name:        leagueResponse.Details.Name,
								Country:     leagueResponse.Details.Country,
								CountryCode: leagueResponse.Details.CountryCode,
							}
						}
						allMatches = append(allMatches, m.toAPIMatch())
					}
				}
			}
		}
	}

	return allMatches, nil
}

// MatchDetails retrieves detailed information about a specific match.
func (c *Client) MatchDetails(ctx context.Context, matchID int) (*api.MatchDetails, error) {
	// Apply rate limiting
	c.rateLimiter.Wait()

	url := fmt.Sprintf("%s/matchDetails?matchId=%d", c.baseURL, matchID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch match details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response fotmobMatchDetails

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.toAPIMatchDetails(), nil
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
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch league table: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Data struct {
			Table []fotmobTableRow `json:"table"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	entries := make([]api.LeagueTableEntry, 0, len(response.Data.Table))
	for _, row := range response.Data.Table {
		entries = append(entries, row.toAPITableEntry())
	}

	return entries, nil
}
