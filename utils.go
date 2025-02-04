package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gobwas/glob"
	"github.com/tiktoken-go/tokenizer"
)

func countTokens(text string) int {
	enc, err := tokenizer.Get(tokenizer.O200kBase)
	if err != nil {
		panic(err)
	}
	ids, _, _ := enc.Encode(text)
	return len(ids)
}

func normalizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.Path == "/" {
		u.Path = ""
	}
	return u.Scheme + "://" + u.Host + u.Path
}

func formatNumber(num int) string {
	if num > 1000000 {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	} else if num > 1000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	}
	return fmt.Sprintf("%d", num)
}

func serializePages(pages map[string]Page) string {

	var results []string
	for _, p := range pages {
		pageStr := fmt.Sprintf(`<page>
  <title>%s</title>
  <url>%s</url>
  <content>%s</content>
</page>`, p.Title, p.URL, p.Content)
		results = append(results, pageStr)
	}
	return strings.Join(results, "\n\n")
}

func enqueue(urlStr string, skipMatch bool, opts Options, logger *Logger) {
	norm := normalizeURL(urlStr)

	visitedMu.Lock()
	if visited[norm] {
		visitedMu.Unlock()
		return
	}

	if !skipMatch && len(opts.Matches) > 0 {
		u, err := url.Parse(urlStr)
		if err != nil {
			visitedMu.Unlock()
			return
		}
		matched := false
		for _, pattern := range opts.Matches {
			g, err := glob.Compile(pattern)
			if err != nil {
				continue
			}
			if g.Match(u.Path) {
				matched = true
				break
			}
		}
		if !matched {
			visitedMu.Unlock()
			return
		}
	}
	visited[norm] = true
	visitedMu.Unlock()

	pagesMu.Lock()
	if opts.Limit > 0 && len(pages) >= opts.Limit {
		pagesMu.Unlock()
		return
	}
	pagesMu.Unlock()

	logger.Info("Fetching ", urlStr)
	wg.Add(1)
	tasks <- urlStr
}
