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
	LawType   string // UU, PP, PERPRES, etc.
	LawYear   string // e.g. 2023
	LawStatus string // berlaku, dicabut, diubah
	Limit     int    // default 10, max 50
}

type pasalResponse struct {
	Total   int `json:"total"`
	Results []struct {
		Snippet  string `json:"snippet"`
		Metadata struct {
			Type       string `json:"type"`
			NodeType   string `json:"node_type"`
			NodeNumber string `json:"node_number"`
		} `json:"metadata"`
		Work struct {
			Title           string `json:"title"`
			Number          string `json:"number"`
			Year            int    `json:"year"`
			Status          string `json:"status"`
			Type            string `json:"type"`
			FRBRURI         string `json:"frbr_uri"`
			ContentVerified bool   `json:"content_verified"`
		} `json:"work"`
	} `json:"results"`
	// For /laws endpoint
	Laws []struct {
		Title           string `json:"title"`
		Number          string `json:"number"`
		Year            int    `json:"year"`
		Status          string `json:"status"`
		Type            string `json:"type"`
		FRBRURI         string `json:"frbr_uri"`
		ContentVerified bool   `json:"content_verified"`
	} `json:"laws"`
}

func (p *PasalEngine) Name() string {
	name := "Pasal.id"
	if p.LawType != "" {
		name += " (" + strings.ToUpper(p.LawType) + ")"
	}
	return name
}

func (p *PasalEngine) Search(query string) ([]types.Result, error) {
	var apiURL string
	var useLawsEndpoint bool

	if query != "" {
		apiURL = "https://pasal.id/api/v1/search"
	} else {
		apiURL = "https://pasal.id/api/v1/laws"
		useLawsEndpoint = true
	}

	u, _ := url.Parse(apiURL)
	q := u.Query()

	if !useLawsEndpoint {
		q.Set("q", query)
	}
	
	if p.LawType != "" {
		q.Set("type", p.LawType)
	}
	if p.LawYear != "" {
		q.Set("year", p.LawYear)
	}
	if p.LawStatus != "" {
		q.Set("status", p.LawStatus)
	}

	limit := p.Limit
	if limit <= 0 {
		limit = 20
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
		if resp.StatusCode == 429 {
			return nil, fmt.Errorf("pasal.id rate limit exceeded (60 req/min)")
		}
		return nil, fmt.Errorf("pasal.id api returned status %d", resp.StatusCode)
	}

	var data pasalResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var results []types.Result

	if !useLawsEndpoint {
		for _, r := range data.Results {
			results = append(results, p.formatResult(
				r.Work.Type, r.Work.Number, r.Work.Year, r.Work.Title,
				r.Work.Status, r.Work.FRBRURI, r.Snippet,
				r.Metadata.NodeType, r.Metadata.NodeNumber, r.Work.ContentVerified,
			))
		}
	} else {
		for _, l := range data.Laws {
			results = append(results, p.formatResult(
				l.Type, l.Number, l.Year, l.Title,
				l.Status, l.FRBRURI, "",
				"", "", l.ContentVerified,
			))
		}
	}

	return results, nil
}

func (p *PasalEngine) formatResult(lawType, number string, year int, title, status, uri, snippet, nodeType, nodeNum string, verified bool) types.Result {
	// Professional Title
	cleanTitle := fmt.Sprintf("[%s] No. %s Th %d - %s", lawType, number, year, title)
	if nodeNum != "" {
		cleanTitle = fmt.Sprintf("%s %s | %s", strings.Title(nodeType), nodeNum, cleanTitle)
	}

	// Prepare snippet with status and verification info
	vTag := ""
	if !verified {
		vTag = "[UNVERIFIED] "
	}
	displaySnippet := fmt.Sprintf("[STATUS:%s] %s%s", strings.ToUpper(status), vTag, snippet)

	link := "https://pasal.id" + uri
	if nodeType == "pasal" && nodeNum != "" {
		link = fmt.Sprintf("%s#pasal-%s", link, nodeNum)
	}

	return types.Result{
		Title:   cleanTitle,
		URL:     link,
		Snippet: displaySnippet,
	}
}
