package ui

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/constants"
	"github.com/charmbracelet/lipgloss"
)

// RenderStatsView renders the stats view with finished matches and high-level statistics.
// Layout: 50% left (matches list), 50% right split vertically (overview top, statistics bottom)
func RenderStatsView(width, height int, matches []MatchDisplay, selected int, details *api.MatchDetails) string {
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

	panelHeight := height - 2
	rightPanelHeight := panelHeight / 2 // Split right panel vertically

	// Render left panel (finished matches list) - full height
	leftPanel := renderFinishedMatchesPanel(leftWidth, panelHeight, matches, selected)

	// Render right panels (overview top, statistics bottom)
	overviewPanel := renderMatchOverviewPanel(rightWidth, rightPanelHeight, details)
	statisticsPanel := renderMatchStatisticsPanel(rightWidth, rightPanelHeight, details)

	// Create vertical separator between left and right
	verticalSeparatorStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Height(panelHeight).
		Padding(0, 1)
	verticalSeparator := verticalSeparatorStyle.Render("│")

	// Create horizontal separator between overview and statistics
	horizontalSeparatorStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Width(rightWidth).
		Padding(0, 1)
	horizontalSeparator := horizontalSeparatorStyle.Render(strings.Repeat("─", rightWidth-2))

	// Combine right panels vertically
	rightPanels := lipgloss.JoinVertical(
		lipgloss.Left,
		overviewPanel,
		horizontalSeparator,
		statisticsPanel,
	)

	// Combine left and right horizontally
	combined := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		verticalSeparator,
		rightPanels,
	)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, combined)
}

// renderFinishedMatchesPanel renders the left panel with finished matches list.
func renderFinishedMatchesPanel(width, height int, matches []MatchDisplay, selected int) string {
	title := panelTitleStyle.Width(width - 6).Render(constants.PanelFinishedMatches)

	items := make([]string, 0, len(matches))
	contentWidth := width - 6

	if len(matches) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(1, 0).
			Align(lipgloss.Center).
			Width(contentWidth)
		items = append(items, emptyStyle.Render(constants.EmptyNoFinishedMatches))
	} else {
		for i, match := range matches {
			item := renderFinishedMatchListItem(match, i == selected, contentWidth)
			items = append(items, item)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, items...)

	panelContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
	)

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(panelContent)

	return panel
}

// renderFinishedMatchListItem renders a single finished match list item.
func renderFinishedMatchListItem(match MatchDisplay, selected bool, width int) string {
	style := matchListItemStyle
	if selected {
		style = matchListItemSelectedStyle
	}

	// Format: "Team A 2 - 1 Team B"
	scoreText := "vs"
	if match.HomeScore != nil && match.AwayScore != nil {
		scoreText = fmt.Sprintf("%d - %d", *match.HomeScore, *match.AwayScore)
	}

	homeTeam := Truncate(match.HomeTeam.ShortName, 12)
	awayTeam := Truncate(match.AwayTeam.ShortName, 12)

	matchLine := fmt.Sprintf("%s %s %s", homeTeam, scoreText, awayTeam)

	// Add league name below
	leagueName := Truncate(match.League.Name, width-4)

	// Add date if available
	dateText := ""
	if match.MatchTime != nil {
		dateText = match.MatchTime.Format("Jan 2")
	}

	item := matchLine
	if leagueName != "" {
		item += "\n" + lipgloss.NewStyle().
			Foreground(dimColor).
			Render(leagueName)
	}
	if dateText != "" {
		item += " " + lipgloss.NewStyle().
			Foreground(dimColor).
			Render(dateText)
	}

	return style.Width(width).Render(item)
}

