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
			if !seenURLs[r.URL] {
				allResults = append(allResults, r)
				seenURLs[r.URL] = true
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
		return nil, fmt.Errorf("no results found")
	}

	return allResults, nil
}

func (g *GodEngine) fetchWithLightpanda(lpPath, name, targetURL string) ([]types.Result, error) {
	// Use Lightpanda to fetch HTML
	args := []string{"fetch", "--dump", "html", "--strip_mode", "js,css", targetURL}
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
		doc.Find("div.g").Each(func(i int, s *goquery.Selection) {
			title := s.Find("h3").Text()
			link, _ := s.Find("a").Attr("href")
			snippet := s.Find("div.VwiC3b").Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: title, URL: link, Snippet: snippet})
			}
		})
	case "Brave":
		doc.Find("div.snippet").Each(func(i int, s *goquery.Selection) {
			title := s.Find(".snippet-title, h3").Text()
			link, _ := s.Find("a").First().Attr("href")
			snippet := s.Find(".snippet-description, .snippet-content").Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: title, URL: link, Snippet: snippet})
			}
		})
	case "DuckDuckGo":
		doc.Find(".result").Each(func(i int, s *goquery.Selection) {
			title := s.Find(".result__title").Text()
			link, _ := s.Find(".result__url").Attr("href")
			snippet := s.Find(".result__snippet").Text()
			if title != "" && link != "" {
				results = append(results, types.Result{Title: strings.TrimSpace(title), URL: strings.TrimSpace(link), Snippet: strings.TrimSpace(snippet)})
			}
		})
	}

	return results, nil
}
