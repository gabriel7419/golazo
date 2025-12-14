package data

import (
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockMatchDetails returns detailed match information for a specific match ID.
func MockMatchDetails(matchID int) (*api.MatchDetails, error) {
	matches, err := MockMatches()
	if err != nil {
		return nil, err
	}

	// Find the match
	var match *api.Match
	for i := range matches {
		if matches[i].ID == matchID {
			match = &matches[i]
			break
		}
	}

	if match == nil {
		return nil, nil
	}

	// Generate events based on match ID
	events := generateMockEvents(matchID, *match)

	return &api.MatchDetails{
		Match:  *match,
		Events: events,
	}, nil
}

// generateMockEvents generates mock events for a live match.
func generateMockEvents(matchID int, match api.Match) []api.MatchEvent {
	events := []api.MatchEvent{}

	switch matchID {
	case 2001: // Chelsea 2-1 Spurs (Live)
		events = []api.MatchEvent{
			{ID: 1, Minute: 12, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Palmer"), Timestamp: time.Now()},
			{ID: 2, Minute: 34, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Son"), Timestamp: time.Now()},
			{ID: 3, Minute: 45, Type: "card", Team: match.AwayTeam, Player: stringPtr("Romero"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 4, Minute: 56, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Jackson"), Assist: stringPtr("Palmer"), Timestamp: time.Now()},
		}
	case 2002: // Real Madrid 1-0 Atletico (Live)
		events = []api.MatchEvent{
			{ID: 5, Minute: 18, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Bellingham"), Timestamp: time.Now()},
			{ID: 6, Minute: 28, Type: "card", Team: match.AwayTeam, Player: stringPtr("Savic"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
		}
	case 2003: // Leipzig 0-0 Dortmund (Live)
		events = []api.MatchEvent{
			{ID: 7, Minute: 15, Type: "card", Team: match.HomeTeam, Player: stringPtr("Simakan"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
		}
	case 2004: // Juventus 2-2 AC Milan (Live)
		events = []api.MatchEvent{
			{ID: 8, Minute: 8, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Vlahovic"), Timestamp: time.Now()},
			{ID: 9, Minute: 23, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Leao"), Timestamp: time.Now()},
			{ID: 10, Minute: 45, Type: "card", Team: match.HomeTeam, Player: stringPtr("Locatelli"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 11, Minute: 56, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Chiesa"), Timestamp: time.Now()},
			{ID: 12, Minute: 67, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Giroud"), Assist: stringPtr("Pulisic"), Timestamp: time.Now()},
		}
	}

	return events
}

func stringPtr(s string) *string {
	return &s
}
