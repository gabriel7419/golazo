package app

import (
	"context"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/fotmob"
	tea "github.com/charmbracelet/bubbletea"
)

// fetchLiveMatches fetches live matches from the API.
// Returns mock data if useMockData is true, otherwise uses real API.
func fetchLiveMatches(client *fotmob.Client, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			return liveMatchesMsg{matches: data.MockLiveMatches()}
		}

		if client == nil {
			return liveMatchesMsg{matches: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		matches, err := client.LiveMatches(ctx)
		if err != nil {
			return liveMatchesMsg{matches: nil}
		}

		return liveMatchesMsg{matches: matches}
	}
}

// fetchMatchDetails fetches match details from the API.
// Returns mock data if useMockData is true, otherwise uses real API.
func fetchMatchDetails(client *fotmob.Client, matchID int, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			details, _ := data.MockMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			return matchDetailsMsg{details: nil}
		}

		return matchDetailsMsg{details: details}
	}
}

// pollMatchDetails polls match details every 90 seconds for live updates.
// Conservative interval to avoid rate limiting.
func pollMatchDetails(client *fotmob.Client, parser *fotmob.LiveUpdateParser, matchID int, lastEvents []api.MatchEvent, useMockData bool) tea.Cmd {
	return tea.Tick(90*time.Second, func(t time.Time) tea.Msg {
		if useMockData {
			details, _ := data.MockMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			return matchDetailsMsg{details: nil}
		}

		return matchDetailsMsg{details: details}
	})
}

// fetchFinishedMatchesFotmob fetches finished matches from FotMob API.
// days specifies how many days to fetch (1 or 3).
// For 1-day view, uses optimized MatchesForToday to avoid duplicate API calls.
// Triggers background pre-fetching of match details for the first few matches.
func fetchFinishedMatchesFotmob(client *fotmob.Client, useMockData bool, days int) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			return finishedMatchesMsg{matches: data.MockFinishedMatches()}
		}

		if client == nil {
			return finishedMatchesMsg{matches: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var matches []api.Match

		// Optimized path for 1-day view
		if days == 1 {
			finished, _, err := client.MatchesForToday(ctx)
			if err != nil {
				return finishedMatchesMsg{matches: nil}
			}
			matches = finished
		} else {
			// Standard path for multi-day view
			fetched, err := client.RecentFinishedMatches(ctx, days)
			if err != nil {
				return finishedMatchesMsg{matches: nil}
			}
			matches = fetched
		}

		// Trigger background pre-fetching for the first 5 matches
		// This improves perceived performance when user navigates the list
		if len(matches) > 0 {
			matchIDs := make([]int, 0, min(5, len(matches)))
			for i := 0; i < len(matches) && i < 5; i++ {
				matchIDs = append(matchIDs, matches[i].ID)
			}
			// Use a separate context for background pre-fetching
			prefetchCtx, prefetchCancel := context.WithTimeout(context.Background(), 60*time.Second)
			go func() {
				defer prefetchCancel()
				client.PreFetchMatchDetails(prefetchCtx, matchIDs, 5)
			}()
		}

		return finishedMatchesMsg{matches: matches}
	}
}

// fetchUpcomingMatchesFotmob fetches upcoming matches from FotMob API for today.
// Only used when 1-day period is selected in stats view.
func fetchUpcomingMatchesFotmob(client *fotmob.Client, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			return upcomingMatchesMsg{matches: nil}
		}

		if client == nil {
			return upcomingMatchesMsg{matches: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_, upcoming, err := client.MatchesForToday(ctx)
		if err != nil {
			return upcomingMatchesMsg{matches: nil}
		}

		return upcomingMatchesMsg{matches: upcoming}
	}
}

// fetchStatsMatchDetailsFotmob fetches match details from FotMob API for stats view.
func fetchStatsMatchDetailsFotmob(client *fotmob.Client, matchID int, useMockData bool) tea.Cmd {
	return func() tea.Msg {
		if useMockData {
			details, _ := data.MockFinishedMatchDetails(matchID)
			return matchDetailsMsg{details: details}
		}

		if client == nil {
			return matchDetailsMsg{details: nil}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			return matchDetailsMsg{details: nil}
		}

		return matchDetailsMsg{details: details}
	}
}

