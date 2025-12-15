package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// RenderLiveMatchesListPanel renders the left panel using bubbletea list component.
// Note: listModel is passed by value, so SetSize must be called before this function.
func RenderLiveMatchesListPanel(width, height int, listModel list.Model) string {
	// Wrap list in panel
	title := panelTitleStyle.Width(width - 6).Render(constants.PanelLiveMatches)
	listView := listModel.View()

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		listView,
	)

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(content)

	return panel
}

// RenderStatsListPanel renders the left panel for stats view using bubbletea list component.
// Note: listModel is passed by value, so SetSize must be called before this function.
// upcomingMatches are displayed in the bottom half when dateRange is 1 (1-day selection).
func RenderStatsListPanel(width, height int, listModel list.Model, dateRange int, apiKeyMissing bool, upcomingMatches []MatchDisplay) string {
	// Render date range selector
	dateSelector := renderDateRangeSelector(width-6, dateRange)

	// Wrap list in panel
	title := panelTitleStyle.Width(width - 6).Render(constants.PanelFinishedMatches)

	var listView string
	if apiKeyMissing {
		// Show API key missing message instead of empty list
		emptyStyle := lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(2, 2).
			Align(lipgloss.Center).
			Width(width - 6)
		listView = emptyStyle.Render(constants.EmptyAPIKeyMissing)
	} else {
		listView = listModel.View()
		// Check if list is empty and show appropriate message
		if listModel.Items() == nil || len(listModel.Items()) == 0 {
			emptyStyle := lipgloss.NewStyle().
				Foreground(dimColor).
				Padding(2, 2).
				Align(lipgloss.Center).
				Width(width - 6)
			listView = emptyStyle.Render(constants.EmptyNoFinishedMatches + "\n\nTry selecting a different date range (h/l keys)")
		}
	}

	// Check if we have upcoming matches for 1d selection
	hasUpcoming := dateRange == 1 && len(upcomingMatches) > 0
	var upcomingMatchesSection string

	if hasUpcoming {
		// Render upcoming matches section
		upcomingTitle := panelTitleStyle.Width(width - 6).Render("Upcoming Matches")
		var upcomingList []string
		for _, match := range upcomingMatches {
			// Format: "Team1 vs Team2 - League"
			home := match.HomeTeam.ShortName
			if home == "" {
				home = match.HomeTeam.Name
			}
			away := match.AwayTeam.ShortName
			if away == "" {
				away = match.AwayTeam.Name
			}
			matchTime := ""
			if match.MatchTime != nil {
				matchTime = match.MatchTime.Format("15:04")
			}
			matchLine := fmt.Sprintf("  %s vs %s", home, away)
			if matchTime != "" {
				matchLine += fmt.Sprintf(" (%s)", matchTime)
			}
			if match.League.Name != "" {
				matchLine += fmt.Sprintf(" - %s", match.League.Name)
			}
			upcomingList = append(upcomingList, matchListItemStyle.Render(matchLine))
		}

		if len(upcomingList) == 0 {
			emptyStyle := lipgloss.NewStyle().
				Foreground(dimColor).
				Padding(1, 2).
				Align(lipgloss.Center).
				Width(width - 6)
			upcomingList = []string{emptyStyle.Render("No upcoming matches")}
		}

		upcomingContent := lipgloss.JoinVertical(lipgloss.Left, upcomingList...)
		upcomingMatchesSection = lipgloss.JoinVertical(
			lipgloss.Left,
			"",
			upcomingTitle,
			upcomingContent,
		)
	}

	// Truncate listView to fit finished matches height if needed
	finishedContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		dateSelector,
		"",
		listView,
	)

	var content string
	if hasUpcoming {
		// Add separator between finished and upcoming
		separator := lipgloss.NewStyle().
			Foreground(borderColor).
			Width(width - 6).
			Render(strings.Repeat("─", width-8))
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			finishedContent,
			separator,
			upcomingMatchesSection,
		)
	} else {
		content = finishedContent
	}

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(content)

	return panel
}

// renderDateRangeSelector renders a horizontal date range selector (1d, 3d).
func renderDateRangeSelector(width int, selected int) string {
	options := []struct {
		days  int
		label string
	}{
		{1, "1d"},
		{3, "3d"},
	}

	items := make([]string, 0, len(options))
	for _, opt := range options {
		if opt.days == selected {
			// Selected option - use highlight color
			item := matchListItemSelectedStyle.Render(opt.label)
			items = append(items, item)
		} else {
			// Unselected option - use normal color
			item := matchListItemStyle.Render(opt.label)
			items = append(items, item)
		}
	}

	// Join items with separator
	separator := "  "
	selector := strings.Join(items, separator)

	// Center the selector
	selectorStyle := lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Padding(0, 1)

	return selectorStyle.Render(selector)
}

