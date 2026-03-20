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

type ExaEngine struct {
	APIKey             string
	Type               string   // "neural" or "keyword"
	Category           string   // "company", "research paper", "news", etc.
	IncludeDomains     []string
	ExcludeDomains     []string
	StartPublishedDate string   // ISO 8601
	EndPublishedDate   string   // ISO 8601
	NumResults         int
	UseAutoprompt      bool
}

type exaRequest struct {
	Query              string       `json:"query"`
	Type               string       `json:"type,omitempty"`
	Category           string       `json:"category,omitempty"`
	IncludeDomains     []string     `json:"includeDomains,omitempty"`
	ExcludeDomains     []string     `json:"excludeDomains,omitempty"`
	StartPublishedDate string       `json:"startPublishedDate,omitempty"`
	EndPublishedDate   string       `json:"endPublishedDate,omitempty"`
	NumResults         int          `json:"numResults,omitempty"`
	UseAutoprompt      bool         `json:"useAutoprompt,omitempty"`
	Contents           *exaContents `json:"contents,omitempty"`
}

type exaContents struct {
	Highlights bool `json:"highlights,omitempty"`
	Summary    bool `json:"summary,omitempty"`
}

type exaResponse struct {
	Results []exaResult `json:"results"`
}

type exaResult struct {
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	ID         string   `json:"id"`
	Score      float64  `json:"score"`
	PublishedDate string `json:"publishedDate"`
	Highlights []string `json:"highlights"`
}

func (e *ExaEngine) Name() string {
	name := "Exa Neural Search"
	if e.Type == "keyword" {
		name = "Exa Keyword Search"
	}
	return name
}

func (e *ExaEngine) Search(query string) ([]types.Result, error) {
	apiKey := e.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("EXA_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("EXA_API_KEY not found. Please set it in your .env file")
	}

	numResults := e.NumResults
	if numResults <= 0 {
		numResults = 10
	}

	reqBody := exaRequest{
		Query:              query,
		Type:               e.Type,
		Category:           e.Category,
		IncludeDomains:     e.IncludeDomains,
		ExcludeDomains:     e.ExcludeDomains,
		StartPublishedDate: e.StartPublishedDate,
		EndPublishedDate:   e.EndPublishedDate,
		NumResults:         numResults,
		UseAutoprompt:      e.UseAutoprompt,
		Contents: &exaContents{
			Highlights: true,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 25 * time.Second}
	req, err := http.NewRequest("POST", "https://api.exa.ai/search", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("Exa API error: status %d, message: %v", resp.StatusCode, errResp["message"])
	}

	var exaResp exaResponse
	if err := json.NewDecoder(resp.Body).Decode(&exaResp); err != nil {
		return nil, err
	}

	var results []types.Result
	for _, r := range exaResp.Results {
		snippet := ""
		if len(r.Highlights) > 0 {
			snippet = r.Highlights[0]
		}
		
		// Encode score and date into snippet if needed, or handle in UI
		// We'll prefix the snippet with score and date for the standard printer, 
		// but our specialized printer will handle it better.
		meta := ""
		if r.PublishedDate != "" {
			meta = fmt.Sprintf("[%s] ", r.PublishedDate[:10])
		}
		
		results = append(results, types.Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: fmt.Sprintf("%sScore: %.4f | %s", meta, r.Score, snippet),
		})
	}

	return results, nil
}
