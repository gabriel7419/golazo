package fotmob

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
)

// LiveMatches retrieves all currently live matches for today.
// Fetches matches from supported leagues and filters for those that have started but not finished.
// Matches are already filtered by supported leagues in MatchesByDate.
func (c *Client) LiveMatches(ctx context.Context) ([]api.Match, error) {
	today := time.Now()
	matches, err := c.MatchesByDate(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("fetch matches for date %s: %w", today.Format("2006-01-02"), err)
	}

	// Filter for live matches only (started but not finished)
	var liveMatches []api.Match
	for _, match := range matches {
		// Only include matches that are live (started but not finished)
		if match.Status == api.MatchStatusLive {
			liveMatches = append(liveMatches, match)
		}
	}

	return liveMatches, nil
}

// LiveUpdateParser parses match events into live update strings.
type LiveUpdateParser struct{}

// NewLiveUpdateParser creates a new live update parser.
func NewLiveUpdateParser() *LiveUpdateParser {
	return &LiveUpdateParser{}
}

// ParseEvents converts match events into human-readable update strings.
func (p *LiveUpdateParser) ParseEvents(events []api.MatchEvent, homeTeam, awayTeam api.Team) []string {
	updates := make([]string, 0, len(events))

	for _, event := range events {
		update := p.formatEvent(event, homeTeam, awayTeam)
		if update != "" {
			updates = append(updates, update)
		}
	}

	return updates
}

// formatEvent formats a single event into a readable string.
func (p *LiveUpdateParser) formatEvent(event api.MatchEvent, homeTeam, awayTeam api.Team) string {
	teamName := event.Team.ShortName
	if teamName == "" {
		// Fallback to team ID matching
		if event.Team.ID == homeTeam.ID {
			teamName = homeTeam.ShortName
		} else {
			teamName = awayTeam.ShortName
		}
	}

	switch strings.ToLower(event.Type) {
	case "goal":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		assistText := ""
		if event.Assist != nil && *event.Assist != "" {
			assistText = fmt.Sprintf(" (assist: %s)", *event.Assist)
		}
		return fmt.Sprintf("%d' Goal: %s%s - %s", event.Minute, player, assistText, teamName)

	case "card":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		cardType := "yellow"
		if event.EventType != nil {
			cardType = *event.EventType
		}
		return fmt.Sprintf("%d' Card (%s): %s - %s", event.Minute, cardType, player, teamName)

	case "substitution":
		player := "Unknown"
		if event.Player != nil {
			player = *event.Player
		}
		subType := "substitution"
		if event.EventType != nil {
			subType = *event.EventType
		}
		arrow := "→"
		if subType == "in" {
			arrow = "←"
		}
		return fmt.Sprintf("%d' Substitution: %s %s - %s", event.Minute, arrow, player, teamName)

	default:
		player := ""
		if event.Player != nil {
			player = *event.Player
		}
		if player != "" {
			return fmt.Sprintf("%d' • %s - %s (%s)", event.Minute, player, teamName, event.Type)
		}
		return fmt.Sprintf("%d' • %s - %s", event.Minute, event.Type, teamName)
	}
}

// NewEvents compares two event lists and returns only new events.
// This is useful for detecting new updates when polling match details.
func (p *LiveUpdateParser) NewEvents(oldEvents, newEvents []api.MatchEvent) []api.MatchEvent {
	// Create a map of old event IDs for quick lookup
	oldEventMap := make(map[int]bool)
	for _, event := range oldEvents {
		oldEventMap[event.ID] = true
	}

	// Find events that don't exist in old events
	var newOnly []api.MatchEvent
	for _, event := range newEvents {
		if !oldEventMap[event.ID] {
			newOnly = append(newOnly, event)
		}
	}

	return newOnly
}
