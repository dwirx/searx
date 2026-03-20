package engine

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"searx-cli/internal/types"
	"searx-cli/internal/util"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type GodEngine struct {
	Timeout time.Duration
}

func (g *GodEngine) Name() string {
	return "God Mode (Multi-Engine Anti-Bot)"
}

func (g *GodEngine) Search(query string) ([]types.Result, error) {
	lightpandaPath, err := util.LightpandaBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("God Mode requires Lightpanda: %v. Run `search setup` first", err)
	}

	engines := []struct {
		name string
		url  string
	}{
		{"Google", "https://www.google.com/search?q=" + url.QueryEscape(query)},
		{"Brave", "https://search.brave.com/search?q=" + url.QueryEscape(query)},
		{"DuckDuckGo", "https://duckduckgo.com/html/?q=" + url.QueryEscape(query)},
		{"Yahoo", "https://search.yahoo.com/search?p=" + url.QueryEscape(query)},
		{"Mojeek", "https://www.mojeek.com/search?q=" + url.QueryEscape(query)},
		{"Bing", "https://www.bing.com/search?q=" + url.QueryEscape(query)},
		{"Ask", "https://www.ask.com/web?q=" + url.QueryEscape(query)},
	}

	var wg sync.WaitGroup
	resultChan := make(chan []types.Result, len(engines))
	errChan := make(chan error, len(engines))

	for _, eng := range engines {
		wg.Add(1)
		go func(engName, engURL string) {
			defer wg.Done()
			results, err := g.fetchWithLightpanda(lightpandaPath, engName, engURL)
			if err != nil {
				errChan <- fmt.Errorf("%s error: %v", engName, err)
				return
			}
			resultChan <- results
		}(eng.name, eng.url)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	var allResults []types.Result
	seenURLs := make(map[string]bool)

	for res := range resultChan {
		for _, r := range res {
			// Basic normalization of URLs to avoid dupes
			cleanURL := strings.TrimRight(r.URL, "/")
			if !seenURLs[cleanURL] && r.Title != "" && r.URL != "" {
				allResults = append(allResults, r)
				seenURLs[cleanURL] = true
			}
		}
	}

	if len(allResults) == 0 {
		var errs []string
		for err := range errChan {
			errs = append(errs, err.Error())
		}
		if len(errs) > 0 {
			return nil, fmt.Errorf("all engines failed: %s", strings.Join(errs, "; "))
		}
		return nil, fmt.Errorf("no results found for: %s", query)
	}

	return allResults, nil
}

func (g *GodEngine) fetchWithLightpanda(lpPath, name, targetURL string) ([]types.Result, error) {
	// Use Lightpanda to fetch HTML
	// We add --user_agent_suffix to look slightly different if needed
	args := []string{"fetch", "--dump", "html", "--strip_mode", "js,css", "--http_timeout", "15000", targetURL}
	cmd := exec.Command(lpPath, args...)
	
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	// Set a reasonable timeout for each engine
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case <-time.After(20 * time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, fmt.Errorf("timeout")
	case err := <-done:
		if err != nil {
			return nil, err
		}
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		return nil, err
	}

	var results []types.Result
	switch name {
	case "Google":
		// Google standard selectors
		doc.Find("div.g, div.tF2Cxc").Each(func(i int, s *goquery.Selection) {
			title := s.Find("h3").First().Text()
			link, _ := s.Find("a").First().Attr("href")
			snippet := s.Find("div.VwiC3b, span.st").First().Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: title, URL: link, Snippet: snippet})
			}
		})
	case "Brave":
		doc.Find("div.snippet").Each(func(i int, s *goquery.Selection) {
			title := s.Find(".snippet-title, h3").First().Text()
			link, _ := s.Find("a").First().Attr("href")
			snippet := s.Find(".snippet-description, .snippet-content").First().Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: title, URL: link, Snippet: snippet})
			}
		})
	case "DuckDuckGo":
		doc.Find(".result").Each(func(i int, s *goquery.Selection) {
			title := s.Find(".result__title").First().Text()
			link, _ := s.Find(".result__url").First().Attr("href")
			snippet := s.Find(".result__snippet").First().Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: strings.TrimSpace(title), URL: strings.TrimSpace(link), Snippet: strings.TrimSpace(snippet)})
			}
		})
	case "Yahoo":
		doc.Find("div.dd.algo").Each(func(i int, s *goquery.Selection) {
			title := s.Find("h3.title").Text()
			link, _ := s.Find("a").First().Attr("href")
			snippet := s.Find(".compText, .algo-desc").Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: title, URL: link, Snippet: snippet})
			}
		})
	case "Mojeek":
		doc.Find("li.result").Each(func(i int, s *goquery.Selection) {
			title := s.Find("a.title").Text()
			link, _ := s.Find("a.title").Attr("href")
			snippet := s.Find("p.s").Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: title, URL: link, Snippet: snippet})
			}
		})
	case "Bing":
		doc.Find("li.b_algo").Each(func(i int, s *goquery.Selection) {
			title := s.Find("h2").Text()
			link, _ := s.Find("a").First().Attr("href")
			snippet := s.Find(".b_caption p, .b_snippet").Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: title, URL: link, Snippet: snippet})
			}
		})
	case "Ask":
		doc.Find("div.PartialSearchResults-item").Each(func(i int, s *goquery.Selection) {
			title := s.Find("a.PartialSearchResults-item-title-link").Text()
			link, _ := s.Find("a.PartialSearchResults-item-title-link").Attr("href")
			snippet := s.Find("p.PartialSearchResults-item-abstract").Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: title, URL: link, Snippet: snippet})
			}
		})
	}

	return results, nil
}
