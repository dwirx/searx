package engine

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"searx-cli/internal/types"
	"searx-cli/internal/util"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type rssFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel channel  `xml:"channel"`
}

type channel struct {
	Title string `xml:"title"`
	Items []item  `xml:"item"`
}

type item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
}

type RSSEngine struct {
	Feeds        map[string]string // Name -> URL
	FilterSource string            // If non-empty, only fetch this source
}

func (r *RSSEngine) Name() string {
	if r.FilterSource != "" {
		return "RSS (" + r.FilterSource + ")"
	}
	return "RSS (All Feeds)"
}

func (r *RSSEngine) Search(query string) ([]types.Result, error) {
	var results []types.Result
	var mu sync.Mutex
	var wg sync.WaitGroup

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			MaxIdleConnsPerHost: 10,
		},
	}

	filterLower := strings.ToLower(r.FilterSource)

	for name, url := range r.Feeds {
		if r.FilterSource != "" && !strings.Contains(strings.ToLower(name), filterLower) {
			continue
		}

		wg.Add(1)
		go func(n, u string) {
			defer wg.Done()
			
			var items []types.Result
			var err error

			// Specialized handling for Bloomberg Asia (Direct Scraping via Lightpanda)
			if n == "bloomberg_asia" {
				items, err = r.fetchBloombergAsiaStealth(u, query)
			} else if strings.Contains(strings.ToLower(n), "bloomberg") || 
					 strings.Contains(strings.ToLower(n), "reuters") {
				items, err = r.fetchWithPandaEngine(u, n, query)
			} else {
				items, err = r.fetchGenericFeed(client, u, n, query)
				if (err != nil || len(items) == 0) && (strings.Contains(strings.ToLower(n), "nyt") || strings.Contains(strings.ToLower(n), "ft")) {
					items, err = r.fetchWithPandaEngine(u, n, query)
				}
			}

			if err == nil {
				mu.Lock()
				results = append(results, items...)
				mu.Unlock()
			}
		}(name, url)
	}

	wg.Wait()
	return results, nil
}

func (r *RSSEngine) fetchBloombergAsiaStealth(urlStr, query string) ([]types.Result, error) {
	lightpandaPath, err := util.LightpandaBinaryPath()
	if err != nil {
		return nil, err
	}

	// Fetch without stripping JS to get full rendered content
	cmd := exec.Command(lightpandaPath, "fetch", "--dump", "html", urlStr)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		return nil, err
	}

	var results []types.Result
	
	// Bloomberg often has article info in a big JSON block or specific tags
	// We'll try common selectors for their latest news
	doc.Find("article, div[data-component='story']").Each(func(i int, s *goquery.Selection) {
		if len(results) >= 15 { return }
		
		title := strings.TrimSpace(s.Find("h3, h2, a[href*='/articles/']").First().Text())
		link, _ := s.Find("a[href*='/articles/']").First().Attr("href")
		summary := strings.TrimSpace(s.Find("p, div[class*='summary']").First().Text())

		if title == "" || link == "" { return }
		
		if !strings.HasPrefix(link, "http") {
			link = "https://www.bloomberg.com" + link
		}

		if r.matchQuery(title, summary, query) {
			results = append(results, types.Result{
				Title: "[Bloomberg Asia] " + title,
				URL: link,
				Snippet: summary,
			})
		}
	})

	// Fallback to simple link scraper if articles not found in components
	if len(results) == 0 {
		return r.parseDocument(doc, "bloomberg_asia", query), nil
	}

	return results, nil
}

func (r *RSSEngine) fetchWithPandaEngine(urlStr, feedName, query string) ([]types.Result, error) {
	lightpandaPath, err := util.LightpandaBinaryPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(lightpandaPath, "fetch", "--dump", "html", urlStr)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		return nil, err
	}

	return r.parseDocument(doc, feedName, query), nil
}

