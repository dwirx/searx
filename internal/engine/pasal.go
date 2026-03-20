package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"searx-cli/internal/types"
	"strings"
	"time"
)

type PasalEngine struct {
	LawType string // UU, PP, PERPRES, etc.
	Limit   int    // default 10, max 50
}

type pasalResponse struct {
	Query   string `json:"query"`
	Total   int    `json:"total"`
	Results []struct {
		ID       int     `json:"id"`
		Snippet  string  `json:"snippet"`
		Score    float64 `json:"score"`
		Metadata struct {
			Type       string `json:"type"` // UU, PP, etc.
			NodeType   string `json:"node_type"`
			NodeNumber string `json:"node_number"`
		} `json:"metadata"`
		Work struct {
			Title   string `json:"title"`
			Number  string `json:"number"`
			Year    int    `json:"year"`
			Status  string `json:"status"` // berlaku, dicabut, diubah
			Type    string `json:"type"`
			FRBRURI string `json:"frbr_uri"`
		} `json:"work"`
	} `json:"results"`
}

func (p *PasalEngine) Name() string {
	if p.LawType != "" {
		return fmt.Sprintf("Pasal.id (%s)", strings.ToUpper(p.LawType))
	}
	return "Pasal.id (Indonesian Laws)"
}

func (p *PasalEngine) Search(query string) ([]types.Result, error) {
	baseURL := "https://pasal.id/api/v1/search"
	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set("q", query)
	if p.LawType != "" {
		q.Set("type", p.LawType)
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 15
	}
	q.Set("limit", fmt.Sprintf("%d", limit))
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", "searx-cli/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pasal.id api returned status %d", resp.StatusCode)
	}

	var data pasalResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []types.Result
	for _, r := range data.Results {
		snippet := strings.TrimSpace(r.Snippet)
		snippet = strings.ReplaceAll(snippet, "\n", " ")
		
		// Status Tag
		status := strings.ToUpper(r.Work.Status)
		
		// Professional Legal Title
		title := fmt.Sprintf("[%s] No. %s Th %d - %s", r.Work.Type, r.Work.Number, r.Work.Year, r.Work.Title)
		
		if r.Metadata.NodeNumber != "" {
			title = fmt.Sprintf("Pasal %s | %s", r.Metadata.NodeNumber, title)
		}

		// Embed status in snippet for UI to pick up
		snippet = fmt.Sprintf("[STATUS:%s] %s", status, snippet)

		link := "https://pasal.id" + r.Work.FRBRURI
		if r.Metadata.NodeType == "pasal" && r.Metadata.NodeNumber != "" {
			link = fmt.Sprintf("%s#pasal-%s", link, r.Metadata.NodeNumber)
		}

		results = append(results, types.Result{
			Title:   title,
			URL:     link,
			Snippet: snippet,
		})
	}

	return results, nil
}
