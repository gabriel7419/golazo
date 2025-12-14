package data

import (
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockFinishedMatchDetails returns detailed match information for finished matches.
func MockFinishedMatchDetails(matchID int) (*api.MatchDetails, error) {
	finishedMatches := MockFinishedMatches()

	// Find the match
	var match *api.Match
	for i := range finishedMatches {
		if finishedMatches[i].ID == matchID {
			match = &finishedMatches[i]
			break
		}
	}

	if match == nil {
		return nil, nil
	}

	// Generate events based on match ID
	events := generateFinishedMatchEvents(matchID, *match)

	return &api.MatchDetails{
		Match:  *match,
		Events: events,
	}, nil
}

// generateFinishedMatchEvents generates mock events for finished matches.
func generateFinishedMatchEvents(matchID int, match api.Match) []api.MatchEvent {
	events := []api.MatchEvent{}

	switch matchID {
	case 1001: // Man City 2-1 Arsenal
		events = []api.MatchEvent{
			{ID: 1, Minute: 12, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Haaland"), Timestamp: time.Now()},
			{ID: 2, Minute: 34, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Saka"), Assist: stringPtr("Odegaard"), Timestamp: time.Now()},
			{ID: 3, Minute: 45, Type: "card", Team: match.AwayTeam, Player: stringPtr("Partey"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 4, Minute: 67, Type: "goal", Team: match.HomeTeam, Player: stringPtr("De Bruyne"), Assist: stringPtr("Foden"), Timestamp: time.Now()},
		}
	case 1002: // Real Madrid 3-2 Barcelona
		events = []api.MatchEvent{
			{ID: 5, Minute: 8, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Lewandowski"), Timestamp: time.Now()},
			{ID: 6, Minute: 23, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Vinicius Jr"), Timestamp: time.Now()},
			{ID: 7, Minute: 34, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Benzema"), Assist: stringPtr("Modric"), Timestamp: time.Now()},
			{ID: 8, Minute: 45, Type: "card", Team: match.AwayTeam, Player: stringPtr("Gavi"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 9, Minute: 56, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Pedri"), Timestamp: time.Now()},
			{ID: 10, Minute: 78, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Rodrygo"), Timestamp: time.Now()},
		}
	case 1003: // Bayern 4-0 Dortmund
		events = []api.MatchEvent{
			{ID: 11, Minute: 15, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Kane"), Timestamp: time.Now()},
			{ID: 12, Minute: 28, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Musiala"), Assist: stringPtr("Sane"), Timestamp: time.Now()},
			{ID: 13, Minute: 45, Type: "card", Team: match.AwayTeam, Player: stringPtr("Hummels"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 14, Minute: 62, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Sane"), Timestamp: time.Now()},
			{ID: 15, Minute: 89, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Kane"), Assist: stringPtr("Muller"), Timestamp: time.Now()},
		}
	case 1004: // AC Milan 1-0 Inter
		events = []api.MatchEvent{
			{ID: 16, Minute: 23, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Leao"), Assist: stringPtr("Giroud"), Timestamp: time.Now()},
			{ID: 17, Minute: 45, Type: "card", Team: match.AwayTeam, Player: stringPtr("Barella"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 18, Minute: 67, Type: "card", Team: match.HomeTeam, Player: stringPtr("Theo"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
		}
	case 1005: // PSG 2-2 Marseille
		events = []api.MatchEvent{
			{ID: 19, Minute: 12, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Mbappe"), Timestamp: time.Now()},
			{ID: 20, Minute: 34, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Aubameyang"), Timestamp: time.Now()},
			{ID: 21, Minute: 56, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Dembele"), Assist: stringPtr("Mbappe"), Timestamp: time.Now()},
			{ID: 22, Minute: 78, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Clauss"), Timestamp: time.Now()},
		}
	case 1006: // Man City 1-1 Bayern (Champions League)
		events = []api.MatchEvent{
			{ID: 23, Minute: 18, Type: "goal", Team: match.HomeTeam, Player: stringPtr("Haaland"), Timestamp: time.Now()},
			{ID: 24, Minute: 45, Type: "card", Team: match.AwayTeam, Player: stringPtr("Upamecano"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 25, Minute: 67, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Kane"), Assist: stringPtr("Sane"), Timestamp: time.Now()},
		}
	case 1007: // Man Utd 0-3 Liverpool
		events = []api.MatchEvent{
			{ID: 26, Minute: 5, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Salah"), Timestamp: time.Now()},
			{ID: 27, Minute: 23, Type: "card", Team: match.HomeTeam, Player: stringPtr("Casemiro"), EventType: stringPtr("yellow"), Timestamp: time.Now()},
			{ID: 28, Minute: 45, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Nunez"), Assist: stringPtr("Salah"), Timestamp: time.Now()},
			{ID: 29, Minute: 67, Type: "card", Team: match.HomeTeam, Player: stringPtr("Martinez"), EventType: stringPtr("red"), Timestamp: time.Now()},
			{ID: 30, Minute: 89, Type: "goal", Team: match.AwayTeam, Player: stringPtr("Diaz"), Timestamp: time.Now()},
		}
	}

	return events
}
