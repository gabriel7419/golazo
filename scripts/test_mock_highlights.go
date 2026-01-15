package main

import (
	"fmt"

	"github.com/0xjuanma/golazo/internal/data"
)

func main() {
	fmt.Println("Testing highlights in mock data...")

	// Test match IDs that should have highlights
	testMatches := []int{1001, 1005, 1002}

	for _, matchID := range testMatches {
		details, err := data.MockFinishedMatchDetails(matchID)
		if err != nil {
			fmt.Printf("Error getting details for match %d: %v\n", matchID, err)
			continue
		}

		if details == nil {
			fmt.Printf("Match %d: No details found\n", matchID)
			continue
		}

		fmt.Printf("Match %d: %s vs %s\n", matchID, details.HomeTeam.Name, details.AwayTeam.Name)

		if details.Highlight != nil {
			fmt.Printf("  ✅ Has highlight: %s\n", details.Highlight.Title)
			fmt.Printf("  URL: %s\n", details.Highlight.URL)
		} else {
			fmt.Printf("  ❌ No highlight\n")
		}
		fmt.Println()
	}
}
