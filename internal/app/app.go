// Package app implements the main application model and view navigation logic.
package app

import (
	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/0xjuanma/golazo/internal/ui"
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
}

// NewModel creates a new application model with default values.
func NewModel() model {
	return model{
		currentView: viewMain,
		selected:    0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.currentView != viewMain {
				m.currentView = viewMain
				m.selected = 0
				m.matchDetails = nil
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
			// Load matches and switch to live matches view
			matches, err := data.MockMatches()
			if err != nil {
				// If loading fails, switch view with empty matches
				// Error is silently ignored for now - could be logged in future
				m.currentView = viewLiveMatches
				m.matches = []ui.MatchDisplay{}
				return m, nil
			}

			// Convert to display format
			displayMatches := make([]ui.MatchDisplay, 0, len(matches))
			for _, match := range matches {
				displayMatches = append(displayMatches, ui.MatchDisplay{
					Match: match,
				})
			}

			m.matches = displayMatches
			m.currentView = viewLiveMatches
			m.selected = 0

			// Load details for first match if available
			if len(m.matches) > 0 {
				if details, err := data.MockMatchDetails(m.matches[0].ID); err == nil {
					m.matchDetails = details
				}
			}

			return m, nil
		}
		// Favourites does nothing for now
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
				if details, err := data.MockMatchDetails(m.matches[m.selected].ID); err == nil {
					m.matchDetails = details
				}
			}
		}
		return m, nil
	case "k", "up":
		if m.selected > 0 {
			m.selected--
			// Load details for newly selected match
			if m.selected >= 0 && m.selected < len(m.matches) {
				if details, err := data.MockMatchDetails(m.matches[m.selected].ID); err == nil {
					m.matchDetails = details
				}
			}
		}
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	switch m.currentView {
	case viewMain:
		return ui.RenderMainMenu(m.width, m.height, m.selected)
	case viewLiveMatches:
		return ui.RenderMultiPanelView(m.width, m.height, m.matches, m.selected, m.matchDetails)
	default:
		return ui.RenderMainMenu(m.width, m.height, m.selected)
	}
}
