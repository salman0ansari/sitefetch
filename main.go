package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gobwas/glob"
)

var (
	tasks     chan string
	wg        sync.WaitGroup
	visited   = make(map[string]bool)
	visitedMu sync.Mutex
	pages     = make(map[string]Page)
	pagesMu   sync.Mutex
)

func worker(logger *Logger, opts Options) {
	for urlStr := range tasks {
		fetchPage(urlStr, logger, opts)
		wg.Done()
	}
}

func main() {
	logger := &Logger{silent: false}
	outfile := "pages.txt"
	concurrency := 3
	match := ""
	selector := ""
	limit := 10
	silent := false

	siteURL := "https://hisalman.in"
	var matches []string
	if match != "" {
		parts := strings.Split(match, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				matches = append(matches, p)
			}
		}
	}

	opts := Options{
		Concurrency:     concurrency,
		Matches:         matches,
		ContentSelector: selector,
		Limit:           limit,
		Silent:          silent,
	}

	tasks = make(chan string, 100)
	for i := 0; i < concurrency; i++ {
		go worker(logger, opts)
	}

	logger.Info("Started fetching", siteURL, "with a concurrency of", concurrency)

	enqueue(siteURL, true, opts, logger)

	wg.Wait()
	close(tasks)

	totalTokens := 0
	pagesMu.Lock()
	for _, page := range pages {
		totalTokens += countTokens(page.Content)
	}
	count := len(pages)
	pagesMu.Unlock()
	logger.Info("Total token count for ", count, " pages: ", formatNumber(totalTokens))

	output := serializePages(pages)
	if outfile != "" {
		if err := os.MkdirAll(filepath.Dir(outfile), os.ModePerm); err != nil {
			logger.Warn("Failed to create directory:", err)
			return
		}
		if err := os.WriteFile(outfile, []byte(output), 0644); err != nil {
			logger.Warn("Failed to write file:", err)
			return
		}
	} else {
		fmt.Println(output)
	}

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
