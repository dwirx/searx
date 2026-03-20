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
	colorWhite   = "\033[37m"
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
	if strings.Contains(strings.ToLower(engineName), "pasal") {
		PrintPasalResults(engineName, query, results)
		return
	}
	if strings.Contains(strings.ToLower(engineName), "exa") {
		PrintExaResults(engineName, query, results)
		return
	}
	if strings.Contains(strings.ToLower(engineName), "firecrawl") {
		PrintFirecrawlResults(engineName, query, results)
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
			snippet := strings.TrimSpace(res.Snippet)
			if len(snippet) > 250 {
				snippet = snippet[:247] + "..."
			}
			fmt.Printf("   %s\n", snippet)
		}
		fmt.Printf("   %s🔗 %s%s\n\n", colorCyan, res.URL, colorReset)
	}
}

func PrintPasalResults(engineName string, query string, results []types.Result) {
	fmt.Printf("\n%s🏛️  RI Law Search: %s %s (via %s)\n", colorBold, query, colorReset, engineName)
	fmt.Println(strings.Repeat("━", 60))
	
	if len(results) == 0 {
		fmt.Println("Tidak ada pasal atau undang-undang yang ditemukan.")
		return
	}

	for i, res := range results {
		title := res.Title
		snippet := strings.TrimSpace(res.Snippet)
		
		// Extract Status from snippet prefix
		statusColor := colorWhite
		statusText := "BERLAKU"
		if strings.HasPrefix(snippet, "[STATUS:") {
			endIdx := strings.Index(snippet, "]")
			if endIdx != -1 {
				statusText = snippet[8:endIdx]
				snippet = strings.TrimSpace(snippet[endIdx+1:])
				switch statusText {
				case "BERLAKU":
					statusColor = colorGreen
				case "DICABUT":
					statusColor = colorRed
				case "DIUBAH":
					statusColor = colorYellow
				}
			}
		}

		// Handle [UNVERIFIED] tag
		verifiedText := ""
		if strings.HasPrefix(snippet, "[UNVERIFIED]") {
			verifiedText = fmt.Sprintf(" %s[UNVERIFIED]%s", colorRed, colorReset)
			snippet = strings.TrimSpace(strings.TrimPrefix(snippet, "[UNVERIFIED]"))
		}

		// Color the type tag [UU], [PP], etc.
		if strings.HasPrefix(title, "Pasal") {
			parts := strings.SplitN(title, "|", 2)
			if len(parts) == 2 {
				pasalPart := strings.TrimSpace(parts[0])
				titlePart := strings.TrimSpace(parts[1])
				if strings.HasPrefix(titlePart, "[") {
					idx := strings.Index(titlePart, "]")
					tag := titlePart[:idx+1]
					rest := titlePart[idx+1:]
					title = fmt.Sprintf("%s%s%s %s%s%s %s", colorGreen, pasalPart, colorReset, colorMagenta, tag, colorReset, colorBold+rest+colorReset)
				} else {
					title = fmt.Sprintf("%s%s%s | %s", colorGreen, pasalPart, colorReset, colorBold+titlePart+colorReset)
				}
			}
		} else if strings.HasPrefix(title, "[") {
			idx := strings.Index(title, "]")
			tag := title[:idx+1]
			rest := title[idx+1:]
			title = fmt.Sprintf("%s%s%s %s", colorMagenta, tag, colorReset, colorBold+rest+colorReset)
		}

		fmt.Printf("%s%d.%s %s\n", colorBold, i+1, colorReset, title)
		fmt.Printf("   Status: %s%s%s%s\n", statusColor, statusText, colorReset, verifiedText)
		
		if snippet != "" {
			lines := strings.Split(snippet, "\n")
			for _, line := range lines {
				if len(line) > 120 {
					line = line[:117] + "..."
				}
				fmt.Printf("   %s%s%s\n", colorWhite, line, colorReset)
			}
		}
		fmt.Printf("   %s🔗 %s%s\n\n", colorCyan, res.URL, colorReset)
	}
}

func PrintExaResults(engineName string, query string, results []types.Result) {
	fmt.Printf("\n%s🚀 %s %s (via %s)\n", colorBold, query, colorReset, engineName)
	fmt.Println(strings.Repeat("━", 60))
	
	if len(results) == 0 {
		fmt.Println("No results found in Exa's neural network.")
		return
	}

	for i, res := range results {
		title := res.Title
		fmt.Printf("%s%d.%s %s\n", colorBold, i+1, colorReset, colorBold+title+colorReset)
		
		snippet := res.Snippet
		meta := ""
		highlights := ""
		
		if strings.Contains(snippet, " | ") {
			parts := strings.SplitN(snippet, " | ", 2)
			meta = parts[0]
			highlights = parts[1]
		} else {
			highlights = snippet
		}

		if meta != "" {
			fmt.Printf("   %s%s%s\n", colorCyan, meta, colorReset)
		}

		if highlights != "" {
			lines := strings.Split(highlights, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" { continue }
				if len(line) > 150 {
					line = line[:147] + "..."
				}
				fmt.Printf("   %s%s%s\n", colorWhite, line, colorReset)
			}
		}
		
		fmt.Printf("   %s🔗 %s%s%s\n\n", colorBlue, colorCyan, res.URL, colorReset)
	}
}

func PrintFirecrawlResults(engineName string, query string, results []types.Result) {
	fmt.Printf("\n%s🔥 %s %s (via %s)\n", colorBold, query, colorReset, engineName)
	fmt.Println(strings.Repeat("━", 60))
	
	if len(results) == 0 {
		fmt.Println("No results found in Firecrawl's web index.")
		return
	}

	for i, res := range results {
		title := res.Title
		fmt.Printf("%s%d.%s %s\n", colorBold, i+1, colorReset, colorBold+title+colorReset)
		
		snippet := res.Snippet
		if snippet != "" {
			lines := strings.Split(snippet, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" { continue }
				if len(line) > 150 {
					line = line[:147] + "..."
				}
				fmt.Printf("   %s%s%s\n", colorWhite, line, colorReset)
			}
		}
		
		fmt.Printf("   %s🔗 %s%s%s\n\n", colorBlue, colorCyan, res.URL, colorReset)
	}
}