// renderMatchOverviewPanel renders Panel 2: Match Overview & Timeline (right top, 50% height)
func renderMatchOverviewPanel(width, height int, details *api.MatchDetails) string {
	title := panelTitleStyle.Width(width - 6).Render("Match Overview")

	if details == nil {
		emptyStyle := lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(1, 0).
			Align(lipgloss.Center).
			Width(width - 6)
		content := lipgloss.JoinVertical(
			lipgloss.Center,
			emptyStyle.Render(constants.EmptySelectMatch),
		)
		panelContent := lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			content,
		)
		panel := panelStyle.
			Width(width).
			Height(height).
			Render(panelContent)
		return panel
	}

	contentWidth := width - 6
	content := make([]string, 0)

	// Match header with winner indicator
	header := renderMatchHeaderWithWinner(details, contentWidth)
	content = append(content, header)

	// Score section
	if details.HomeScore != nil && details.AwayScore != nil {
		scoreText := fmt.Sprintf("%d - %d", *details.HomeScore, *details.AwayScore)
		content = append(content, lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center).
			MarginTop(0).
			MarginBottom(0).
			Render(matchScoreStyle.Render(scoreText)))
	}

	// Half-time score
	if details.HalfTimeScore != nil {
		htText := "HT: "
		if details.HalfTimeScore.Home != nil && details.HalfTimeScore.Away != nil {
			htText += fmt.Sprintf("%d - %d", *details.HalfTimeScore.Home, *details.HalfTimeScore.Away)
		}
		content = append(content, lipgloss.NewStyle().
			Width(contentWidth).
			Align(lipgloss.Center).
			Foreground(dimColor).
			Render(htText))
	}

	// Match context (venue, league, date)
	info := make([]string, 0)
	if details.Venue != "" {
		info = append(info, lipgloss.NewStyle().
			Foreground(dimColor).
			Render("Venue: "+details.Venue))
	}
	if details.League.Name != "" {
		info = append(info, lipgloss.NewStyle().
			Foreground(dimColor).
			Render(constants.LabelLeague+details.League.Name))
	}
	if details.MatchTime != nil {
		info = append(info, lipgloss.NewStyle().
			Foreground(dimColor).
			Render(constants.LabelDate+details.MatchTime.Format("Jan 2, 2006")))
	}
	if len(info) > 0 {
		content = append(content, strings.Join(info, " | "))
	}

	// Match duration indicator
	durationText := fmt.Sprintf("%d'", details.MatchDuration)
	if details.ExtraTime {
		durationText += " (AET)"
	}
	if details.Penalties != nil {
		if details.Penalties.Home != nil && details.Penalties.Away != nil {
			durationText += fmt.Sprintf(" (Pens: %d-%d)", *details.Penalties.Home, *details.Penalties.Away)
		}
	}
	content = append(content, lipgloss.NewStyle().
		Foreground(dimColor).
		Render(durationText))

	// Separator
	content = append(content, "")

	// Goal timeline
	if len(details.Events) > 0 {
		goals := make([]api.MatchEvent, 0)
		for _, event := range details.Events {
			if event.Type == "goal" {
				goals = append(goals, event)
			}
		}

		if len(goals) > 0 {
			content = append(content, lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true).
				Render("Goals Timeline:"))
			for _, goal := range goals {
				player := "Unknown"
				if goal.Player != nil {
					player = *goal.Player
				}
				teamName := goal.Team.ShortName
				assistText := ""
				if goal.Assist != nil && *goal.Assist != "" {
					assistText = fmt.Sprintf(" (assist: %s)", *goal.Assist)
				}
				goalText := fmt.Sprintf("%d' %s - %s%s", goal.Minute, teamName, player, assistText)
				content = append(content, lipgloss.NewStyle().
					Foreground(secondaryColor).
					Render(goalText))
			}
		}

		// Key events summary
		goalsCount := 0
		cardsCount := 0
		subsCount := 0
		for _, event := range details.Events {
			switch event.Type {
			case "goal":
				goalsCount++
			case "card":
				cardsCount++
			case "substitution":
				subsCount++
			}
		}

		if goalsCount > 0 || cardsCount > 0 || subsCount > 0 {
			content = append(content, "")
			summary := make([]string, 0)
			if goalsCount > 0 {
				summary = append(summary, fmt.Sprintf("Goals: %d", goalsCount))
			}
			if cardsCount > 0 {
				summary = append(summary, fmt.Sprintf("Cards: %d", cardsCount))
			}
			if subsCount > 0 {
				summary = append(summary, fmt.Sprintf("Subs: %d", subsCount))
			}
			content = append(content, lipgloss.NewStyle().
				Foreground(dimColor).
				Render(strings.Join(summary, " | ")))
		}
	}

	panelContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(panelContent)

	return panel
}

// renderMatchStatisticsPanel renders Panel 3: Match Statistics & Performance (right bottom, 50% height)
func renderMatchStatisticsPanel(width, height int, details *api.MatchDetails) string {
	title := panelTitleStyle.Width(width - 6).Render("Match Statistics")

	if details == nil {
		emptyStyle := lipgloss.NewStyle().
			Foreground(dimColor).
			Padding(1, 0).
			Align(lipgloss.Center).
			Width(width - 6)
		content := lipgloss.JoinVertical(
			lipgloss.Center,
			emptyStyle.Render(constants.EmptySelectMatch),
		)
		panelContent := lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			content,
		)
		panel := panelStyle.
			Width(width).
			Height(height).
			Render(panelContent)
		return panel
	}

	stats := make([]string, 0)

	// For now, show placeholder for statistics
	// Phase 3 will add actual statistics from API
	stats = append(stats, lipgloss.NewStyle().
		Foreground(dimColor).
		Render("Statistics will be displayed here"))

	// Count cards for display (from events)
	if len(details.Events) > 0 {
		homeYellow := 0
		homeRed := 0
		awayYellow := 0
		awayRed := 0

		for _, event := range details.Events {
			if event.Type == "card" {
				isHome := event.Team.ID == details.HomeTeam.ID
				if event.EventType != nil {
					if *event.EventType == "yellow" {
						if isHome {
							homeYellow++
						} else {
							awayYellow++
						}
					} else if *event.EventType == "red" {
						if isHome {
							homeRed++
						} else {
							awayRed++
						}
					}
				}
			}
		}

		if homeYellow > 0 || homeRed > 0 || awayYellow > 0 || awayRed > 0 {
			stats = append(stats, "")
			stats = append(stats, lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true).
				Render("Cards:"))
			cardsText := fmt.Sprintf("🟨%d 🟥%d ━━━━━━━━━━ 🟨%d 🟥%d",
				homeYellow, homeRed, awayYellow, awayRed)
			stats = append(stats, cardsText)
		}
	}

	panelContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		lipgloss.JoinVertical(lipgloss.Left, stats...),
	)

	panel := panelStyle.
		Width(width).
		Height(height).
		Render(panelContent)

	return panel
}

// renderMatchHeaderWithWinner renders the match header with team names and winner indicator.
func renderMatchHeaderWithWinner(details *api.MatchDetails, width int) string {
	homeTeam := Truncate(details.HomeTeam.ShortName, 18)
	awayTeam := Truncate(details.AwayTeam.ShortName, 18)

	header := fmt.Sprintf("%s vs %s", homeTeam, awayTeam)

	// Add winner indicator
	if details.Winner != nil {
		if *details.Winner == "home" {
			header += " ✓"
		} else if *details.Winner == "away" {
			header = "✓ " + header
		}
	}

	return matchTitleStyle.Width(width).Render(header)
}
