package reader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"searx-cli/internal/util"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Article struct {
	Title   string
	Content string
}

func SmartRead(urlStr string) (*Article, error) {
	// Specialized handling for Pasal.id (uses their official API for better structure)
	// Supports both /akn/id/act/ and /peraturan/ paths
	if strings.Contains(urlStr, "pasal.id/akn/id/act/") || strings.Contains(urlStr, "pasal.id/peraturan/") {
		finalURL := urlStr
		if strings.Contains(urlStr, "/peraturan/") {
			// Redirect web URL to API logic if possible, or just let API fetcher handle it
			// The fetchPasalContent currently expects /akn/ but we can make it smarter
		}
		article, err := fetchPasalContent(finalURL)
		// If API succeeds but has low content, fallback to scraping
		if err == nil && len(article.Content) > 500 {
			return article, nil
		}
		fmt.Printf("Pasal.id API has limited content (%d chars). Falling back to web scraping via Lightpanda...\n", len(article.Content))
		return ReadURLWithLightpanda(urlStr)
	}

	if strings.Contains(urlStr, "kompas.id") {
		fmt.Printf("Attempting authenticated Kompas.id fetch: %s...\n", urlStr)
		article, err := ReadURL(urlStr)
		if err == nil && len(article.Content) > 500 {
			return article, nil
		}
		fmt.Printf("Standard Kompas.id fetch limited. Trying with Lightpanda browser...\n")
		return ReadURLWithLightpanda(urlStr)
	}

	fmt.Printf("Attempting standard fetch: %s...\n", urlStr)
	article, err := ReadURL(urlStr)

	if err != nil && (strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "429")) {
		fmt.Printf("Standard fetch blocked (%v). Retrying with Lightpanda browser...\n", err)
		return ReadURLWithLightpanda(urlStr)
	}

	if err == nil && len(article.Content) < 300 {
		fmt.Printf("Low content detected (%d chars). Retrying with Lightpanda...\n", len(article.Content))
		return ReadURLWithLightpanda(urlStr)
	}

	return article, err
}

func ReadURL(urlStr string) (*Article, error) {
	isArchive := strings.Contains(urlStr, "archive")

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", urlStr, nil)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.google.com/")

	if strings.Contains(urlStr, "kompas.id") {
		kompasCookies := os.Getenv("KOMPAS_COOKIES")
		if kompasCookies != "" {
			req.Header.Set("Cookie", kompasCookies)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return extractContent(doc, isArchive)
}

func extractContent(doc *goquery.Document, isArchive bool) (*Article, error) {
	// Kompas.id paywall bypass: remove blockers and hidden classes
	doc.Find(".paywall, #paywall-wrapper, .subscription-wrapper, .subscription-banner").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})
	// Remove blur and other styles that hide content
	doc.Find("[style*='filter: blur']").Each(func(i int, s *goquery.Selection) {
		s.SetAttr("style", "")
	})

	title := strings.TrimSpace(doc.Find("title").Text())
	var lines []string

	// Target primary content containers
	selectors := "article, main, .content, .post-content, .article-body, .media-ui-Story_body"
	contentArea := doc.Find(selectors).First()
	if contentArea.Length() == 0 {
		contentArea = doc.Find("body")
	}

	contentArea.Find("p, h1, h2, h3, h4, li, blockquote").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == "" {
			return
		}

		switch goquery.NodeName(s) {
		case "h1", "h2":
			if text == "Most Popular" || text == "The Verge Daily" || text == "Comments" || text == "Related" {
				return
			}
			lines = append(lines, "## "+text+"\n")
		case "h3":
			lines = append(lines, "### "+text+"\n")
		case "h4":
			lines = append(lines, "#### "+text+"\n")
		case "blockquote":
			lines = append(lines, "> "+text+"\n")
		case "li":
			lines = append(lines, "- "+text)
		default: // paragraph
			if len(text) > 30 {
				lines = append(lines, text+"\n")
			}
		}
	})

	content := strings.Join(lines, "\n")
	content = strings.TrimSpace(content)

	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	return &Article{
		Title:   title,
		Content: content,
	}, nil
}

