package ui

import (
	"fmt"
	"strings"
	"time"
	"searx-cli/internal/types"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorBold   = "\033[1m"
)

type Spinner struct {
	stopChan chan bool
	message  string
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		stopChan: make(chan bool),
		message:  message,
	}
}

func (s *Spinner) Start() {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	go func() {
		for {
			for _, frame := range frames {
				select {
				case <-s.stopChan:
					return
				default:
					fmt.Printf("\r%s%s%s %s", colorBlue, frame, colorReset, s.message)
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.stopChan <- true
	fmt.Print("\r\033[K") // Clear the line
}

func PrintResults(engineName string, query string, results []types.Result) {
	// Specialized printers for better UX
	if strings.Contains(strings.ToLower(engineName), "polymarket") {
		PrintPolymarketResults(engineName, query, results)
		return
	}
	if strings.Contains(strings.ToLower(engineName), "rss") {
		PrintRSSResults(engineName, query, results)
		return
	}

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
		fmt.Printf("[%d] %s%s%s\n", i+1, colorBold, title, colorReset)
		fmt.Printf("    %s\n", res.URL)
		if res.Snippet != "" {
			snippet := res.Snippet
			if len(snippet) > 150 {
				snippet = snippet[:147] + "..."
			}
			fmt.Printf("    %s\n", snippet)
		}
		fmt.Println()
	}
}

func PrintPolymarketResults(engineName string, query string, results []types.Result) {
	fmt.Printf("\n%s📊 %s %s (via %s)\n", colorBold, query, colorReset, engineName)
	fmt.Println(strings.Repeat("━", 60))
	
	if len(results) == 0 {
		fmt.Println("No active markets found for this category.")
		return
	}

	for i, res := range results {
		title := res.Title
		if strings.HasPrefix(title, "[") {
			endIdx := strings.Index(title, "]")
			if endIdx != -1 {
				tag := title[:endIdx+1]
				rest := title[endIdx+1:]
				title = fmt.Sprintf("%s%s%s%s", colorBlue, tag, colorReset, rest)
			}
		}

		fmt.Printf("%s%d.%s %s\n", colorBold, i+1, colorReset, title)
		
		if res.Snippet != "" {
			snippet := res.Snippet
			if strings.Contains(snippet, "[UP]") {
				snippet = strings.Replace(snippet, "[UP]", colorGreen+"↑"+colorReset, 1)
				snippet = colorGreen + snippet + colorReset
			} else if strings.Contains(snippet, "[DOWN]") {
				snippet = strings.Replace(snippet, "[DOWN]", colorRed+"↓"+colorReset, 1)
				snippet = colorRed + snippet + colorReset
			} else if strings.Contains(snippet, "[MOVE]") {
				snippet = strings.Replace(snippet, "[MOVE]", colorYellow+"•"+colorReset, 1)
			}
			fmt.Printf("   %s\n", snippet)
		}
		fmt.Printf("   %s🔗 %s%s\n\n", colorBlue, res.URL, colorReset)
	}
}

func PrintRSSResults(engineName string, query string, results []types.Result) {
	fmt.Printf("\n%s📰 %s %s (via %s)\n", colorBold, query, colorReset, engineName)
	fmt.Println(strings.Repeat("━", 60))
	
	if len(results) == 0 {
		fmt.Println("No news found for this search.")
		return
	}

	for i, res := range results {
		title := res.Title
		// Color source tag
		if strings.HasPrefix(title, "[") {
			endIdx := strings.Index(title, "]")
			if endIdx != -1 {
				source := title[:endIdx+1]
				titleRest := title[endIdx+1:]
				title = fmt.Sprintf("%s%s%s%s", colorMagenta, source, colorReset, colorBold+titleRest+colorReset)
			}
		}

		fmt.Printf("%s%d.%s %s\n", colorBold, i+1, colorReset, title)
		
		if res.Snippet != "" {
			// Clean up any remaining junk and indent
			snippet := strings.TrimSpace(res.Snippet)
			if len(snippet) > 250 {
				snippet = snippet[:247] + "..."
			}
			fmt.Printf("   %s\n", snippet)
		}
		fmt.Printf("   %s🔗 %s%s\n\n", colorCyan, res.URL, colorReset)
	}
}
