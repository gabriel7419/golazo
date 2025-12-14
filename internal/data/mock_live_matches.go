package data

import (
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// MockLiveMatches returns live matches for the live matches view.
func MockLiveMatches() []api.Match {
	now := time.Now()

	return []api.Match{
		{
			ID: 2001,
			League: api.League{
				ID:   39,
				Name: "Premier League",
			},
			HomeTeam: api.Team{
				ID:        49,
				Name:      "Chelsea",
				ShortName: "Chelsea",
			},
			AwayTeam: api.Team{
				ID:        66,
				Name:      "Tottenham",
				ShortName: "Spurs",
			},
			Status:    api.MatchStatusLive,
			HomeScore: intPtr(2),
			AwayScore: intPtr(1),
			LiveTime:  stringPtr("67'"),
			MatchTime: &now,
			Round:     "Matchday 17",
		},
		{
			ID: 2002,
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
				ID:        531,
				Name:      "Atletico Madrid",
				ShortName: "Atletico",
			},
			Status:    api.MatchStatusLive,
			HomeScore: intPtr(1),
			AwayScore: intPtr(0),
			LiveTime:  stringPtr("34'"),
			MatchTime: &now,
			Round:     "Matchday 18",
		},
		{
			ID: 2003,
			League: api.League{
				ID:   78,
				Name: "Bundesliga",
			},
			HomeTeam: api.Team{
				ID:        173,
				Name:      "RB Leipzig",
				ShortName: "Leipzig",
			},
			AwayTeam: api.Team{
				ID:        165,
				Name:      "Borussia Dortmund",
				ShortName: "Dortmund",
			},
			Status:    api.MatchStatusLive,
			HomeScore: intPtr(0),
			AwayScore: intPtr(0),
			LiveTime:  stringPtr("23'"),
			MatchTime: &now,
			Round:     "Matchday 15",
		},
		{
			ID: 2004,
			League: api.League{
				ID:   135,
				Name: "Serie A",
			},
			HomeTeam: api.Team{
				ID:        109,
				Name:      "Juventus",
				ShortName: "Juventus",
			},
			AwayTeam: api.Team{
				ID:        489,
				Name:      "AC Milan",
				ShortName: "AC Milan",
			},
			Status:    api.MatchStatusLive,
			HomeScore: intPtr(2),
			AwayScore: intPtr(2),
			LiveTime:  stringPtr("78'"),
			MatchTime: &now,
			Round:     "Matchday 16",
		},
	}
}