func ReadURLWithLightpanda(urlStr string) (*Article, error) {
	lightpandaPath, err := util.LightpandaBinaryPath()
	if err != nil {
		return nil, err
	}

	// For Kompas.id, try to use stored cookies if available
	cookies := os.Getenv("KOMPAS_COOKIES")
	args := []string{"fetch", "--dump", "html", "--strip_mode", "js,css"}
	if cookies != "" && strings.Contains(urlStr, "kompas.id") {
		// Note: Lightpanda 'fetch' doesn't have a direct --cookie flag in help, 
		// but if it supports custom headers we could use it. 
		// Since it doesn't seem to, we'll rely on our smart reader to use them via http client.
	}

	cmd := exec.Command(lightpandaPath, append(args, urlStr)...)
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

	isArchive := strings.Contains(urlStr, "archive")
	return extractContent(doc, isArchive)
}

func fetchPasalContent(urlStr string) (*Article, error) {
	var apiURL string
	if strings.Contains(urlStr, "/akn/id/act/") {
		parts := strings.Split(urlStr, "pasal.id")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid pasal.id URL")
		}
		uri := strings.Split(parts[1], "#")[0]
		apiURL = "https://pasal.id/api/v1/laws" + uri
	} else if strings.Contains(urlStr, "/peraturan/") {
		// Slugs are not directly supported by /laws/{uri} API yet, 
		// so we let the web scraper handled it for better structure.
		return nil, fmt.Errorf("redirect to web scraper for slug")
	} else {
		return nil, fmt.Errorf("unknown pasal.id format")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pasal.id API error: %d", resp.StatusCode)
	}

	var data struct {
		Work struct {
			Title           string `json:"title"`
			Number          string `json:"number"`
			Year            int    `json:"year"`
			Status          string `json:"status"`
			ContentVerified bool   `json:"content_verified"`
		} `json:"work"`
		Articles []struct {
			Type   string `json:"type"` // bab, pasal
			Number string `json:"number"`
			Title  string `json:"title"`
			Body   string `json:"body"`
		} `json:"articles"`
		Relationships []struct {
			Type        string `json:"type"`
			RelatedWork struct {
				Title  string `json:"title"`
				Number string `json:"number"`
				Year   int    `json:"year"`
				Status string `json:"status"`
			} `json:"related_work"`
		} `json:"relationships"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	title := fmt.Sprintf("%s (No. %s Th %d)", data.Work.Title, data.Work.Number, data.Work.Year)
	var content strings.Builder

	if !data.Work.ContentVerified {
		content.WriteString("⚠️ PERINGATAN: Naskah digital untuk peraturan ini belum terverifikasi atau mungkin tidak tersedia secara lengkap.\n\n")
	}

	content.WriteString(fmt.Sprintf("STATUS: %s\n", strings.ToUpper(data.Work.Status)))

	if len(data.Relationships) > 0 {
		content.WriteString("\nRELASI HUKUM:\n")
		for _, rel := range data.Relationships {
			content.WriteString(fmt.Sprintf("- %s: %s No. %s Th %d [%s]\n", 
				rel.Type, rel.RelatedWork.Title, rel.RelatedWork.Number, rel.RelatedWork.Year, strings.ToUpper(rel.RelatedWork.Status)))
		}
		content.WriteString("\n" + strings.Repeat("-", 40) + "\n")
	}

	for _, art := range data.Articles {
		if art.Type == "bab" {
			content.WriteString(fmt.Sprintf("\n## BAB %s: %s\n\n", art.Number, art.Title))
		} else if art.Type == "pasal" {
			content.WriteString(fmt.Sprintf("### Pasal %s\n", art.Number))
			if art.Title != "" {
				content.WriteString(fmt.Sprintf("*%s*\n\n", art.Title))
			}
			content.WriteString(fmt.Sprintf("%s\n\n", art.Body))
		}
	}

	return &Article{
		Title:   title,
		Content: content.String(),
	}, nil
}