func (r *RSSEngine) parseDocument(doc *goquery.Document, feedName, query string) []types.Result {
	var results []types.Result
	
	items := doc.Find("item")
	if items.Length() == 0 {
		items = doc.Find("entry")
	}
	if items.Length() == 0 {
		items = doc.Find("channel > item")
	}
	
	if items.Length() > 0 {
		items.Each(func(i int, s *goquery.Selection) {
			if len(results) >= 20 { return }
			title := cleanString(s.Find("title").First().Text())
			
			// Try link tag, then guid, then link[href]
			link := s.Find("link").First().Text()
			if link == "" {
				link = s.Find("guid").First().Text()
			}
			if link == "" { 
				link, _ = s.Find("link").First().Attr("href") 
			}

			desc := cleanString(s.Find("description").First().Text())
			if desc == "" { desc = cleanString(s.Find("summary").First().Text()) }

			if r.matchQuery(title, desc, query) {
				if len(desc) > 250 { desc = desc[:247] + "..." }
				results = append(results, types.Result{
					Title: fmt.Sprintf("[%s] %s", strings.Title(feedName), title),
					URL: link,
					Snippet: desc,
				})
			}
		})
	} else {
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if len(results) >= 15 { return }
			title := strings.TrimSpace(s.Text())
			href, _ := s.Attr("href")
			isNewsLink := strings.Contains(href, "/articles/") || strings.Contains(href, "/news/")
			
			if len(title) < 25 || !isNewsLink { return } 
			
			if r.matchQuery(title, "", query) {
				fullURL := href
				if !strings.HasPrefix(fullURL, "http") {
					if strings.Contains(feedName, "bloomberg") {
						fullURL = "https://www.bloomberg.com" + href
					} else {
						return
					}
				}
				
				isDup := false
				for _, res := range results {
					if res.URL == fullURL { isDup = true; break }
				}
				if isDup { return }

				results = append(results, types.Result{
					Title: fmt.Sprintf("[%s] %s", strings.Title(feedName), title),
					URL: fullURL,
					Snippet: "Latest update from " + feedName + " (Scraped)",
				})
			}
		})
	}
	return results
}

func (r *RSSEngine) matchQuery(title, desc, query string) bool {
	if query == "" { return true }
	q := strings.ToLower(query)
	return strings.Contains(strings.ToLower(title), q) || strings.Contains(strings.ToLower(desc), q)
}

func (r *RSSEngine) Validate() map[string]error {
	broken := make(map[string]error)
	var mu sync.Mutex
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 5 * time.Second}

	for name, url := range r.Feeds {
		wg.Add(1)
		go func(n, u string) {
			defer wg.Done()
			
			if n == "bloomberg_asia" { return } // Skip head check for web scraper

			resp, err := client.Head(u)
			if err != nil {
				resp, err = client.Get(u)
			}
			if err != nil {
				mu.Lock()
				broken[n] = err
				mu.Unlock()
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusNotFound {
				mu.Lock()
				broken[n] = fmt.Errorf("404 not found")
				mu.Unlock()
			}
		}(name, url)
	}
	wg.Wait()
	return broken
}

var cdataRegex = regexp.MustCompile(`<!\[CDATA\[(.*?)\]\]>`)

func cleanString(s string) string {
	s = strings.TrimSpace(s)
	if matches := cdataRegex.FindStringSubmatch(s); len(matches) > 1 {
		s = matches[1]
	}
	s = strings.ReplaceAll(s, "<![CDATA[", "")
	s = strings.ReplaceAll(s, "]]>", "")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err == nil {
		s = doc.Text()
	}
	return strings.TrimSpace(s)
}

func (r *RSSEngine) fetchGenericFeed(client *http.Client, urlStr, feedName, query string) ([]types.Result, error) {
	req, _ := http.NewRequest("GET", urlStr, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("status %d", resp.StatusCode) }
	
	bodyBytes, _ := io.ReadAll(resp.Body)

	// Try standard XML parsing first
	var feed rssFeed
	if err := xml.Unmarshal(bodyBytes, &feed); err == nil && len(feed.Channel.Items) > 0 {
		var results []types.Result
		for _, it := range feed.Channel.Items {
			title := cleanString(it.Title)
			link := strings.TrimSpace(it.Link)
			if link == "" {
				link = strings.TrimSpace(it.GUID)
			}
			desc := cleanString(it.Description)

			if r.matchQuery(title, desc, query) {
				if len(desc) > 250 { desc = desc[:247] + "..." }
				results = append(results, types.Result{
					Title: fmt.Sprintf("[%s] %s", strings.Title(feedName), title),
					URL: link,
					Snippet: desc,
				})
			}
			if len(results) >= 20 { break }
		}
		return results, nil
	}

	// Fallback to goquery for non-standard or Atom feeds
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
	if err != nil { return nil, err }

	return r.parseDocument(doc, feedName, query), nil
}
