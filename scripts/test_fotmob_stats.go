package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/0xjuanma/golazo/internal/fotmob"
)

func main() {
	fmt.Println("Testing FotMob Stats Data Retrieval (Optimized)")
	fmt.Println("================================================\n")

	// Create FotMob client
	client := fotmob.NewClient()

	// Test: Fetch all stats data using the optimized unified function
	fmt.Println("Fetching stats data (5 days finished + today upcoming)...")
	fmt.Println("This makes 84 API calls: 28 for today + 14 each for 4 past days")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	startTime := time.Now()
	statsData, err := client.StatsData(ctx)
	elapsed := time.Since(startTime)

	if err != nil {
		fmt.Printf("❌ Error fetching stats data: %v\n", err)
		return
	}

	fmt.Printf("✓ Stats data fetched in %v\n\n", elapsed)

	// Display results
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("FINISHED MATCHES (All 3 days)")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Total: %d matches\n\n", len(statsData.AllFinished))

	if len(statsData.AllFinished) > 0 {
		fmt.Println("Sample (first 5):")
		for i, match := range statsData.AllFinished {
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
			dateStr := "?"
			if match.MatchTime != nil {
				dateStr = match.MatchTime.Format("Jan 2 15:04")
			}
			fmt.Printf("  %d. %s %s-%s %s (%s) - %s\n",
				i+1,
				match.HomeTeam.ShortName,
				homeScore,
				awayScore,
				match.AwayTeam.ShortName,
				dateStr,
				match.League.Name,
			)
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("TODAY'S FINISHED MATCHES")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Total: %d matches\n", len(statsData.TodayFinished))

	fmt.Println()
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("TODAY'S UPCOMING MATCHES")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Total: %d matches\n\n", len(statsData.TodayUpcoming))

	if len(statsData.TodayUpcoming) > 0 {
		fmt.Println("Sample (first 5):")
		for i, match := range statsData.TodayUpcoming {
			if i >= 5 {
				break
			}
			dateStr := "?"
			if match.MatchTime != nil {
				dateStr = match.MatchTime.Format("15:04")
			}
			fmt.Printf("  %d. %s vs %s (%s) - %s\n",
				i+1,
				match.HomeTeam.ShortName,
				match.AwayTeam.ShortName,
				dateStr,
				match.League.Name,
			)
		}
	}

	// Test match details if we have matches
	fmt.Println()
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println("MATCH DETAILS TEST")
	fmt.Println(strings.Repeat("-", 50))

	if len(statsData.AllFinished) > 0 {
		matchID := statsData.AllFinished[0].ID
		fmt.Printf("Fetching details for match ID: %d...\n", matchID)

		ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel2()

		details, err := client.MatchDetails(ctx2, matchID)
		if err != nil {
			fmt.Printf("❌ Error fetching match details: %v\n", err)
		} else {
			fmt.Printf("✓ Match details retrieved\n")
			fmt.Printf("  Home: %s, Away: %s\n", details.HomeTeam.Name, details.AwayTeam.Name)
			fmt.Printf("  Events: %d\n", len(details.Events))
			if details.Venue != "" {
				fmt.Printf("  Venue: %s\n", details.Venue)
			}
		}
	} else {
		fmt.Println("Skipped (no matches available)")
	}

	// Summary
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("SUMMARY")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("  Fetch time:        %v\n", elapsed)
	fmt.Printf("  5-day finished:    %d matches\n", len(statsData.AllFinished))
	fmt.Printf("  Today finished:    %d matches\n", len(statsData.TodayFinished))
	fmt.Printf("  Today upcoming:    %d matches\n", len(statsData.TodayUpcoming))
	fmt.Println()
	fmt.Println("Note: Switching between Today/5d view in the app")
	fmt.Println("      will be INSTANT (client-side filtering)")
	fmt.Println(strings.Repeat("=", 50))
}
