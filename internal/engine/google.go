package engine

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type GoogleEngine struct{}

func (g *GoogleEngine) Name() string {
	return "Google"
}

func (g *GoogleEngine) Search(query string) ([]Result, error) {
	// Using more robust URL and params
	u, _ := url.Parse("https://www.google.com/search")
	q := u.Query()
	q.Set("q", query)
	q.Set("gbv", "1") // No Javascript version (Old but reliable for scraping)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", u.String(), nil)
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google returned %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []Result
	// Selectors for GBV=1 mode
	doc.Find("div.ZINbbc").Each(func(i int, s *goquery.Selection) {
		if len(results) >= 10 {
			return
		}
		
		// Each result usually has a title in an h3
		title := s.Find("h3").Text()
		if title == "" {
			return
		}

		linkNode := s.Find("a").First()
		href, _ := linkNode.Attr("href")
		
		// Google links in GBV=1 are often /url?q=...
		cleanURL := strings.TrimPrefix(href, "/url?q=")
		if idx := strings.Index(cleanURL, "&"); idx != -1 {
			cleanURL = cleanURL[:idx]
		}
		
		// Snippet is usually in a div with specific class or just text after title
		snippet := s.Find(".VwiC3b, .BNeawe.s3v9rd.AP7Wnd").First().Text()

		if cleanURL != "" && !strings.HasPrefix(cleanURL, "/") {
			results = append(results, Result{
				Title:   title,
				URL:     cleanURL,
				Snippet: snippet,
			})
		}
	})

	return results, nil
}
