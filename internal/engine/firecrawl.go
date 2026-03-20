package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"searx-cli/internal/types"
	"time"
)

type FirecrawlEngine struct {
	APIKey        string
	Limit         int
	Lang          string
	Country       string
	Location      string
	Tbs           string
	ScrapeFormats []string // e.g. ["markdown", "html"]
}

type firecrawlRequest struct {
	Query         string                 `json:"query"`
	Limit         int                    `json:"limit,omitempty"`
	Lang          string                 `json:"lang,omitempty"`
	Country       string                 `json:"country,omitempty"`
	Location      string                 `json:"location,omitempty"`
	Tbs           string                 `json:"tbs,omitempty"`
	ScrapeOptions *firecrawlScrapeOpts   `json:"scrapeOptions,omitempty"`
}

type firecrawlScrapeOpts struct {
	Formats []string `json:"formats,omitempty"`
}

type firecrawlResponse struct {
	Success bool              `json:"success"`
	Data    firecrawlData     `json:"data"`
	Error   string            `json:"error,omitempty"`
}

type firecrawlData struct {
	Web    []firecrawlResult `json:"web"`
	News   []firecrawlResult `json:"news"`
	Images []firecrawlResult `json:"images"`
}

type firecrawlResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Snippet     string `json:"snippet"` // for news
	Date        string `json:"date"`    // for news
	Markdown    string `json:"markdown,omitempty"`
	HTML        string `json:"html,omitempty"`
}

func (f *FirecrawlEngine) Name() string {
	return "Firecrawl Search"
}

func (f *FirecrawlEngine) Search(query string) ([]types.Result, error) {
	apiKey := f.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("FIRECRAWL_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("FIRECRAWL_API_KEY not found. Please set it in your .env file")
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 5
	}

	reqBody := firecrawlRequest{
		Query:    query,
		Limit:    limit,
		Lang:     f.Lang,
		Country:  f.Country,
		Location: f.Location,
		Tbs:      f.Tbs,
	}

	if len(f.ScrapeFormats) > 0 {
		reqBody.ScrapeOptions = &firecrawlScrapeOpts{
			Formats: f.ScrapeFormats,
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", "https://api.firecrawl.dev/v2/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errMap map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errMap)
		return nil, fmt.Errorf("Firecrawl API error: status %d, message: %v", resp.StatusCode, errMap["error"])
	}

	var fireResp firecrawlResponse
	if err := json.NewDecoder(resp.Body).Decode(&fireResp); err != nil {
		return nil, err
	}

	if !fireResp.Success {
		return nil, fmt.Errorf("Firecrawl API unsuccessful: %s", fireResp.Error)
	}

	var results []types.Result
	
	// Process Web results
	for _, r := range fireResp.Data.Web {
		results = append(results, types.Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Description,
		})
	}
	
	// Process News results
	for _, r := range fireResp.Data.News {
		snippet := r.Snippet
		if r.Date != "" {
			snippet = fmt.Sprintf("[%s] %s", r.Date, snippet)
		}
		results = append(results, types.Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: snippet,
		})
	}

	// Limit total results
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