// RenderMultiPanelViewWithList renders the live matches view with list component.
func RenderMultiPanelViewWithList(width, height int, listModel list.Model, details *api.MatchDetails, liveUpdates []string, sp spinner.Model, loading bool, randomSpinner *RandomCharSpinner, viewLoading bool) string {
	// Handle edge case: if width/height not set, use defaults
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Reserve 3 lines at top for spinner (always reserve to prevent layout shift)
	spinnerHeight := 3
	availableHeight := height - spinnerHeight
	if availableHeight < 10 {
		availableHeight = 10 // Minimum height for panels
	}

	// Render spinner centered in reserved space
	var spinnerArea string
	if viewLoading && randomSpinner != nil {
		spinnerView := randomSpinner.View()
		if spinnerView != "" {
			// Center the spinner horizontally using style with width and alignment
			spinnerStyle := lipgloss.NewStyle().
				Width(width).
				Height(spinnerHeight).
				Align(lipgloss.Center).
				AlignVertical(lipgloss.Center)
			spinnerArea = spinnerStyle.Render(spinnerView)
		} else {
			// Fallback if spinner view is empty
			spinnerStyle := lipgloss.NewStyle().
				Width(width).
				Height(spinnerHeight).
				Align(lipgloss.Center).
				AlignVertical(lipgloss.Center)
			spinnerArea = spinnerStyle.Render("Loading...")
		}
	} else {
		// Reserve space with empty lines - ensure it takes up exactly spinnerHeight lines
		spinnerArea = strings.Repeat("\n", spinnerHeight)
	}

	// Calculate panel dimensions
	leftWidth := width * 35 / 100
	if leftWidth < 25 {
		leftWidth = 25
	}
	rightWidth := width - leftWidth - 1
	if rightWidth < 35 {
		rightWidth = 35
		leftWidth = width - rightWidth - 1
	}

	// Use panelHeight similar to stats view to ensure proper spacing
	panelHeight := availableHeight - 2

	// Render left panel (matches list) - shifted down
	leftPanel := RenderLiveMatchesListPanel(leftWidth, panelHeight, listModel)

	// Render right panel (match details with live updates) - shifted down
	rightPanel := renderMatchDetailsPanel(rightWidth, panelHeight, details, liveUpdates, sp, loading)

	// Create separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Height(panelHeight).
		Padding(0, 1)
	separator := separatorStyle.Render("│")

	// Combine panels
	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		separator,
		rightPanel,
	)

	// Combine spinner area and panels - this shifts panels down
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		spinnerArea,
		panels,
	)

	return content
}

// RenderStatsViewWithList renders the stats view with list component.
func RenderStatsViewWithList(width, height int, listModel list.Model, details *api.MatchDetails, randomSpinner *RandomCharSpinner, viewLoading bool, dateRange int, apiKeyMissing bool, upcomingMatches []MatchDisplay) string {
	// Handle edge case: if width/height not set, use defaults
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	// Reserve 3 lines at top for spinner (always reserve to prevent layout shift)
	spinnerHeight := 3
	availableHeight := height - spinnerHeight
	if availableHeight < 10 {
		availableHeight = 10 // Minimum height for panels
	}

	// Render spinner centered in reserved space - EXACTLY like live view
	var spinnerArea string
	if viewLoading && randomSpinner != nil {
		spinnerView := randomSpinner.View()
		if spinnerView != "" {
			// Center the spinner horizontally using style with width and alignment - EXACTLY like live view
			spinnerStyle := lipgloss.NewStyle().
				Width(width).
				Height(spinnerHeight).
				Align(lipgloss.Center).
				AlignVertical(lipgloss.Center)
			spinnerArea = spinnerStyle.Render(spinnerView)
		} else {
			// Fallback if spinner view is empty
			spinnerStyle := lipgloss.NewStyle().
				Width(width).
				Height(spinnerHeight).
				Align(lipgloss.Center).
				AlignVertical(lipgloss.Center)
			spinnerArea = spinnerStyle.Render("Loading...")
		}
	} else {
		// Reserve space with empty lines - ensure it takes up exactly spinnerHeight lines
		spinnerArea = strings.Repeat("\n", spinnerHeight)
	}

	// Calculate panel dimensions
	// Left side: 50% width (matches list, full height)
	// Right side: 50% width (split vertically: overview top, statistics bottom)
	leftWidth := width * 50 / 100
	if leftWidth < 30 {
		leftWidth = 30
	}
	rightWidth := width - leftWidth - 1
	if rightWidth < 30 {
		rightWidth = 30
		leftWidth = width - rightWidth - 1
	}

	panelHeight := availableHeight - 2
	rightPanelHeight := panelHeight / 2 // Split right panel vertically

	// Render left panel (finished matches list) - full height
	leftPanel := RenderStatsListPanel(leftWidth, panelHeight, listModel, dateRange, apiKeyMissing, upcomingMatches)

	// Render right panels (overview top, statistics bottom)
	overviewPanel := renderMatchOverviewPanel(rightWidth, rightPanelHeight, details)
	statisticsPanel := renderMatchStatisticsPanel(rightWidth, rightPanelHeight, details)

	// Create vertical separator between left and right - match panel height exactly
	verticalSeparatorStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Height(panelHeight).
		Padding(0, 0)
	verticalSeparator := verticalSeparatorStyle.Render("│")

	// Create horizontal separator between overview and statistics
	// Use exact width without extra padding to avoid extra lines
	horizontalSeparatorStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Width(rightWidth)
	horizontalSeparator := horizontalSeparatorStyle.Render(strings.Repeat("─", rightWidth-4))

	// Combine right panels vertically - join directly without extra spacing
	rightPanels := lipgloss.JoinVertical(
		lipgloss.Top,
		overviewPanel,
		horizontalSeparator,
		statisticsPanel,
	)

	// Combine left and right horizontally
	panels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		verticalSeparator,
		rightPanels,
	)

	// Combine spinner area and panels - this shifts panels down
	// Use lipgloss.Left (not Top) to match live view exactly
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		spinnerArea,
		panels,
	)

	return content
}
