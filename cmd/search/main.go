package main

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"searx-cli/internal/engine"
	"searx-cli/internal/reader"
	"searx-cli/internal/ui"
	"searx-cli/internal/util"
)

const defaultSearxInstance = "https://searx.be"

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

func main() {
	engineName := flag.String("e", "ddg", "Search engine: ddg|google|brave|mojeek|hn|searx")
	readURL := flag.String("read", "", "URL of article to read")
	save := flag.Bool("save", false, "Save article to markdown")
	forcePanda := flag.Bool("panda", false, "Force Lightpanda browser fetch")
	forceArchive := flag.Bool("archive", false, "Force archive.today prefix")
	hnCategory := flag.String("hn", "top", "HN category: top|new|best|ask|show|job")
	searxInstance := flag.String("i", defaultSearxInstance, "Custom Searx instance URL")
	flag.Parse()

	args := flag.Args()
	if *readURL == "" && len(args) == 1 {
		switch strings.ToLower(strings.TrimSpace(args[0])) {
		case "setup":
			exitIfErr(runSetup())
			return
		case "update":
			exitIfErr(util.UpdateLightpanda())
			return
		}
	}

	if *readURL != "" {
		exitIfErr(runReadMode(*readURL, *save, *forcePanda, *forceArchive))
		return
	}

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	query := strings.Join(args, " ")
	exitIfErr(runSearchMode(*engineName, *hnCategory, *searxInstance, query))
}

func runSetup() error {
	if err := util.EnsureLightpanda(); err != nil {
		return err
	}
	fmt.Printf("Lightpanda: %s\n", util.GetLocalLightpandaVersion())
	return nil
}

func runSearchMode(engineName, hnCategory, searxInstance, query string) error {
	e, err := selectEngine(engineName, hnCategory, searxInstance)
	if err != nil {
		return err
	}

	results, err := e.Search(query)
	if err != nil {
		return err
	}

	ui.PrintResults(e.Name(), query, results)
	return nil
}

func runReadMode(rawURL string, save, forcePanda, forceArchive bool) error {
	targetURL := buildReaderURL(rawURL, forceArchive)

	var (
		article *reader.Article
		err     error
	)

	if forcePanda {
		article, err = reader.ReadURLWithLightpanda(targetURL)
	} else {
		article, err = reader.SmartRead(targetURL)
	}
	if err != nil {
		return err
	}

	title := strings.TrimSpace(article.Title)
	if title == "" {
		title = targetURL
	}

	fmt.Printf("# %s\n\n%s\n", title, strings.TrimSpace(article.Content))

	if save {
		path, saveErr := saveArticleMarkdown(article)
		if saveErr != nil {
			return saveErr
		}
		fmt.Printf("[✔] Article saved to: %s\n", path)
	}

	return nil
}

func selectEngine(engineName, hnCategory, searxInstance string) (engine.SearchEngine, error) {
	switch strings.ToLower(strings.TrimSpace(engineName)) {
	case "", "ddg":
		return &engine.DuckDuckGoEngine{}, nil
	case "google":
		return &engine.GoogleEngine{}, nil
	case "brave":
		return &engine.BraveEngine{}, nil
	case "mojeek":
		return &engine.MojeekEngine{}, nil
	case "hn":
		category := strings.TrimSpace(hnCategory)
		if category == "" {
			category = "top"
		}
		return &engine.HackerNewsEngine{Category: category}, nil
	case "searx":
		instance := strings.TrimSpace(searxInstance)
		if instance == "" {
			instance = defaultSearxInstance
		}
		return &engine.SearxEngine{InstanceURL: strings.TrimRight(instance, "/")}, nil
	default:
		return nil, fmt.Errorf("unknown engine %q", engineName)
	}
}

func buildReaderURL(rawURL string, forceArchive bool) string {
	cleanURL := strings.TrimSpace(rawURL)
	if cleanURL == "" || isArchiveURL(cleanURL) {
		return cleanURL
	}
	if forceArchive || isPaywalledDomain(cleanURL) {
		return toArchiveURL(cleanURL)
	}
	return cleanURL
}

func toArchiveURL(rawURL string) string {
	if isArchiveURL(rawURL) {
		return rawURL
	}
	return "https://archive.today/" + strings.TrimSpace(rawURL)
}

func isArchiveURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return strings.HasPrefix(host, "archive.")
}

func isPaywalledDomain(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false
	}

	known := []string{
		"nytimes.com",
		"wsj.com",
		"bloomberg.com",
		"ft.com",
		"economist.com",
	}

	for _, d := range known {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

func saveArticleMarkdown(article *reader.Article) (string, error) {
	if article == nil {
		return "", errors.New("article is nil")
	}

	title := strings.TrimSpace(article.Title)
	if title == "" {
		title = "Article"
	}

	filename := sanitizeFileName(title) + ".md"
	content := fmt.Sprintf("# %s\n\n%s\n", title, strings.TrimSpace(article.Content))
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		return "", err
	}
	return filename, nil
}

func sanitizeFileName(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "article"
	}
	return s
}

func exitIfErr(err error) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  search [flags] <query>\n")
	fmt.Fprintf(os.Stderr, "  search -read <url> [-save] [-panda] [-archive]\n")
	fmt.Fprintf(os.Stderr, "  search setup | update\n\n")
	flag.PrintDefaults()
}
