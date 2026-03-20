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

type MojeekEngine struct{}

func (m *MojeekEngine) Name() string {
	return "Mojeek"
}

func (m *MojeekEngine) Search(query string) ([]types.Result, error) {
	u, _ := url.Parse("https://www.mojeek.com/search")
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
	// Updated Mojeek selectors
	doc.Find(".results-standard li").Each(func(i int, s *goquery.Selection) {
		if len(results) >= 10 {
			return
		}
		
		titleLink := s.Find("a.title")
		title := strings.TrimSpace(titleLink.Text())
		link, _ := titleLink.Attr("href")
		snippet := strings.TrimSpace(s.Find("p").Text())

		if title != "" && link != "" && !strings.HasPrefix(link, "/") {
			results = append(results, Result{
				Title:   title,
				URL:     link,
				Snippet: snippet,
			})
		}
	})

	return results, nil
}
