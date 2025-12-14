package data

import (
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockFinishedMatches returns finished matches for the stats view.
// These are realistic matches from recent fixtures.
func MockFinishedMatches() []api.Match {
	now := time.Now()

	return []api.Match{
		{
			ID: 1001,
			League: api.League{
				ID:   39,
				Name: "Premier League",
			},
			HomeTeam: api.Team{
				ID:        50,
				Name:      "Manchester City",
				ShortName: "Man City",
			},
			AwayTeam: api.Team{
				ID:        42,
				Name:      "Arsenal",
				ShortName: "Arsenal",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(2),
			AwayScore: intPtr(1),
			MatchTime: timePtr(now.AddDate(0, 0, -2)),
			Round:     "Matchday 16",
		},
		{
			ID: 1002,
			League: api.League{
				ID:   140,
				Name: "La Liga",
			},
			HomeTeam: api.Team{
				ID:        541,
				Name:      "Real Madrid",
				ShortName: "Real Madrid",
			},
			AwayTeam: api.Team{
				ID:        529,
				Name:      "Barcelona",
				ShortName: "Barcelona",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(3),
			AwayScore: intPtr(2),
			MatchTime: timePtr(now.AddDate(0, 0, -1)),
			Round:     "Matchday 17",
		},
		{
			ID: 1003,
			League: api.League{
				ID:   78,
				Name: "Bundesliga",
			},
			HomeTeam: api.Team{
				ID:        157,
				Name:      "Bayern Munich",
				ShortName: "Bayern",
			},
			AwayTeam: api.Team{
				ID:        165,
				Name:      "Borussia Dortmund",
				ShortName: "Dortmund",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(4),
			AwayScore: intPtr(0),
			MatchTime: timePtr(now.AddDate(0, 0, -3)),
			Round:     "Matchday 14",
		},
		{
			ID: 1004,
			League: api.League{
				ID:   135,
				Name: "Serie A",
			},
			HomeTeam: api.Team{
				ID:        489,
				Name:      "AC Milan",
				ShortName: "AC Milan",
			},
			AwayTeam: api.Team{
				ID:        108,
				Name:      "Inter Milan",
				ShortName: "Inter",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(1),
			AwayScore: intPtr(0),
			MatchTime: timePtr(now.AddDate(0, 0, -4)),
			Round:     "Matchday 15",
		},
		{
			ID: 1005,
			League: api.League{
				ID:   61,
				Name: "Ligue 1",
			},
			HomeTeam: api.Team{
				ID:        85,
				Name:      "Paris Saint-Germain",
				ShortName: "PSG",
			},
			AwayTeam: api.Team{
				ID:        516,
				Name:      "Marseille",
				ShortName: "Marseille",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(2),
			AwayScore: intPtr(2),
			MatchTime: timePtr(now.AddDate(0, 0, -5)),
			Round:     "Matchday 16",
		},
		{
			ID: 1006,
			League: api.League{
				ID:   2,
				Name: "UEFA Champions League",
			},
			HomeTeam: api.Team{
				ID:        50,
				Name:      "Manchester City",
				ShortName: "Man City",
			},
			AwayTeam: api.Team{
				ID:        157,
				Name:      "Bayern Munich",
				ShortName: "Bayern",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(1),
			AwayScore: intPtr(1),
			MatchTime: timePtr(now.AddDate(0, 0, -6)),
			Round:     "Group Stage - Matchday 5",
		},
		{
			ID: 1007,
			League: api.League{
				ID:   39,
				Name: "Premier League",
			},
			HomeTeam: api.Team{
				ID:        33,
				Name:      "Manchester United",
				ShortName: "Man Utd",
			},
			AwayTeam: api.Team{
				ID:        40,
				Name:      "Liverpool",
				ShortName: "Liverpool",
			},
			Status:    api.MatchStatusFinished,
			HomeScore: intPtr(0),
			AwayScore: intPtr(3),
			MatchTime: timePtr(now.AddDate(0, 0, -7)),
			Round:     "Matchday 15",
		},
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
