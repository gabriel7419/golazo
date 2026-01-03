// find_league_id.go - Find FotMob league IDs by name or country
//
// Usage:
//   go run scripts/find_league_id.go <search_term>
//
// Examples:
//   go run scripts/find_league_id.go "Premier League"
//   go run scripts/find_league_id.go Japan
//   go run scripts/find_league_id.go Denmark

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/find_league_id.go <search_term>")
		fmt.Println("Example: go run scripts/find_league_id.go Japan")
		os.Exit(1)
	}

	term := strings.Join(os.Args[1:], " ")
	results := search(term)

	if len(results) == 0 {
		fmt.Println("No leagues found")
		os.Exit(0)
	}

	fmt.Printf("%-6s %-35s %s\n", "ID", "Name", "Country")
	fmt.Println(strings.Repeat("-", 55))
	for _, r := range results {
		fmt.Printf("%-6d %-35s %s\n", r.ID, truncate(r.Name, 35), r.Country)
	}
}

type league struct {
	ID      int
	Name    string
	Country string
}

func search(term string) []league {
	client := &http.Client{Timeout: 10 * time.Second}
	url := "https://www.fotmob.com/api/search/suggest?term=" + url.QueryEscape(term)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var data []struct {
		Suggestions []struct {
			Type    string `json:"type"`
			ID      string `json:"id"`
			Name    string `json:"name"`
			Country string `json:"ccode"`
		} `json:"suggestions"`
	}

	json.NewDecoder(resp.Body).Decode(&data)

	seen := make(map[int]bool)
	var results []league

	for _, group := range data {
		for _, s := range group.Suggestions {
			if s.Type != "league" {
				continue
			}
			var id int
			fmt.Sscanf(s.ID, "%d", &id)
			if seen[id] {
				continue
			}
			seen[id] = true
			results = append(results, league{ID: id, Name: s.Name, Country: s.Country})
		}
	}
	return results
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
