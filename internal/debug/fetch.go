// Package debug provides utilities for debugging and inspecting API data.
// Used by scripts in the scripts/ directory.
package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/fotmob"
)

const (
	baseURL = "https://www.fotmob.com/api"
)

// FetchLeagueData fetches raw API response and converts to matches for a specific league, date, and tab.
// Supports optional season parameter for historical data (e.g., "2022" for World Cup 2022).
func FetchLeagueData(ctx context.Context, leagueID int, date time.Time, tab string, season string) (map[string]any, []api.Match, error) {
	url := fmt.Sprintf("%s/leagues?id=%d&tab=%s", baseURL, leagueID, tab)
	if season != "" {
		url += "&season=" + season
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Read the full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}

	// Parse as raw JSON for display
	var rawResponse map[string]any
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, nil, fmt.Errorf("parse json: %w", err)
	}

	// Parse matches directly from the response (to respect season parameter)
	requestDateStr := date.UTC().Format("2006-01-02")
	matches := parseMatchesFromResponse(body, requestDateStr, leagueID)

	return rawResponse, matches, nil
}

// FetchMatchDetails fetches raw match details and converts to MatchDetails struct.
func FetchMatchDetails(ctx context.Context, matchID int) (map[string]any, *api.MatchDetails, error) {
	url := fmt.Sprintf("%s/matchDetails?matchId=%d", baseURL, matchID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Read the full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}

	// Parse as raw JSON for display
	var rawResponse map[string]any
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, nil, fmt.Errorf("parse json: %w", err)
	}

	// Use the fotmob client to get properly converted details
	fotmobClient := fotmob.NewClient()
	details, err := fotmobClient.MatchDetails(ctx, matchID)
	if err != nil {
		// Return raw response even if conversion fails
		return rawResponse, nil, nil
	}

	return rawResponse, details, nil
}

// parseMatchesFromResponse parses matches from raw API response and filters by date
func parseMatchesFromResponse(body []byte, requestDateStr string, leagueID int) []api.Match {
	var response struct {
		Details struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Country     string `json:"country"`
			CountryCode string `json:"countryCode,omitempty"`
		} `json:"details"`
		Fixtures struct {
			AllMatches []struct {
				ID    string `json:"id"`
				Round string `json:"round"`
				Home  struct {
					ID        string `json:"id"`
					Name      string `json:"name"`
					ShortName string `json:"shortName"`
				} `json:"home"`
				Away struct {
					ID        string `json:"id"`
					Name      string `json:"name"`
					ShortName string `json:"shortName"`
				} `json:"away"`
				Status struct {
					UTCTime   string `json:"utcTime"`
					Started   *bool  `json:"started"`
					Finished  *bool  `json:"finished"`
					Cancelled *bool  `json:"cancelled"`
					LiveTime  *struct {
						Short string `json:"short"`
					} `json:"liveTime,omitempty"`
					Score *struct {
						Home int `json:"home"`
						Away int `json:"away"`
					} `json:"score,omitempty"`
				} `json:"status"`
			} `json:"allMatches"`
		} `json:"fixtures"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil
	}

	var matches []api.Match
	for _, m := range response.Fixtures.AllMatches {
		if m.Status.UTCTime == "" {
			continue
		}

		// Parse match time
		var matchTime time.Time
		var parseErr error
		matchTime, parseErr = time.Parse(time.RFC3339, m.Status.UTCTime)
		if parseErr != nil {
			matchTime, parseErr = time.Parse("2006-01-02T15:04:05.000Z", m.Status.UTCTime)
		}
		if parseErr != nil {
			continue
		}

		// Filter by date
		matchDateStr := matchTime.UTC().Format("2006-01-02")
		if matchDateStr != requestDateStr {
			continue
		}

		// Convert IDs
		matchID, _ := strconv.Atoi(m.ID)
		homeID, _ := strconv.Atoi(m.Home.ID)
		awayID, _ := strconv.Atoi(m.Away.ID)

		match := api.Match{
			ID: matchID,
			League: api.League{
				ID:      response.Details.ID,
				Name:    response.Details.Name,
				Country: response.Details.Country,
			},
			HomeTeam: api.Team{
				ID:        homeID,
				Name:      m.Home.Name,
				ShortName: m.Home.ShortName,
			},
			AwayTeam: api.Team{
				ID:        awayID,
				Name:      m.Away.Name,
				ShortName: m.Away.ShortName,
			},
			MatchTime: &matchTime,
			Round:     m.Round,
		}

		// Determine status
		if m.Status.Cancelled != nil && *m.Status.Cancelled {
			match.Status = api.MatchStatusCancelled
		} else if m.Status.Finished != nil && *m.Status.Finished {
			match.Status = api.MatchStatusFinished
		} else if m.Status.Started != nil && *m.Status.Started {
			match.Status = api.MatchStatusLive
			if m.Status.LiveTime != nil {
				match.LiveTime = &m.Status.LiveTime.Short
			}
		} else {
			match.Status = api.MatchStatusNotStarted
		}

		// Set scores
		if m.Status.Score != nil {
			match.HomeScore = &m.Status.Score.Home
			match.AwayScore = &m.Status.Score.Away
		}

		matches = append(matches, match)
	}

	return matches
}
