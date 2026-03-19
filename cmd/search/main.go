package main

import (
	"flag"
	"fmt"
	"os"
	"searx-cli/internal/engine"
	"searx-cli/internal/reader"
	"searx-cli/internal/ui"
	"searx-cli/internal/util"
	"strings"
)

var version = "dev"

var defaultSearxInstances = []string{
	"https://searx.be",
	"https://paulgo.io",
	"https://searx.org",
}

func main() {
	// Custom usage/help
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: search [options] [query]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  setup         Download and install the latest Lightpanda browser\n")
		fmt.Fprintf(os.Stderr, "  update        Force update the Lightpanda browser\n")
		fmt.Fprintf(os.Stderr, "  version       Show Search CLI and Lightpanda versions\n")
		fmt.Fprintf(os.Stderr, "  --version     Show Search CLI version only\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -e <engine>   Search engine: ddg, google, brave, mojeek, hn, searx\n")
		fmt.Fprintf(os.Stderr, "  -read <url>   Read full article content from the given URL\n")
		fmt.Fprintf(os.Stderr, "  -save         Save the read article to a markdown file\n")
		fmt.Fprintf(os.Stderr, "  -panda        Force use of Lightpanda for reading\n")
		fmt.Fprintf(os.Stderr, "  -archive      Use archive.today prefix for paywalls\n")
		fmt.Fprintf(os.Stderr, "  -hn <cat>     Hacker News category: top, new, best, ask, show, job\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  search \"golang news\"\n")
		fmt.Fprintf(os.Stderr, "  search -e hn -hn best\n")
		fmt.Fprintf(os.Stderr, "  search -read \"https://...\" -save\n")
	}

	engineFlag := flag.String("e", "searx", "Engine: ddg, google, brave, mojeek, hn, searx")
	instanceFlag := flag.String("i", "", "Searx instance URL (only for -e searx)")
	hnCatFlag := flag.String("hn", "top", "HN category: top, new, best, ask, show, job")
	readURL := flag.String("read", "", "URL to read article content from")
	archiveFlag := flag.Bool("archive", false, "Use archive.today to read the URL (for paywalls)")
	pandaFlag := flag.Bool("panda", false, "Use lightpanda headless browser for reading")
	saveFlag := flag.Bool("save", false, "Save the read article to a markdown file")
	versionFlag := flag.Bool("version", false, "Show Search CLI version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		return
	}

	// Handle special commands
	if flag.NArg() > 0 {
		switch strings.ToLower(flag.Arg(0)) {
		case "setup":
			if err := util.EnsureLightpanda(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "update":
			if err := util.UpdateLightpanda(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "version":
			fmt.Printf("Search CLI: %s\n", version)
			fmt.Printf("Lightpanda: %s\n", util.GetLocalLightpandaVersion())
			return
		}
	}

	if *readURL != "" {
		// Ensure Lightpanda is installed if likely needed
		_ = util.EnsureLightpanda()

		finalURL := *readURL
		if *archiveFlag {
			finalURL = "https://archive.today/" + finalURL
		}

		var article *reader.Article
		var err error

		if *pandaFlag {
			fmt.Printf("Forcing Lightpanda for: %s...\n", finalURL)
			article, err = reader.ReadURLWithLightpanda(finalURL)
		} else {
			article, err = reader.SmartRead(finalURL)
			isPaywall := strings.Contains(finalURL, "nytimes.com") || strings.Contains(finalURL, "wsj.com") || strings.Contains(finalURL, "bloomberg.com")
			if (err != nil || len(article.Content) < 100) && isPaywall && !*archiveFlag {
				fmt.Println("\nDetected a paywalled site. Automatically retrying via archive.today...")
				archiveURL := "https://archive.today/" + finalURL
				article, err = reader.SmartRead(archiveURL)
			}
		}

		if err != nil {
			fmt.Printf("\nError: All fetch methods failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nTITLE: %s\n", article.Title)
		fmt.Println(strings.Repeat("=", 60))
		fmt.Printf("\n%s\n", article.Content)

		if *saveFlag {
			filename := sanitizeFilename(article.Title) + ".md"
			content := fmt.Sprintf("# %s\n\nSource: %s\n\n%s", article.Title, finalURL, article.Content)
			err := os.WriteFile(filename, []byte(content), 0644)
			if err != nil {
				fmt.Printf("\nError saving file: %v\n", err)
			} else {
				fmt.Printf("\n[✔] Article saved to: %s\n", filename)
			}
		}
		return
	}

	query := strings.Join(flag.Args(), " ")

	if strings.ToLower(*engineFlag) != "hn" && query == "" {
		flag.Usage()
		os.Exit(1)
	}

	var searchEngine engine.SearchEngine

	switch strings.ToLower(*engineFlag) {
	case "google":
		searchEngine = &engine.GoogleEngine{}
	case "ddg":
		searchEngine = &engine.DuckDuckGoEngine{}
	case "brave":
		searchEngine = &engine.BraveEngine{}
	case "mojeek":
		searchEngine = &engine.MojeekEngine{}
	case "hn":
		searchEngine = &engine.HackerNewsEngine{Category: *hnCatFlag}
	case "searx":
		if *instanceFlag != "" {
			searchEngine = &engine.SearxEngine{InstanceURL: *instanceFlag}
		} else {
			for _, instance := range defaultSearxInstances {
				fmt.Printf("Trying Searx instance: %s...\n", instance)
				s := &engine.SearxEngine{InstanceURL: instance}
				results, err := s.Search(query)
				if err == nil && len(results) > 0 {
					ui.PrintResults(s.Name(), query, results)
					return
				}
			}
			fmt.Println("All default Searx instances failed.")
			return
		}
	default:
		fmt.Printf("Unknown engine: %s\n", *engineFlag)
		os.Exit(1)
	}

	results, err := searchEngine.Search(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	ui.PrintResults(searchEngine.Name(), query, results)
}

func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	res := result.String()
	for strings.Contains(res, "--") {
		res = strings.ReplaceAll(res, "--", "-")
	}
	return strings.Trim(res, "-")
}
