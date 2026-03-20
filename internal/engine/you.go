package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"searx-cli/internal/types"
	"strings"
	"time"
)

type YouEngine struct {
	APIKey   string
	Count    int
	Country  string
	Language string
	Mode     string // "search", "research", "contents"
}

// Search API Response
type youFullResponse struct {
	Results struct {
		Web  []youWebResult  `json:"web"`
		News []youNewsResult `json:"news"`
	} `json:"results"`
}

type youWebResult struct {
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Snippets    []string `json:"snippets"`
}

type youNewsResult struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Research API Response
type youResearchResponse struct {
	Output struct {
		Content string `json:"content"`
		Sources []struct {
			URL   string `json:"url"`
			Title string `json:"title"`
		} `json:"sources"`
	} `json:"output"`
}

// Contents API Response
type youContentsResponse []struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (y *YouEngine) Name() string {
	switch strings.ToLower(y.Mode) {
	case "research":
		return "You.com Research"
	case "contents":
		return "You.com Contents"
	default:
		return "You.com Search"
	}
}

func (y *YouEngine) Search(query string) ([]types.Result, error) {
	apiKey := y.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("YOU_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("YOU_API_KEY not found. Please set it in your .env file")
	}

	mode := strings.ToLower(y.Mode)
	if mode == "" {
		mode = "search"
	}

	switch mode {
	case "research":
		return y.searchResearch(query, apiKey)
	case "contents":
		return y.searchContents(query, apiKey)
	default:
		return y.searchWeb(query, apiKey)
	}
}

func (y *YouEngine) searchWeb(query string, apiKey string) ([]types.Result, error) {
	u, _ := url.Parse("https://ydc-index.io/v1/search")
	q := u.Query()
	q.Set("query", query)
	if y.Count > 0 {
		q.Set("count", fmt.Sprintf("%d", y.Count))
	}
	if y.Country != "" {
		q.Set("country", y.Country)
	}
	if y.Language != "" {
		q.Set("language", y.Language)
	}
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 25 * time.Second}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("You.com API error: status %d, message: %v", resp.StatusCode, errResp["message"])
	}

	var youResp youFullResponse
	if err := json.NewDecoder(resp.Body).Decode(&youResp); err != nil {
		return nil, err
	}

	var results []types.Result
	for _, r := range youResp.Results.Web {
		snippet := r.Description
		if len(r.Snippets) > 0 {
			snippet = strings.Join(r.Snippets, " ")
		}
		results = append(results, types.Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: snippet,
		})
	}

	if len(results) < 10 {
		for _, r := range youResp.Results.News {
			results = append(results, types.Result{
				Title:   "[NEWS] " + r.Title,
				URL:     r.URL,
				Snippet: r.Description,
			})
			if len(results) >= 20 {
				break
			}
		}
	}

	return results, nil
}

func (y *YouEngine) searchResearch(query string, apiKey string) ([]types.Result, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"input": query,
	})

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", "https://api.you.com/v1/research", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("You.com Research API error: status %d, message: %v", resp.StatusCode, errResp["message"])
	}

	var res youResearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	results := []types.Result{
		{
			Title:   "Research Answer",
			URL:     "https://you.com",
			Snippet: res.Output.Content,
		},
	}

	for _, s := range res.Output.Sources {
		results = append(results, types.Result{
			Title:   "[Source] " + s.Title,
			URL:     s.URL,
			Snippet: "",
		})
	}

	return results, nil
}

func (y *YouEngine) searchContents(query string, apiKey string) ([]types.Result, error) {
	// Query is expected to be a URL or comma separated URLs
	urls := strings.Split(query, ",")
	for i := range urls {
		urls[i] = strings.TrimSpace(urls[i])
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"urls": urls,
	})

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://api.you.com/v1/contents", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("You.com Contents API error: status %d, message: %v", resp.StatusCode, errResp["message"])
	}

	var res youContentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	var results []types.Result
	for _, r := range res {
		results = append(results, types.Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
		})
	}

	return results, nil
}
