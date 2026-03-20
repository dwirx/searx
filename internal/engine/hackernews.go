package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"searx-cli/internal/types"
	"sync"
	"time"
)

type HackerNewsEngine struct {
	Category string // top, new, best, ask, show, job
}

type hnItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
	By    string `json:"by"`
	Score int    `json:"score"`
	Descendants int `json:"descendants"` // comment count
}

func (h *HackerNewsEngine) Name() string {
	return "Hacker News API (" + h.Category + ")"
}

func (h *HackerNewsEngine) Search(query string) ([]types.Result, error) {
	var endpoint string
	switch h.Category {
	case "new", "newest":
		endpoint = "newstories"
	case "best":
		endpoint = "beststories"
	case "ask":
		endpoint = "askstories"
	case "show":
		endpoint = "showstories"
	case "job":
		endpoint = "jobstories"
	default:
		endpoint = "topstories"
	}

	client := &http.Client{Timeout: 10 * time.Second}
	
	// 1. Get IDs
	resp, err := client.Get(fmt.Sprintf("https://hacker-news.firebaseio.com/v0/%s.json", endpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ids []int
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, err
	}

	// Limit to top 20
	if len(ids) > 20 {
		ids = ids[:20]
	}

	// 2. Fetch item details in parallel
	results := make([]types.Result, len(ids))
	var wg sync.WaitGroup
	errChan := make(chan error, len(ids))

	for i, id := range ids {
		wg.Add(1)
		go func(i int, id int) {
			defer wg.Done()
			item, err := h.fetchItem(client, id)
			if err != nil {
				errChan <- err
				return
			}
			
			url := item.URL
			if url == "" {
				url = fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID)
			}

			snippet := fmt.Sprintf("%d points by %s | %d comments", item.Score, item.By, item.Descendants)
			results[i] = types.Result{
				Title:   item.Title,
				URL:     url,
				Snippet: snippet,
			}
		}(i, id)
	}

	wg.Wait()
	close(errChan)

	// Filter out empty results in case of partial failures
	var finalResults []types.Result
	for _, r := range results {
		if r.Title != "" {
			finalResults = append(finalResults, r)
		}
	}

	return finalResults, nil
}

func (h *HackerNewsEngine) fetchItem(client *http.Client, id int) (*hnItem, error) {
	resp, err := client.Get(fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", id))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var item hnItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}
