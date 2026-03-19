package reader

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Article struct {
	Title   string
	Content string
}

func SmartRead(urlStr string) (*Article, error) {
	fmt.Printf("Attempting standard fetch: %s...\n", urlStr)
	article, err := ReadURL(urlStr)
	
	// If it's a 403 or 429, or the content is suspicious (e.g. "Access Denied"), try Lightpanda
	if err != nil && (strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "429")) {
		fmt.Printf("Standard fetch blocked (%v). Retrying with Lightpanda browser...\n", err)
		return ReadURLWithLightpanda(urlStr)
	}

	// If it succeeded but returned very little content (common for paywalls/bot protection)
	if err == nil && len(article.Content) < 200 {
		fmt.Printf("Low content detected. Retrying with Lightpanda to handle JavaScript/Protection...\n")
		return ReadURLWithLightpanda(urlStr)
	}

	return article, err
}

func ReadURL(urlStr string) (*Article, error) {
	// Support archive.is and variants
	isArchive := strings.Contains(urlStr, "archive.is") || 
				 strings.Contains(urlStr, "archive.today") || 
				 strings.Contains(urlStr, "archive.ph") ||
				 strings.Contains(urlStr, "archive.li") ||
				 strings.Contains(urlStr, "archive.vn") ||
				 strings.Contains(urlStr, "archive.fo") ||
				 strings.Contains(urlStr, "archive.md")

	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", urlStr, nil)
	
	// Better headers and Referer bypass
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://www.google.com/") // Try to look like a search result click
	
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
	title := strings.TrimSpace(doc.Find("title").First().Text())
	if title == "" {
		title = doc.Find("h1").First().Text()
	}
	
	var contentBuilder strings.Builder

	if isArchive {
		// Archive.is often wraps the content in a specific way
		doc.Find("#wm-ipp-base, #wm-ipp-print, #header, .header").Remove() 
		
		doc.Find("article, .article, .story, #article, div[role='main'], #ins_storybody, .story-content").Find("p").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if len(text) > 30 {
				contentBuilder.WriteString(text + "\n\n")
			}
		})
	} else {
		// Standard news site logic
		doc.Find("section[name='articleBody'], .StoryBodyCompanionColumn, .article-body, article, main, .post-content").Find("p").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if len(text) > 30 {
				contentBuilder.WriteString(text + "\n\n")
			}
		})
	}

	content := contentBuilder.String()
	if content == "" {
		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if len(text) > 60 && !strings.Contains(text, "©") && !strings.Contains(text, "Terms of") {
				contentBuilder.WriteString(text + "\n\n")
			}
		})
		content = contentBuilder.String()
	}

	return &Article{
		Title:   title,
		Content: content,
	}, nil
}

func ReadURLWithLightpanda(urlStr string) (*Article, error) {
	// Use lightpanda fetch with stealth flags
	cmd := exec.Command("./lightpanda", "fetch", "--dump", "html", "--strip_mode", "js,css", urlStr)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
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
