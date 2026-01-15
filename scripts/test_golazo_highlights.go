package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/0xjuanma/golazo/internal/api"
	"github.com/0xjuanma/golazo/internal/fotmob"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("🔬 Golazo Highlights Integration Test")
		fmt.Println("")
		fmt.Println("Usage: go run scripts/test_golazo_highlights.go <match_id>")
		fmt.Println("Example: go run scripts/test_golazo_highlights.go 4803233")
		fmt.Println("")
		fmt.Println("This tool tests the complete golazo highlights pipeline:")
		fmt.Println("  1. Raw API response")
		fmt.Println("  2. Golazo parsing")
		fmt.Println("  3. MatchDetails structure")
		fmt.Println("  4. UI display logic simulation")
		os.Exit(1)
	}

	matchIDStr := os.Args[1]
	var matchID int
	fmt.Sscanf(matchIDStr, "%d", &matchID)

	fmt.Printf("🔬 Testing golazo highlights pipeline for match ID: %s\n\n", matchIDStr)

	// 1. Check raw API response
	fmt.Println("1️⃣  Raw API Response Analysis...")
	rawHighlights := checkRawAPI(matchIDStr)

	// 2. Test golazo parsing
	fmt.Println("\n2️⃣  Golazo FotMob Client Parsing...")
	parsedHighlights := testGolazoParsing(matchID)

	// 3. Compare results
	fmt.Println("\n3️⃣  Comparison & Analysis...")
	compareResults(rawHighlights, parsedHighlights)

	// 4. Simulate UI display
	fmt.Println("\n4️⃣  UI Display Simulation...")
	simulateUIDisplay(parsedHighlights)
}

func checkRawAPI(matchIDStr string) map[string]interface{} {
	url := fmt.Sprintf("https://www.fotmob.com/api/matchDetails?matchId=%s", matchIDStr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ Error creating request: %v\n", err)
		return nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Golazo/1.0)")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error fetching: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Status code: %d\n", resp.StatusCode)
		return nil
	}

	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		fmt.Printf("❌ Error decoding: %v\n", err)
		return nil
	}

	// Extract highlights data
	content, ok := rawResponse["content"].(map[string]interface{})
	if !ok {
		fmt.Println("❌ No content in response")
		return nil
	}

	matchFacts, ok := content["matchFacts"].(map[string]interface{})
	if !ok {
		fmt.Println("❌ No matchFacts in content")
		return nil
	}

	highlights, exists := matchFacts["highlights"]
	if !exists {
		fmt.Println("❌ No highlights field in matchFacts")
		return nil
	}

	if highlights == nil {
		fmt.Println("⚠️  Highlights field is null")
		return nil
	}

	highlightsObj, ok := highlights.(map[string]interface{})
	if !ok {
		fmt.Printf("❌ Highlights not an object (type: %T)\n", highlights)
		return nil
	}

	fmt.Println("✅ Found highlights object in raw API")

	// Display raw highlights data
	if url, ok := highlightsObj["url"].(string); ok && url != "" {
		fmt.Printf("   Raw URL: %s\n", url)
	}
	if image, ok := highlightsObj["image"].(string); ok && image != "" {
		fmt.Printf("   Raw Image: %s\n", image)
	}
	if source, ok := highlightsObj["source"].(string); ok && source != "" {
		fmt.Printf("   Raw Source: %s\n", source)
	}

	return highlightsObj
}

func testGolazoParsing(matchID int) *api.MatchHighlight {
	client := fotmob.NewClient()
	ctx := context.Background()

	details, err := client.MatchDetails(ctx, matchID)
	if err != nil {
		fmt.Printf("❌ Golazo parsing error: %v\n", err)
		return nil
	}

	if details == nil {
		fmt.Println("❌ Golazo returned nil details")
		return nil
	}

	if details.Highlight == nil {
		fmt.Println("❌ Golazo parsed details but Highlight is nil")
		return nil
	}

	fmt.Println("✅ Golazo successfully parsed highlight data")
	fmt.Printf("   Parsed URL: %s\n", details.Highlight.URL)
	if details.Highlight.Source != "" {
		fmt.Printf("   Parsed Source: %s\n", details.Highlight.Source)
	}
	if details.Highlight.Image != "" {
		fmt.Printf("   Parsed Image: %s\n", details.Highlight.Image)
	}

	return details.Highlight
}

func compareResults(rawHighlights map[string]interface{}, parsedHighlights *api.MatchHighlight) {
	if rawHighlights == nil && parsedHighlights == nil {
		fmt.Println("✅ Both raw API and golazo show no highlights")
		return
	}

	if rawHighlights == nil && parsedHighlights != nil {
		fmt.Println("❌ MISMATCH: Raw API has no highlights but golazo parsed some")
		return
	}

	if rawHighlights != nil && parsedHighlights == nil {
		fmt.Println("❌ MISMATCH: Raw API has highlights but golazo parsed none")
		fmt.Println("💡 This indicates a PARSING BUG in golazo")
		return
	}

	// Both have highlights, compare fields
	fmt.Println("✅ Both raw API and golazo have highlights")

	rawURL, _ := rawHighlights["url"].(string)
	if rawURL != parsedHighlights.URL {
		fmt.Printf("❌ URL MISMATCH:\n")
		fmt.Printf("   Raw: %s\n", rawURL)
		fmt.Printf("   Parsed: %s\n", parsedHighlights.URL)
	}

	rawSource, _ := rawHighlights["source"].(string)
	if rawSource != parsedHighlights.Source {
		fmt.Printf("❌ Source MISMATCH:\n")
		fmt.Printf("   Raw: %s\n", rawSource)
		fmt.Printf("   Parsed: %s\n", parsedHighlights.Source)
	}

	rawImage, _ := rawHighlights["image"].(string)
	if rawImage != parsedHighlights.Image {
		fmt.Printf("❌ Image MISMATCH:\n")
		fmt.Printf("   Raw: %s\n", rawImage)
		fmt.Printf("   Parsed: %s\n", parsedHighlights.Image)
	}
}

func simulateUIDisplay(highlight *api.MatchHighlight) {
	if highlight == nil {
		fmt.Println("❌ No highlight data to display")
		fmt.Println("💡 UI would show no highlights section")
		return
	}

	if highlight.URL == "" {
		fmt.Println("❌ Highlight has no URL")
		fmt.Println("💡 UI condition 'details.Highlight.URL != \"\"' would be false")
		return
	}

	fmt.Println("✅ UI would display highlights section")
	fmt.Printf("   Would show: 🎥 HIGHLIGHTS AVAILABLE (%s)\n", highlight.Source)
	fmt.Println("💡 If you don't see this in the app, there might be a UI bug")
	fmt.Println("   Check if the match details view is being rendered correctly")
}
