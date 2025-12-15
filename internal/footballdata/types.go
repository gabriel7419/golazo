package footballdata

import (
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// footballdataMatch represents a match in API-Football.com format
type footballdataMatch struct {
	Fixture struct {
		ID     int    `json:"id"`
		Date   string `json:"date"`
		Status struct {
			Short string `json:"short"` // FT, NS, LIVE, etc.
			Long  string `json:"long"`
		} `json:"status"`
		Venue struct {
			Name string `json:"name"`
		} `json:"venue"`
	} `json:"fixture"`

	League struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Country string `json:"country"`
		Logo    string `json:"logo"`
		Round   string `json:"round,omitempty"`
	} `json:"league"`

	Teams struct {
		Home struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Winner *bool  `json:"winner,omitempty"`
		} `json:"home"`
		Away struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Winner *bool  `json:"winner,omitempty"`
		} `json:"away"`
	} `json:"teams"`

	Goals struct {
		Home *int `json:"home"`
		Away *int `json:"away"`
	} `json:"goals"`

	Score struct {
		FullTime struct {
			Home *int `json:"home"`
			Away *int `json:"away"`
		} `json:"fulltime"`
		HalfTime struct {
			Home *int `json:"home"`
			Away *int `json:"away"`
		} `json:"halftime"`
		ExtraTime struct {
			Home *int `json:"home"`
			Away *int `json:"away"`
		} `json:"extratime,omitempty"`
		Penalty struct {
			Home *int `json:"home"`
			Away *int `json:"away"`
		} `json:"penalty,omitempty"`
	} `json:"score"`
}

// footballdataMatchesResponse represents the response from API-Football.com fixtures endpoint
type footballdataMatchesResponse struct {
	Response []footballdataMatch `json:"response"`
	Errors   map[string]string   `json:"errors,omitempty"`
}

// toAPIMatch converts a footballdataMatch to api.Match
func (m footballdataMatch) toAPIMatch() api.Match {
	match := api.Match{
		ID: m.Fixture.ID,
		League: api.League{
			ID:      m.League.ID,
			Name:    m.League.Name,
			Country: m.League.Country,
			Logo:    m.League.Logo,
		},
		HomeTeam: api.Team{
			ID:   m.Teams.Home.ID,
			Name: m.Teams.Home.Name,
		},
		AwayTeam: api.Team{
			ID:   m.Teams.Away.ID,
			Name: m.Teams.Away.Name,
		},
		Round: m.League.Round,
	}

	// Set short names (use name if short name not available)
	match.HomeTeam.ShortName = m.Teams.Home.Name
	match.AwayTeam.ShortName = m.Teams.Away.Name

	// Parse match time
	if m.Fixture.Date != "" {
		if t, err := time.Parse(time.RFC3339, m.Fixture.Date); err == nil {
			match.MatchTime = &t
		}
	}

	// Map status (API-Football.com uses short status codes)
	switch m.Fixture.Status.Short {
	case "FT", "AET", "PEN":
		match.Status = api.MatchStatusFinished
	case "LIVE", "HT", "1H", "2H", "ET", "BT", "P", "SUSP", "INT":
		match.Status = api.MatchStatusLive
	case "NS", "TBD":
		match.Status = api.MatchStatusNotStarted
	case "CANC", "PST":
		match.Status = api.MatchStatusCancelled
	case "POST":
		match.Status = api.MatchStatusPostponed
	default:
		match.Status = api.MatchStatusNotStarted
	}

	// Set scores - prefer goals field, fallback to score.fulltime
	if m.Goals.Home != nil {
		homeScore := *m.Goals.Home
		match.HomeScore = &homeScore
	} else if m.Score.FullTime.Home != nil {
		homeScore := *m.Score.FullTime.Home
		match.HomeScore = &homeScore
	}

	if m.Goals.Away != nil {
		awayScore := *m.Goals.Away
		match.AwayScore = &awayScore
	} else if m.Score.FullTime.Away != nil {
		awayScore := *m.Score.FullTime.Away
		match.AwayScore = &awayScore
	}

	return match
}
