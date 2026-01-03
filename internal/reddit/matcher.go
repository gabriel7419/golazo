package reddit

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Matcher provides loose matching for Reddit goal post titles.
// Example titles:
//   - "Wolves [3] - 0 West Ham - Mateus Mane 41'"
//   - "Manchester United [2] - 1 Liverpool - Marcus Rashford 67'"
//   - "Barcelona 0 - [1] Real Madrid - Vinicius Jr 89'"

// findBestMatch finds the best matching search result for a goal.
// Uses loose matching: checks for team names, minute, and date proximity.
func findBestMatch(results []SearchResult, goal GoalInfo) *SearchResult {
	if len(results) == 0 {
		return nil
	}

	// Normalize team names for comparison
	homeNorm := normalizeTeamName(goal.HomeTeam)
	awayNorm := normalizeTeamName(goal.AwayTeam)
	minutePattern := buildMinutePattern(goal.Minute)

	var bestMatch *SearchResult
	bestScore := 0

	for i := range results {
		result := &results[i]
		titleLower := strings.ToLower(result.Title)

		score := 0

		// Filter by date: post must be within reasonable time of match
		// Allow posts from 1 day before to 2 days after the match
		if !goal.MatchTime.IsZero() {
			postDate := result.CreatedAt
			matchStart := goal.MatchTime.Add(-24 * time.Hour)
			matchEnd := goal.MatchTime.Add(48 * time.Hour)

			if postDate.Before(matchStart) || postDate.After(matchEnd) {
				continue // Post is outside the valid date range
			}

			// Bonus for posts very close to match time (within 12 hours)
			if postDate.After(goal.MatchTime.Add(-6*time.Hour)) && postDate.Before(goal.MatchTime.Add(12*time.Hour)) {
				score += 5
			}
		}

		// Check for team names (required)
		homeFound := containsTeamName(titleLower, homeNorm)
		awayFound := containsTeamName(titleLower, awayNorm)

		if !homeFound && !awayFound {
			continue // Must have at least one team name
		}

		if homeFound {
			score += 10
		}
		if awayFound {
			score += 10
		}

		// Check for minute (highly valuable)
		if minutePattern.MatchString(result.Title) {
			score += 25
		}

		// Check for scorer name if available
		if goal.ScorerName != "" {
			scorerNorm := normalizeName(goal.ScorerName)
			if containsName(titleLower, scorerNorm) {
				score += 15
			}
		}

		// Prefer higher Reddit score (upvotes) as tiebreaker
		score += min(result.Score/100, 5) // Max 5 points from upvotes

		if score > bestScore {
			bestScore = score
			bestMatch = result
		}
	}

	// Require minimum score for a match
	if bestScore < 20 {
		return nil
	}

	return bestMatch
}

// normalizeTeamName converts a team name to a normalized form for matching.
func normalizeTeamName(name string) string {
	// Convert to lowercase
	norm := strings.ToLower(name)

	// Remove common suffixes
	suffixes := []string{" fc", " cf", " sc", " afc", " united", " city"}
	for _, suffix := range suffixes {
		norm = strings.TrimSuffix(norm, suffix)
	}

	// Remove special characters
	norm = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(norm, "")

	return strings.TrimSpace(norm)
}

// normalizeName converts a player name to a normalized form for matching.
func normalizeName(name string) string {
	norm := strings.ToLower(name)
	// Remove special characters but keep spaces
	norm = regexp.MustCompile(`[^a-z\s]`).ReplaceAllString(norm, "")
	return strings.TrimSpace(norm)
}

// containsTeamName checks if a title contains a team name (or part of it).
func containsTeamName(title, teamNorm string) bool {
	// First try exact match
	if strings.Contains(title, teamNorm) {
		return true
	}

	// Try matching individual words (for multi-word team names)
	words := strings.Fields(teamNorm)
	if len(words) > 1 {
		// Check if the main word (usually the first significant word) is present
		for _, word := range words {
			if len(word) > 3 && strings.Contains(title, word) {
				return true
			}
		}
	}

	return false
}

// containsName checks if a title contains a player name.
func containsName(title, nameNorm string) bool {
	// First try full name
	if strings.Contains(title, nameNorm) {
		return true
	}

	// Try matching last name (usually more unique)
	parts := strings.Fields(nameNorm)
	if len(parts) > 0 {
		lastName := parts[len(parts)-1]
		if len(lastName) > 2 && strings.Contains(title, lastName) {
			return true
		}
	}

	return false
}

// buildMinutePattern creates a regex to match a minute in various formats.
// Matches: "41'", "41" (at word boundary), "41+2'" etc.
func buildMinutePattern(minute int) *regexp.Regexp {
	// Match the minute with optional added time
	// e.g., "45", "45'", "45+2", "45+2'"
	patternStr := `\b` + strconv.Itoa(minute) + `(\+\d+)?'?\b`
	compiled, err := regexp.Compile(patternStr)
	if err != nil {
		// Fallback to simple string match
		return regexp.MustCompile(strconv.Itoa(minute))
	}
	return compiled
}

// MatchConfidence represents how confident we are in a match.
type MatchConfidence int

const (
	ConfidenceNone   MatchConfidence = 0
	ConfidenceLow    MatchConfidence = 1
	ConfidenceMedium MatchConfidence = 2
	ConfidenceHigh   MatchConfidence = 3
)

// CalculateConfidence returns the confidence level for a match.
func CalculateConfidence(result SearchResult, goal GoalInfo) MatchConfidence {
	titleLower := strings.ToLower(result.Title)
	homeNorm := normalizeTeamName(goal.HomeTeam)
	awayNorm := normalizeTeamName(goal.AwayTeam)

	hasHome := containsTeamName(titleLower, homeNorm)
	hasAway := containsTeamName(titleLower, awayNorm)
	hasMinute := buildMinutePattern(goal.Minute).MatchString(result.Title)

	if hasHome && hasAway && hasMinute {
		return ConfidenceHigh
	}
	if (hasHome || hasAway) && hasMinute {
		return ConfidenceMedium
	}
	if hasHome || hasAway {
		return ConfidenceLow
	}
	return ConfidenceNone
}
