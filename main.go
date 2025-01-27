package main

import (
	"net/url"
	"os"
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
	match := ""

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
		Concurrency: 3,
		Silent:      false,
		Matches:     matches,
	}

	tasks = make(chan string, 100)
	for i := 0; i < opts.Concurrency; i++ {
		worker(logger, opts)
	}

	logger.Info("Started fetching", siteURL, " with a concurrency of ", opts.Concurrency)
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

	outfile := "pages.txt"
	result := serializePages(pages)

	if err := os.WriteFile(outfile, []byte(result), 0644); err != nil {
		logger.Warn("Failed to write file:", err)
		return
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
