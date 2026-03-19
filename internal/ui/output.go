package ui

import (
	"fmt"
	"strings"
	"searx-cli/internal/engine"
)

func PrintResults(engineName string, query string, results []engine.Result) {
	fmt.Printf("\nResults for: %s (via %s)\n", query, engineName)
	fmt.Println(strings.Repeat("=", 60))
	
	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	for i, res := range results {
		title := res.Title
		if len(title) > 75 {
			title = title[:72] + "..."
		}
		fmt.Printf("[%d] %s\n", i+1, title)
		fmt.Printf("    %s\n", res.URL)
		if res.Snippet != "" {
			// Limit snippet length
			snippet := res.Snippet
			if len(snippet) > 150 {
				snippet = snippet[:147] + "..."
			}
			fmt.Printf("    %s\n", snippet)
		}
		fmt.Println()
	}
}
