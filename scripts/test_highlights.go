package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/test_highlights.go <match_id>")
		fmt.Println("Example: go run scripts/test_highlights.go 4813581")
		os.Exit(1)
	}

	matchIDStr := os.Args[1]

	// Fetch raw match details to inspect highlights
	url := fmt.Sprintf("https://www.fotmob.com/api/matchDetails?matchId=%s", matchIDStr)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	// Create HTTP client directly
	httpClient := &http.Client{Timeout: 15 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error fetching match details: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║            FotMob Highlight Video Inspector                 ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("Match ID: %s\n\n", matchIDStr)

	// Extract and display highlights data
	displayHighlights(rawResponse)

	// Also check for any other video-related fields
	findVideoFields(rawResponse, "")
}

func displayHighlights(data map[string]interface{}) {
	fmt.Printf("══════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("HIGHLIGHTS SECTION\n")
	fmt.Printf("══════════════════════════════════════════════════════════════════════\n\n")

	content, ok := data["content"].(map[string]interface{})
	if !ok {
		fmt.Println("❌ No content section found")
		return
	}

	matchFacts, ok := content["matchFacts"].(map[string]interface{})
	if !ok {
		fmt.Println("❌ No matchFacts section found")
		return
	}

	highlights, ok := matchFacts["highlights"].(map[string]interface{})
	if !ok {
		fmt.Println("❌ No highlights section found in matchFacts")
		return
	}

	fmt.Printf("✅ Found highlights section!\n\n")

	// Pretty print the highlights JSON
	highlightsJSON, err := json.MarshalIndent(highlights, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling highlights: %v\n", err)
		return
	}

	fmt.Printf("Raw highlights JSON:\n")
	fmt.Printf("%s\n\n", string(highlightsJSON))

	// Extract specific fields
	if url, ok := highlights["url"].(string); ok && url != "" {
		fmt.Printf("🎬 Video URL: %s\n", url)
	}

	if image, ok := highlights["image"].(string); ok && image != "" {
		fmt.Printf("🖼️  Thumbnail: %s\n", image)
	}

	if source, ok := highlights["source"].(string); ok && source != "" {
		fmt.Printf("📺 Source: %s\n", source)
	}
}

func findVideoFields(data interface{}, path string) {
	fmt.Printf("\n══════════════════════════════════════════════════════════════════════\n")
	fmt.Printf("SEARCHING FOR VIDEO-RELATED FIELDS\n")
	fmt.Printf("══════════════════════════════════════════════════════════════════════\n\n")

	videoFields := []string{
		"video", "videos", "highlight", "highlights", "youtube", "url", "media",
		"stream", "broadcast", "replay", "clip", "footage",
	}

	var foundFields []string

	findFieldsRecursive(data, path, videoFields, &foundFields)

	if len(foundFields) == 0 {
		fmt.Println("❌ No video-related fields found")
	} else {
		fmt.Printf("✅ Found %d video-related fields:\n\n", len(foundFields))
		for _, field := range foundFields {
			fmt.Printf("  🎥 %s\n", field)
		}
	}
}

func findFieldsRecursive(data interface{}, path string, searchTerms []string, found *[]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newPath := key
			if path != "" {
				newPath = path + "." + key
			}

			// Check if this key matches any search term
			for _, term := range searchTerms {
				if containsIgnoreCase(key, term) {
					*found = append(*found, newPath)
					break
				}
			}

			// Recursively search nested objects
			findFieldsRecursive(value, newPath, searchTerms, found)
		}
	case []interface{}:
		for i, item := range v {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			findFieldsRecursive(item, newPath, searchTerms, found)
		}
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) > len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		len(s) > 2*len(substr) && s[len(substr):len(s)-len(substr)] == s[len(substr):len(s)-len(substr)])
}

func toLower(s string) string {
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}
