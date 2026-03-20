package reader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	if strings.Contains(urlStr, "pasal.id/akn/id/act/") {
		return fetchPasalContent(urlStr)
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

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d (Access Denied/Blocked)", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return extractContent(doc, isArchive)
}

func extractContent(doc *goquery.Document, isArchive bool) (*Article, error) {
	// 1. Pre-cleanup: Remove elements that are almost always junk
	junkSelectors := []string{
		"script", "style", "iframe", "noscript", "footer", "nav", "header",
		".header", ".footer", ".nav", ".sidebar", ".ad", ".ads", "#wm-ipp-base",
		".mw-editsection", ".newsletter-signup", ".social-share", ".related-posts",
		".duet--article--newsletter-form", ".duet--content-cards", ".duet--article--author-bio",
		".duet--article--related-list", ".duet--article--social-share", "aside",
		".duet--article--article-footer", ".duet--article--related-item",
		".duet--article--top-rule", ".duet--article--article-header",
		".duet--article--article-meta", ".duet--article--ad-unit",
		".related-content", ".recommendations", ".trending-now", ".comments-area",
		".tags-list", ".post-tags", ".article-tags",
	}
	for _, sel := range junkSelectors {
		doc.Find(sel).Remove()
	}

	// 2. Get Title
	title := strings.TrimSpace(doc.Find("h1").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("title").First().Text())
	}

	var lines []string
	seenParagraphs := make(map[string]bool)

	// 3. Find Article Body
	var body *goquery.Selection
	selectors := []string{
		"article", "main", ".duet--article--article-body", ".article-body",
		".story-body", ".post-content", "section[name='articleBody']",
		".entry-content", ".content-body", "#article-content", ".article__body",
	}

	for _, sel := range selectors {
		if found := doc.Find(sel); found.Length() > 0 {
			body = found
			break
		}
	}

	if body == nil {
		body = doc.Find("body")
	}

	// 4. Extract Structured Content
	body.Find("p, h2, h3, h4, li, blockquote").Each(func(i int, s *goquery.Selection) {
		// Aggressive container check
		if s.Closest(".newsletter, .ad, .social, .related, aside, footer, nav, .duet--article--newsletter-form, .duet--content-cards, #most-popular").Length() > 0 {
			return
		}

		text := strings.TrimSpace(s.Text())
		if text == "" {
			return
		}

		// FILTER JUNK PHRASES
		junkPhrases := []string{
			"daily email digest", "homepage feed", "FollowFollow",
			"Posts from this topic", "Posts from this author",
			"A free daily digest", "Terms of Service", "Privacy Policy",
			"See All by", "Senior Reviewer", "Senior Reporter",
			"link copied", "copy link", "share on facebook", "share on twitter",
			"Read more:", "Sign up for", "Log in", "Subscribe",
		}
		for _, phrase := range junkPhrases {
			if strings.Contains(strings.ToLower(text), strings.ToLower(phrase)) {
				return
			}
		}

		// UI Keywork Filtering (Very short lines)
		if len(text) < 20 {
			uiKeywords := []string{"Link", "Share", "Gift", "Report", "Gaming", "Policy", "Antitrust", "Comment", "Print", "Email"}
			for _, kw := range uiKeywords {
				if text == kw {
					return
				}
			}
		}

		// Deduplication (don't add the same paragraph twice)
		if seenParagraphs[text] {
			return
		}
		seenParagraphs[text] = true

		tagName := goquery.NodeName(s)
		switch tagName {
		case "h2":
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

	// Final whitespace cleanup
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

	cmd := exec.Command(lightpandaPath, "fetch", "--dump", "html", "--strip_mode", "js,css", urlStr)
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

// fetchPasalContent fetches full law details from pasal.id API
func fetchPasalContent(urlStr string) (*Article, error) {
	// Extract FRBR URI from URL
	// URL example: https://pasal.id/akn/id/act/uu/2020/11#pasal-1
	parts := strings.Split(urlStr, "pasal.id")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid pasal.id URL")
	}
	uri := strings.Split(parts[1], "#")[0]
	apiURL := "https://pasal.id/api/v1/laws" + uri

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
			Title  string `json:"title"`
			Number string `json:"number"`
			Year   int    `json:"year"`
			Status string `json:"status"`
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
