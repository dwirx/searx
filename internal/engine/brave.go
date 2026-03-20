package engine

import (
	"fmt"
	"net/http"
	"net/url"
	"searx-cli/internal/types"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type BraveEngine struct{}

func (b *BraveEngine) Name() string {
	return "Brave Search"
}

func (b *BraveEngine) Search(query string) ([]types.Result, error) {
	u, _ := url.Parse("https://search.brave.com/search")
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []Result
	doc.Find("div.snippet").Each(func(i int, s *goquery.Selection) {
		if len(results) >= 10 {
			return
		}
		title := strings.TrimSpace(s.Find(".snippet-title, h3").Text())
		link, _ := s.Find("a").First().Attr("href")
		snippet := strings.TrimSpace(s.Find(".snippet-description, .snippet-content").Text())

		if title != "" && link != "" {
			results = append(results, Result{
				Title:   title,
				URL:     link,
				Snippet: snippet,
			})
		}
	})

	return results, nil
}
