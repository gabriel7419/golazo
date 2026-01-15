package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/0xjuanma/golazo/internal/fotmob"
)

func main() {
	var matchID int
	var clearAll bool
	var listCached bool
	var teamFilter string
	var forceRefresh bool

	flag.IntVar(&matchID, "match", 0, "Match ID to clear from cache (0 = clear all)")
	flag.BoolVar(&clearAll, "all", false, "Clear all cached match details")
	flag.BoolVar(&listCached, "list", false, "List all currently cached matches")
	flag.StringVar(&teamFilter, "team", "", "Clear cache for matches containing this team name (case-insensitive)")
	flag.BoolVar(&forceRefresh, "force", false, "Force refresh a match (clears cache + fetches fresh data)")
	flag.Parse()

	client := fotmob.NewClient()
	cache := client.Cache()

	// List cached matches
	if listCached {
		showCachedMatches(cache)
		fmt.Println("\n💡 Note: This shows FotMob client cache. App may have additional cached data.")
		return
	}

	// Clear by team name
	if teamFilter != "" {
		clearMatchesByTeam(cache, teamFilter)
		return
	}

	// Force refresh specific match
	if forceRefresh && matchID > 0 {
		fmt.Printf("🔄 Force refreshing match ID: %d (clearing cache + fetching fresh)...\n", matchID)
		ctx := context.Background()
		details, err := client.MatchDetailsForceRefresh(ctx, matchID)
		if err != nil {
			log.Printf("❌ Error force refreshing match details: %v", err)
			return
		}

		if details != nil {
			fmt.Printf("✅ Successfully refreshed data:\n")
			fmt.Printf("   Match: %s vs %s\n", details.Match.HomeTeam.Name, details.Match.AwayTeam.Name)
			fmt.Printf("   Status: %s\n", details.Status)
			if details.Match.HomeScore != nil && details.Match.AwayScore != nil {
				fmt.Printf("   Score: %d-%d\n", *details.Match.HomeScore, *details.Match.AwayScore)
			}
			fmt.Printf("   Events: %d events\n", len(details.Events))
		} else {
			fmt.Println("⚠️  No match details found")
		}
		return
	}

	// Clear all matches
	if clearAll || matchID == 0 {
		fmt.Println("Clearing all cached match details...")
		cache.ClearDetails()
		fmt.Println("✅ All match details cache cleared")
		fmt.Println("💡 Note: If golazo app is running, restart it to clear its internal cache too.")
		return
	}

	// Clear specific match
	fmt.Printf("Clearing cache for match ID: %d...\n", matchID)
	cache.ClearMatchDetails(matchID)
	fmt.Printf("✅ Match %d cache cleared\n", matchID)
	fmt.Println("💡 Note: If golazo app is running, restart it to clear its internal cache too.")

	// Test fetching fresh data
	fmt.Printf("\nTesting fresh data retrieval for match %d...\n", matchID)
	ctx := context.Background()
	details, err := client.MatchDetails(ctx, matchID)
	if err != nil {
		log.Printf("❌ Error fetching match details: %v", err)
		return
	}

	if details != nil {
		fmt.Printf("✅ Successfully fetched fresh data:\n")
		fmt.Printf("   Match: %s vs %s\n", details.Match.HomeTeam.Name, details.Match.AwayTeam.Name)
		fmt.Printf("   Status: %s\n", details.Status)
		if details.Match.HomeScore != nil && details.Match.AwayScore != nil {
			fmt.Printf("   Score: %d-%d\n", *details.Match.HomeScore, *details.Match.AwayScore)
		}
		fmt.Printf("   Events: %d events\n", len(details.Events))
	} else {
		fmt.Println("⚠️  No match details found")
	}
}

func showCachedMatches(cache *fotmob.ResponseCache) {
	cachedIDs := cache.CachedMatchIDs()

	if len(cachedIDs) == 0 {
		fmt.Println("No matches currently cached.")
		return
	}

	fmt.Printf("📋 Currently cached matches (%d total):\n\n", len(cachedIDs))

	for _, id := range cachedIDs {
		if details := cache.Details(id); details != nil {
			fmt.Printf("ID: %d | %s vs %s", id, details.Match.HomeTeam.Name, details.Match.AwayTeam.Name)

			if details.Match.HomeScore != nil && details.Match.AwayScore != nil {
				fmt.Printf(" | Score: %d-%d", *details.Match.HomeScore, *details.Match.AwayScore)
			}

			fmt.Printf(" | Status: %s", details.Status)

			if details.Match.MatchTime != nil {
				fmt.Printf(" | Date: %s", details.Match.MatchTime.Format("2006-01-02"))
			}

			fmt.Println()
		}
	}
}

func clearMatchesByTeam(cache *fotmob.ResponseCache, teamFilter string) {
	cachedIDs := cache.CachedMatchIDs()
	filterLower := strings.ToLower(teamFilter)
	var matchesToClear []int

	fmt.Printf("Searching for matches with team containing '%s'...\n", teamFilter)

	for _, id := range cachedIDs {
		if details := cache.Details(id); details != nil {
			homeTeamLower := strings.ToLower(details.Match.HomeTeam.Name)
			awayTeamLower := strings.ToLower(details.Match.AwayTeam.Name)

			if strings.Contains(homeTeamLower, filterLower) || strings.Contains(awayTeamLower, filterLower) {
				matchesToClear = append(matchesToClear, id)
				fmt.Printf("  Found: ID %d - %s vs %s\n", id, details.Match.HomeTeam.Name, details.Match.AwayTeam.Name)
			}
		}
	}

	if len(matchesToClear) == 0 {
		fmt.Println("No matches found matching that team name.")
		return
	}

	fmt.Printf("\nClearing cache for %d match(es)...\n", len(matchesToClear))
	for _, id := range matchesToClear {
		cache.ClearMatchDetails(id)
		fmt.Printf("✅ Cleared match ID: %d\n", id)
	}
}
