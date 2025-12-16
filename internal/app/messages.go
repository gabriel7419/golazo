package app

import "github.com/0xjuanma/golazo/internal/api"

// liveUpdateMsg contains a live update string for match events.
type liveUpdateMsg struct {
	update string
}

// matchDetailsMsg contains match details from API response.
type matchDetailsMsg struct {
	details *api.MatchDetails
}

// liveMatchesMsg contains live matches from API response.
type liveMatchesMsg struct {
	matches []api.Match
}

// finishedMatchesMsg contains finished matches from API response.
type finishedMatchesMsg struct {
	matches []api.Match
}

// upcomingMatchesMsg contains upcoming matches from API response.
type upcomingMatchesMsg struct {
	matches []api.Match
}

