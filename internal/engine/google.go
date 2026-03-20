package engine

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/url"
	"os/exec"
	"searx-cli/internal/types"
	"searx-cli/internal/util"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type GoogleEngine struct{}

func (g *GoogleEngine) Name() string {
	return "Google"
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:123.0) Gecko/20100101 Firefox/123.0",
	"Mozilla/5.0 (AppleWebKit/537.36; KHTML, like Gecko) Chrome/121.0.6167.160 Safari/537.36",
}

var googleDomains = []string{
	"www.google.com",
	"www.google.co.uk",
	"www.google.co.id",
	"www.google.ca",
	"www.google.de",
}

func randomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
}

func (g *GoogleEngine) Search(query string) ([]types.Result, error) {
	rand.Seed(time.Now().UnixNano())
	domain := googleDomains[rand.Intn(len(googleDomains))]
	
	u, _ := url.Parse(fmt.Sprintf("https://%s/search", domain))
	q := u.Query()
	q.Set("q", query)
	q.Set("gbv", "1")
	// udm=14 often gives a cleaner results-only view that might bypass some modern bot checks
	if rand.Intn(2) == 0 {
		q.Set("udm", "14")
	}
	u.RawQuery = q.Encode()

	results, err := g.SearchWithLightpanda(u.String())
	if err == nil && len(results) > 0 {
		return results, nil
	}

	// Fallback strategy: try other domains and randomized delays
	for i := 0; i < 2; i++ {
		time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)
		newDomain := googleDomains[rand.Intn(len(googleDomains))]
		u.Host = newDomain
		results, err = g.SearchWithLightpanda(u.String())
		if err == nil && len(results) > 0 {
			return results, nil
		}
	}

	// Final fallback to other engines
	fmt.Println("Google stealth methods blocked. Falling back to other engines...")
	
	ddg := &DuckDuckGoEngine{}
	res, err := ddg.Search(query)
	if err == nil && len(res) > 0 {
		return res, nil
	}

	instances := []string{"https://searx.be", "https://paulgo.io", "https://searx.org"}
	for _, inst := range instances {
		searx := &SearxEngine{InstanceURL: inst}
		res, err := searx.Search(query)
		if err == nil && len(res) > 0 {
			return res, nil
		}
	}
	
	return nil, fmt.Errorf("google failed and all fallbacks exhausted")
}

func (g *GoogleEngine) SearchWithLightpanda(urlStr string) ([]types.Result, error) {
	lightpandaPath, err := util.LightpandaBinaryPath()
	if err != nil {
		return nil, err
	}

	ua := userAgents[rand.Intn(len(userAgents))]
	suffix := " " + ua

	// Note: Lightpanda might not support --header directly in 'fetch' command based on help output,
	// but we can try to influence it via UA or other flags if available.
	// Since --help didn't show --header, we'll stick to UA suffix and hope for the best.
	cmd := exec.Command(lightpandaPath, "fetch", "--dump", "html", "--user_agent_suffix", suffix, urlStr)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("lightpanda error: %v, stderr: %s", err, stderr.String())
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(stdout.Bytes()))
	if err != nil {
		return nil, err
	}

	return g.parseGoogleResults(doc)
}

func (g *GoogleEngine) parseGoogleResults(doc *goquery.Document) ([]types.Result, error) {
	var results []types.Result
	
	// Modern selectors
	doc.Find("div.g, div.tF2Cxc, div.MjjYud").Each(func(i int, s *goquery.Selection) {
		if len(results) >= 10 {
			return
		}
		title := s.Find("h3").First().Text()
		link, _ := s.Find("a").First().Attr("href")
		snippet := s.Find("div.VwiC3b, span.st").First().Text()

		if title != "" && link != "" && !strings.HasPrefix(link, "/") {
			results = append(results, types.Result{
				Title:   title,
				URL:     link,
				Snippet: snippet,
			})
		}
	})

	// GBV=1 / Legacy / Mobile selectors
	if len(results) == 0 {
		doc.Find("div.ZINbbc, div.kCrYT").Each(func(i int, s *goquery.Selection) {
			if len(results) >= 10 {
				return
			}
			
			title := s.Find("h3").Text()
			if title == "" {
				title = s.Find(".vv798d").Text()
			}
			if title == "" { return }

			linkNode := s.Find("a").First()
			href, _ := linkNode.Attr("href")
			
			cleanURL := strings.TrimPrefix(href, "/url?q=")
			if idx := strings.Index(cleanURL, "&"); idx != -1 {
				cleanURL = cleanURL[:idx]
			}
			
			snippet := s.Find(".VwiC3b, .BNeawe.s3v9rd.AP7Wnd, .st").First().Text()

			if cleanURL != "" && !strings.HasPrefix(cleanURL, "/") {
				results = append(results, types.Result{
					Title:   title,
					URL:     cleanURL,
					Snippet: snippet,
				})
			}
		})
	}

	return results, nil
}
