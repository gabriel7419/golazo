package ui

import (
	"fmt"

	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/0xjuanma/golazo/internal/data"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Settings view uses the same neon colors as the rest of the app (red/cyan theme).
// Minimal design without heavy borders.

// SettingsState holds the state for the settings view.
type SettingsState struct {
	List       list.Model   // List component for league navigation
	Selected   map[int]bool // Map of league ID -> selected
	Leagues    []data.LeagueInfo
	HasChanges bool // Whether there are unsaved changes
}

// NewSettingsState creates a new settings state with current saved preferences.
func NewSettingsState() *SettingsState {
	settings, _ := data.LoadSettings()

	selected := make(map[int]bool)

	// If no leagues are selected in settings, none are checked
	// User sees unchecked = will use default leagues (Premier, La Liga, UCL)
	if len(settings.SelectedLeagues) > 0 {
		for _, id := range settings.SelectedLeagues {
			selected[id] = true
		}
	}

	leagues := data.AllSupportedLeagues

	// Create list items
	items := make([]list.Item, len(leagues))
	for i, league := range leagues {
		items[i] = LeagueListItem{
			League:   league,
			Selected: selected[league.ID],
		}
	}

	// Create and configure the list
	delegate := NewLeagueListDelegate()
	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowFilter(true)
	l.Filter = list.DefaultFilter
	l.SetShowHelp(false) // We use our own help text

	// Apply filter input styles
	filterCursorStyle, filterPromptStyle := FilterInputStyles()
	l.Styles.FilterCursor = filterCursorStyle
	l.FilterInput.PromptStyle = filterPromptStyle
	l.FilterInput.Cursor.Style = filterCursorStyle

	return &SettingsState{
		List:     l,
		Selected: selected,
		Leagues:  leagues,
	}
}

// Toggle toggles the selection state of the currently highlighted league.
func (s *SettingsState) Toggle() {
	if item, ok := s.List.SelectedItem().(LeagueListItem); ok {
		s.Selected[item.League.ID] = !s.Selected[item.League.ID]
		s.HasChanges = true
		s.refreshListItems()
	}
}

// refreshListItems updates the list items to reflect current selection state.
func (s *SettingsState) refreshListItems() {
	items := make([]list.Item, len(s.Leagues))
	for i, league := range s.Leagues {
		items[i] = LeagueListItem{
			League:   league,
			Selected: s.Selected[league.ID],
		}
	}
	s.List.SetItems(items)
}

// Save persists the current selection to settings.yaml.
func (s *SettingsState) Save() error {
	var selectedIDs []int
	for _, league := range s.Leagues {
		if s.Selected[league.ID] {
			selectedIDs = append(selectedIDs, league.ID)
		}
	}

	settings := &data.Settings{
		SelectedLeagues: selectedIDs,
	}

	err := data.SaveSettings(settings)
	if err == nil {
		s.HasChanges = false
	}
	return err
}

// SelectedCount returns the number of selected leagues.
func (s *SettingsState) SelectedCount() int {
	count := 0
	for _, isSelected := range s.Selected {
		if isSelected {
			count++
		}
	}
	return count
}

// Fixed width for settings panel
const settingsBoxWidth = 48

// RenderSettingsView renders the settings view for league customization.
// Uses minimal styling consistent with the rest of the app (red/cyan neon theme).
// debugMode shows a debug indicator when enabled.
func RenderSettingsView(width, height int, state *SettingsState, debugMode bool) string {
	if state == nil {
		return ""
	}

	// Calculate available space for the list
	const (
		titleHeight  = 3 // Title + margin
		infoHeight   = 2 // Selection info
		helpHeight   = 2 // Help text
		extraPadding = 4 // Additional vertical spacing
	)

	listWidth := settingsBoxWidth
	listHeight := height - titleHeight - infoHeight - helpHeight - extraPadding
	if listHeight < 5 {
		listHeight = 5
	}

	// Update list dimensions
	state.List.SetSize(listWidth, listHeight)

	// Add debug indicator if debug mode is enabled
	var debugIndicator string
	if debugMode {
		debugStyle := lipgloss.NewStyle().
			Foreground(neonCyan).
			Bold(true).
			Align(lipgloss.Center).
			Width(settingsBoxWidth)
		debugIndicator = debugStyle.Render("[DEBUG MODE] Logs: ~/.golazo/golazo_debug.log") + "\n"
	}

	// Title - red like other panel titles
	titleStyle := neonPanelTitleStyle.Width(settingsBoxWidth)
	title := titleStyle.Render("League Preferences")

	// Render the list
	listContent := state.List.View()

	// Selection info
	selectedCount := state.SelectedCount()
	var infoText string
	if selectedCount == 0 {
		infoText = "No selection = default leagues"
	} else {
		infoText = fmt.Sprintf("%d of %d selected", selectedCount, len(state.Leagues))
	}
	infoStyle := neonDimStyle.Width(settingsBoxWidth).Align(lipgloss.Center)
	info := infoStyle.Render(infoText)

	// Help text
	helpStyle := neonDimStyle.Align(lipgloss.Center)
	help := helpStyle.Render(constants.HelpSettingsView)

	// Combine content (minimal, no borders)
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		debugIndicator,
		title,
		"",
		listContent,
		"",
		info,
		help,
	)

	// Center in the terminal
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
