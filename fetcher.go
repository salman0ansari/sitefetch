package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	markdown "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
)

type Options struct {
	Concurrency int
	Silent      bool
}

type Page struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Content string `json:"content"`
}

func fetchPage(urlStr string, logger *Logger, opts Options) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Warn("Invalid URL:", urlStr)
		return
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(urlStr)
	if err != nil {
		logger.Warn("Failed to fetch", urlStr, ":", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Warn("Failed to fetch", urlStr, "status:", resp.Status)
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		logger.Warn("Not an HTML page:", urlStr)
		return
	}

	if resp.Request.URL.Host != parsedURL.Host {
		logger.Warn("Redirected from", parsedURL.Host, "to", resp.Request.URL.Host)
		return
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warn("Failed to read body from", urlStr, ":", err)
		return
	}
	bodyStr := string(bodyBytes)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(bodyStr))
	if err != nil {
		logger.Warn("Failed to parse HTML for", urlStr, ":", err)
		return
	}

	doc.Find("script, style, link, img, video").Remove()

	// Extract extra URLs in the order that they appear.
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || strings.TrimSpace(href) == "" {
			return
		}
		linkURL, err := url.Parse(href)
		if err != nil {
			return
		}
		absURL := parsedURL.ResolveReference(linkURL)
		if absURL.Host != parsedURL.Host {
			return
		}
	})

	pageTitle := doc.Find("title").Text()

	var htmlContent string
	if htmlContent == "" {
		htmlContent, err = doc.Html()
		if err != nil {
			logger.Warn("Failed to get HTML content for", urlStr)
			return
		}
	}

	article, err := readability.FromReader(strings.NewReader(htmlContent), parsedURL)
	if err != nil {
		logger.Warn("Failed to parse article for", urlStr, ":", err)
		return
	}

	converter := markdown.NewConverter("", true, nil)
	mdContent, err := converter.ConvertString(article.Content)
	if err != nil {
		logger.Warn("Failed to convert HTML to Markdown for", urlStr, ":", err)
		return
	}

	finalTitle := article.Title
	if finalTitle == "" {
		finalTitle = pageTitle
	}

	norm := normalizeURL(urlStr)
	pagesMu.Lock()
	pages[norm] = Page{
		Title:   finalTitle,
		URL:     urlStr,
		Content: mdContent,
	}
	pagesMu.Unlock()
}
