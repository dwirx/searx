package engine

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"searx-cli/internal/types"
	"searx-cli/internal/util"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type BingEngine struct{}

func (b *BingEngine) Name() string {
	return "Bing Search"
}

func (b *BingEngine) Search(query string) ([]types.Result, error) {
	// Use Lightpanda for Bing to bypass bot detection as standard HTTP fetch often fails
	lightpandaPath, err := util.LightpandaBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("Bing Search requires Lightpanda for anti-bot protection. Run `search setup` first")
	}

	targetURL := "https://www.bing.com/search?q=" + url.QueryEscape(query)
	
	// Fetch using Lightpanda
	args := []string{"fetch", "--dump", "html", "--strip_mode", "js,css", "--http_timeout", "15000", targetURL}
	cmd := exec.Command(lightpandaPath, args...)
	
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case <-time.After(20 * time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, fmt.Errorf("bing search timeout")
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("bing fetch error: %v", err)
		}
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		return nil, err
	}

	var results []types.Result
	doc.Find("li.b_algo").Each(func(i int, s *goquery.Selection) {
		if len(results) >= 15 {
			return
		}
		title := strings.TrimSpace(s.Find("h2").Text())
		link, _ := s.Find("a").First().Attr("href")
		snippet := strings.TrimSpace(s.Find(".b_caption p, .b_snippet").Text())

		if title != "" && link != "" && strings.HasPrefix(link, "http") {
			results = append(results, types.Result{
				Title:   title,
				URL:     link,
				Snippet: snippet,
			})
		}
	})

	if len(results) == 0 {
		return nil, fmt.Errorf("no results found on bing")
	}

	return results, nil
}
