// Package app implements the main application model and view navigation logic.
package app

import (
	"context"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/fotmob"
	"github.com/0xjuanma/golazo/internal/ui"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type view int

const (
	viewMain view = iota
	viewLiveMatches
)

type model struct {
	width        int
	height       int
	currentView  view
	matches      []ui.MatchDisplay
	selected     int
	matchDetails *api.MatchDetails
	liveUpdates  []string
	spinner      spinner.Model
	loading      bool
	fotmobClient *fotmob.Client
	parser       *fotmob.LiveUpdateParser
	lastEvents   []api.MatchEvent
	polling      bool
}

// NewModel creates a new application model with default values.
func NewModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = ui.SpinnerStyle()

	return model{
		currentView:  viewMain,
		selected:     0,
		spinner:      s,
		liveUpdates:  []string{},
		fotmobClient: fotmob.NewClient(),
		parser:       fotmob.NewLiveUpdateParser(),
		lastEvents:   []api.MatchEvent{},
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case spinner.TickMsg:
		if m.loading {
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	case liveUpdateMsg:
		if msg.update != "" {
			m.liveUpdates = append(m.liveUpdates, msg.update)
		}
		// Continue polling if match is live
		if m.polling && m.matchDetails != nil && m.matchDetails.Status == api.MatchStatusLive {
			cmds = append(cmds, pollMatchDetails(m.fotmobClient, m.parser, m.matchDetails.ID, m.lastEvents))
		} else {
			m.loading = false
			m.polling = false
		}
		return m, tea.Batch(cmds...)
	case matchDetailsMsg:
		if msg.details != nil {
			// Detect new events
			newEvents := m.parser.GetNewEvents(m.lastEvents, msg.details.Events)
			if len(newEvents) > 0 {
				// Parse new events into updates
				updates := m.parser.ParseEvents(newEvents, msg.details.HomeTeam, msg.details.AwayTeam)
				m.liveUpdates = append(m.liveUpdates, updates...)
			}
			m.matchDetails = msg.details
			m.lastEvents = msg.details.Events

			// Continue polling if match is live
			if msg.details.Status == api.MatchStatusLive {
				m.polling = true
				m.loading = true
				cmds = append(cmds, pollMatchDetails(m.fotmobClient, m.parser, msg.details.ID, m.lastEvents))
			} else {
				m.loading = false
				m.polling = false
			}
		}
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.currentView != viewMain {
				m.currentView = viewMain
				m.selected = 0
				m.matchDetails = nil
				m.liveUpdates = []string{}
				m.lastEvents = []api.MatchEvent{}
				m.loading = false
				m.polling = false
				return m, nil
			}
		}

		// Handle view-specific key events
		switch m.currentView {
		case viewMain:
			return m.handleMainViewKeys(msg)
		case viewLiveMatches:
			return m.handleLiveMatchesKeys(msg)
		}
	case liveMatchesMsg:
		// Convert to display format
		displayMatches := make([]ui.MatchDisplay, 0, len(msg.matches))
		for _, match := range msg.matches {
			displayMatches = append(displayMatches, ui.MatchDisplay{
				Match: match,
			})
		}

		m.matches = displayMatches
		m.selected = 0
		m.loading = false

		// Load details for first match if available
		if len(m.matches) > 0 {
			return m.loadMatchDetails(m.matches[0].ID)
		}

		return m, nil
	}
	return m, nil
}

func (m model) handleMainViewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < 1 {
			m.selected++
		}
		return m, nil
	case "k", "up":
		if m.selected > 0 {
			m.selected--
		}
		return m, nil
	case "enter":
		if m.selected == 0 {
			// Stats - do nothing, stay on main view
			return m, nil
		} else if m.selected == 1 {
			// Live Matches - fetch live matches from API
			m.currentView = viewLiveMatches
			m.loading = true
			return m, fetchLiveMatches(m.fotmobClient)
		}
		return m, nil
	}
	return m, nil
}

func (m model) handleLiveMatchesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.selected < len(m.matches)-1 {
			m.selected++
			// Load details for newly selected match
			if m.selected < len(m.matches) {
				return m.loadMatchDetails(m.matches[m.selected].ID)
			}
		}
		return m, nil
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			// Load details for newly selected match
			if m.selected >= 0 && m.selected < len(m.matches) {
				return m.loadMatchDetails(m.matches[m.selected].ID)
			}
		}
		return m, nil
	}
	return m, nil
}

// loadMatchDetails loads match details and starts live updates.
func (m model) loadMatchDetails(matchID int) (tea.Model, tea.Cmd) {
	m.liveUpdates = []string{}
	m.lastEvents = []api.MatchEvent{}
	m.loading = true
	return m, tea.Batch(m.spinner.Tick, fetchMatchDetails(m.fotmobClient, matchID))
}

func (m model) View() string {
	switch m.currentView {
	case viewMain:
		return ui.RenderMainMenu(m.width, m.height, m.selected)
	case viewLiveMatches:
		return ui.RenderMultiPanelView(m.width, m.height, m.matches, m.selected, m.matchDetails, m.liveUpdates, m.spinner, m.loading)
	default:
		return ui.RenderMainMenu(m.width, m.height, m.selected)
	}
}

// liveUpdateMsg is a message containing a live update string.
type liveUpdateMsg struct {
	update string
}

// matchDetailsMsg is a message containing match details.
type matchDetailsMsg struct {
	details *api.MatchDetails
}

// fetchLiveMatches fetches live matches from the API.
func fetchLiveMatches(client *fotmob.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		matches, err := client.LiveMatches(ctx)
		if err != nil {
			// Log error for debugging - in production you might want to use a logger
			// For now, return empty matches - will show empty state
			// The error details are in the API client error message
			matches = []api.Match{}
		}

		return liveMatchesMsg{matches: matches}
	}
}

// liveMatchesMsg is a message containing live matches.
type liveMatchesMsg struct {
	matches []api.Match
}

// fetchMatchDetails fetches match details from the API.
func fetchMatchDetails(client *fotmob.Client, matchID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			// Fallback to mock data on error
			details, _ = data.MockMatchDetails(matchID)
		}

		return matchDetailsMsg{details: details}
	}
}

// pollMatchDetails polls match details every 90 seconds for live updates.
// Conservative interval to avoid rate limiting (90 seconds = 1.5 minutes).
func pollMatchDetails(client *fotmob.Client, parser *fotmob.LiveUpdateParser, matchID int, lastEvents []api.MatchEvent) tea.Cmd {
	return tea.Tick(90*time.Second, func(t time.Time) tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		details, err := client.MatchDetails(ctx, matchID)
		if err != nil {
			return matchDetailsMsg{details: nil}
		}

		// Return match details - new events will be detected in the Update handler
		return matchDetailsMsg{details: details}
	})
}
