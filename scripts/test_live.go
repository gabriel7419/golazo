package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/0xjuanma/golazo/internal/fotmob"
)

func main() {
	client := fotmob.NewClient()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Fetching matches for today...")
	fmt.Printf("Supported leagues: %v\n", fotmob.SupportedLeagues)
	today := time.Now()
	fmt.Printf("Today's date: %s\n\n", today.Format("2006-01-02"))

	// Debug: Test leagues directly
	fmt.Println("Debug: Testing Bundesliga (54) directly...")
	testLeagueDirectly(54, today)
	fmt.Println()
	fmt.Println("Debug: Testing Premier League (47) directly...")
	testLeagueDirectly(47, today)
	fmt.Println()

	// First, get all matches for today to verify API is working
	allMatches, err := client.MatchesByDate(ctx, time.Now())
	if err != nil {
		fmt.Printf("❌ Error fetching all matches: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Found %d total matches for today\n", len(allMatches))

	// Count by status
	liveCount := 0
	finishedCount := 0
	notStartedCount := 0
	for _, match := range allMatches {
		switch match.Status {
		case "live":
			liveCount++
		case "finished":
			finishedCount++
		case "not_started":
			notStartedCount++
		}
	}

	fmt.Printf("  - Live: %d\n", liveCount)
	fmt.Printf("  - Finished: %d\n", finishedCount)
	fmt.Printf("  - Not Started: %d\n\n", notStartedCount)

	// Now get only live matches
	liveMatches, err := client.LiveMatches(ctx)
	if err != nil {
		fmt.Printf("❌ Error fetching live matches: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Found %d live matches (started but not finished)\n\n", len(liveMatches))

	if len(liveMatches) == 0 {
		fmt.Println("No live matches found for today.")
		fmt.Println("This could be normal if there are no matches currently in progress.")
		if len(allMatches) > 0 {
			fmt.Println("\nSample matches from today:")
			for i, match := range allMatches {
				if i >= 5 {
					break
				}
				homeScore := "?"
				awayScore := "?"
				if match.HomeScore != nil {
					homeScore = fmt.Sprintf("%d", *match.HomeScore)
				}
				if match.AwayScore != nil {
					awayScore = fmt.Sprintf("%d", *match.AwayScore)
				}
				fmt.Printf("  %s %s-%s %s [%s] - Status: %s\n",
					match.HomeTeam.ShortName,
					homeScore,
					awayScore,
					match.AwayTeam.ShortName,
					match.League.Name,
					match.Status,
				)
			}
		}
	} else {
		fmt.Println("Live matches:")
		for i, match := range liveMatches {
			homeScore := "?"
			awayScore := "?"
			if match.HomeScore != nil {
				homeScore = fmt.Sprintf("%d", *match.HomeScore)
			}
			if match.AwayScore != nil {
				awayScore = fmt.Sprintf("%d", *match.AwayScore)
			}

			liveTime := ""
			if match.LiveTime != nil {
				liveTime = fmt.Sprintf(" (%s)", *match.LiveTime)
			}

			fmt.Printf("  %d. %s %s-%s %s [%s]%s\n",
				i+1,
				match.HomeTeam.ShortName,
				homeScore,
				awayScore,
				match.AwayTeam.ShortName,
				match.League.Name,
				liveTime,
			)
		}
	}
}

func testLeagueDirectly(leagueID int, date time.Time) {
	url := fmt.Sprintf("https://www.fotmob.com/api/leagues?id=%d&tab=fixtures", leagueID)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("  ❌ Parse error: %v\n", err)
		return
	}

	if fixtures, ok := result["fixtures"].(map[string]interface{}); ok {
		if allMatches, ok := fixtures["allMatches"].([]interface{}); ok {
			fmt.Printf("  ✓ Found %d total matches in league\n", len(allMatches))

			dateStr := date.Format("2006-01-02")
			todayMatches := 0
			liveMatches := 0
			
			// Check ALL matches for today and live status (not just first 20)
			for i, m := range allMatches {
				if match, ok := m.(map[string]interface{}); ok {
					if status, ok := match["status"].(map[string]interface{}); ok {
						if utcTime, ok := status["utcTime"].(string); ok {
							if t, err := time.Parse(time.RFC3339, utcTime); err == nil {
								matchDateStr := t.Format("2006-01-02")
								started, _ := status["started"].(bool)
								finished, _ := status["finished"].(bool)

								if matchDateStr == dateStr {
									todayMatches++
									if started && !finished {
										liveMatches++
										// Show live match details
										home := "?"
										away := "?"
										if homeTeam, ok := match["home"].(map[string]interface{}); ok {
											if name, ok := homeTeam["shortName"].(string); ok {
												home = name
											}
										}
										if awayTeam, ok := match["away"].(map[string]interface{}); ok {
											if name, ok := awayTeam["shortName"].(string); ok {
												away = name
											}
										}
										score := ""
										if scoreObj, ok := status["score"].(map[string]interface{}); ok {
											if h, ok := scoreObj["home"].(float64); ok {
												if a, ok := scoreObj["away"].(float64); ok {
													score = fmt.Sprintf(" %d-%d ", int(h), int(a))
												}
											}
										}
										fmt.Printf("    🟢 LIVE: %s%s%s at %s\n", home, score, away, utcTime)
									}
								}

								// Show first few matches and any today's matches
								if i < 5 || matchDateStr == dateStr {
									fmt.Printf("    Match %d: %s (started=%v, finished=%v, date=%s)\n",
										i+1, utcTime, started, finished, matchDateStr)
								}
							}
						}
					}
				}
			}
			fmt.Printf("  Matches for today (%s): %d (Live: %d)\n", dateStr, todayMatches, liveMatches)
		}
	}
}
