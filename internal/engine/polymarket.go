package engine

import (
	"bytes"
	"fmt"
	"os/exec"
	"searx-cli/internal/ui"
	"searx-cli/internal/types"
	"searx-cli/internal/util"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type PolymarketEngine struct {
	Category string // trending, breaking, new, politics, sports, crypto, iran, finance, geopolitics, tech, culture, economy, weather, mentions, elections
	ShowX    bool
}

func (p *PolymarketEngine) Name() string {
	cat := p.Category
	if cat == "" {
		cat = "Breaking"
	}
	return "Polymarket (" + strings.Title(cat) + ")"
}

func (p *PolymarketEngine) Search(query string) ([]types.Result, error) {
	spinner := ui.NewSpinner("Fetching " + p.Name() + " (this may take a moment via Lightpanda)...")
	spinner.Start()
	defer spinner.Stop()

	var results []types.Result

	// 1. Map category to URL
	urlStr := p.getCategoryURL()

	// 2. Fetch Category data
	catResults, err := p.fetchCategory(urlStr)
	if err == nil {
		results = append(results, catResults...)
	} else {
		spinner.Stop()
		fmt.Printf("Warning: Failed to fetch Polymarket %s: %v\n", p.Category, err)
	}

	// 3. If ShowX is true and we're on "breaking" or "trending", fetch from X
	if p.ShowX && (p.Category == "" || p.Category == "breaking" || p.Category == "trending") {
		xResults, _ := p.fetchX()
		if xResults != nil {
			results = append(results, xResults...)
		}
	}

	if len(results) == 0 && err != nil {
		return nil, err
	}

	return results, nil
}

func (p *PolymarketEngine) getCategoryURL() string {
	base := "https://polymarket.com"
	switch strings.ToLower(p.Category) {
	case "trending":
		return base + "/trending"
	case "new":
		return base + "/new"
	case "politics":
		return base + "/politics"
	case "sports":
		return base + "/sports"
	case "crypto":
		return base + "/crypto"
	case "iran":
		return base + "/topic/iran"
	case "finance":
		return base + "/finance"
	case "geopolitics":
		return base + "/geopolitics"
	case "tech":
		return base + "/tech"
	case "culture":
		return base + "/pop-culture"
	case "economy":
		return base + "/economy"
	case "weather":
		return base + "/weather"
	case "mentions":
		return base + "/mentions"
	case "elections":
		return base + "/elections"
	default:
		return base + "/breaking"
	}
}

func (p *PolymarketEngine) fetchCategory(urlStr string) ([]types.Result, error) {
	doc, err := p.fetchWithLightpanda(urlStr)
	if err != nil {
		return nil, err
	}

	var results []types.Result
	// Find all market cards/links
	doc.Find("a[href*='/event/'], a[href*='/market/']").Each(func(i int, s *goquery.Selection) {
		if len(results) >= 20 {
			return
		}

		// Title is usually in a <p> inside the link
		pTag := s.Find("p.font-semibold").First()
		title := strings.TrimSpace(pTag.Text())
		if title == "" {
			title = strings.TrimSpace(s.Text())
		}
		
		// Clean up title
		if len(title) > 2 && title[0] >= '0' && title[0] <= '9' {
			title = strings.TrimLeft(title, "0123456789")
		}

		if title == "" || len(title) < 5 {
			return
		}

		link, _ := s.Attr("href")
		if !strings.HasPrefix(link, "http") {
			link = "https://polymarket.com" + link
		}

		// Advanced extraction for price and 24h change
		var stats []string
		var directions []string // "up" or "down"

		s.Find("div, span, svg").Each(func(j int, ss *goquery.Selection) {
			// Look for percentages
			txt := strings.TrimSpace(ss.Text())
			if strings.Contains(txt, "%") && len(txt) <= 6 {
				// Only add if not already in stats to avoid duplicates from responsive UI
				isNew := true
				for _, st := range stats {
					if st == txt {
						isNew = false
						break
					}
				}
				if isNew {
					stats = append(stats, txt)
				}
			}

			// Look for directional indicators in SVG or classes
			if ss.Is("svg") {
				html, _ := ss.Html()
				if strings.Contains(html, "rotate-45") || strings.Contains(html, "9.25 11 6 7.75") {
					// This is often an arrow
					// In Polymarket, -rotate-45 is often UP, rotate-135 is often DOWN
					outer, _ := ss.Attr("class")
					if strings.Contains(outer, "-rotate-45") {
						directions = append(directions, "up")
					} else if strings.Contains(outer, "rotate-135") || strings.Contains(outer, "rotate-45") {
						// Need careful check here
						directions = append(directions, "down")
					}
				}
			}
			
			// Fallback: Check text color classes if available
			cls, _ := ss.Attr("class")
			if strings.Contains(cls, "text-green") {
				directions = append(directions, "up")
			} else if strings.Contains(cls, "text-red") {
				directions = append(directions, "down")
			}
		})

		snippet := ""
		if len(stats) >= 2 {
			// We'll format the snippet such that the UI can color it
			// Using a convention: [UP] or [DOWN] prefix in snippet for UI to catch
			dirPrefix := "[MOVE]"
			if len(directions) > 0 {
				if directions[0] == "up" {
					dirPrefix = "[UP]"
				} else {
					dirPrefix = "[DOWN]"
				}
			}
			snippet = fmt.Sprintf("%s Price: %s | 24h: %s", dirPrefix, stats[0], stats[1])
		} else if len(stats) == 1 {
			snippet = "Price: " + stats[0]
		}

		// Deduplicate by title
		isDup := false
		for _, r := range results {
			if strings.Contains(r.Title, title) || strings.Contains(title, strings.TrimPrefix(r.Title, "[")) {
				isDup = true
				break
			}
		}

		if !isDup {
			tag := "[" + strings.Title(p.Category) + "] "
			if p.Category == "" {
				tag = "[Breaking] "
			}
			results = append(results, Result{
				Title:   tag + title,
				URL:     link,
				Snippet: snippet,
			})
		}
	})

	return results, nil
}

func (p *PolymarketEngine) fetchX() ([]types.Result, error) {
	// Try web first
	doc, err := p.fetchWithLightpanda("https://polymarket.com/breaking")
	if err == nil {
		var results []types.Result
		doc.Find("p:contains('@Polymarket')").Closest("div").Find("p").Each(func(i int, s *goquery.Selection) {
			txt := strings.TrimSpace(s.Text())
			if txt == "" || strings.Contains(txt, "@Polymarket") || strings.Contains(txt, "Follow on") || strings.Contains(txt, "See all") {
				return
			}
			if len(results) >= 5 {
				return
			}
			results = append(results, types.Result{
				Title:   "[Live] @Polymarket",
				URL:     "https://x.com/polymarket",
				Snippet: txt,
			})
		})
		if len(results) > 0 {
			return results, nil
		}
	}
	return nil, nil
}

func (p *PolymarketEngine) fetchWithLightpanda(urlStr string) (*goquery.Document, error) {
	lightpandaPath, err := util.LightpandaBinaryPath()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(lightpandaPath, "fetch", "--dump", "html", "--strip_mode", "js,css", urlStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("lightpanda error: %v, stderr: %s", err, stderr.String())
	}

	return goquery.NewDocumentFromReader(bytes.NewReader(stdout.Bytes()))
}
