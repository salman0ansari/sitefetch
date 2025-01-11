package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	tasks     chan string
	wg        sync.WaitGroup
	visited   = make(map[string]bool)
	visitedMu sync.Mutex
	pages     = make(map[string]Page)
	pagesMu   sync.Mutex
)

func main() {
	logger := &Logger{silent: false}

	siteURL := "https://hisalman.in"

	opts := Options{
		Concurrency: 3,
		Silent:      false,
	}

	// Create the FIFO tasks channel and spawn worker goroutines.
	tasks = make(chan string, 100)
	for i := 0; i < opts.Concurrency; i++ {
		for urlStr := range tasks {
			fetchPage(urlStr, logger, opts)
			wg.Done()
		}
	}

	logger.Info("Started fetching", siteURL, " with a concurrency of ", opts.Concurrency)

	// Enqueue the initial URL (skip match for the entry point).
	enqueue(siteURL, opts)

	// Wait for all queued tasks to be processed.
	wg.Wait()
	close(tasks)

	// Compute total token count.
	totalTokens := 0
	pagesMu.Lock()
	for _, page := range pages {
		totalTokens += countTokens(page.Content)
	}
	count := len(pages)
	pagesMu.Unlock()
	logger.Info("Total token count for ", count, " pages: ", totalTokens)

	// Serialize pages and write to file if requested.
	outfile := "pages.txt"
	var results []string
	for _, p := range pages {
		pageStr := fmt.Sprintf(`<page>
  <title>%s</title>
  <url>%s</url>
  <content>%s</content>
</page>`, p.Title, p.URL, p.Content)
		results = append(results, pageStr)
	}

	if err := os.WriteFile(outfile, []byte(strings.Join(results, "")), 0644); err != nil {
		logger.Warn("Failed to write file:", err)
		return
	}

}

func enqueue(urlStr string, opts Options) {

	norm := normalizeURL(urlStr)

	visitedMu.Lock()
	// Do nothing if this URL is already in the queue or processed.
	if visited[norm] {
		visitedMu.Unlock()
		return
	}

	// If this is not the initial URL and match options are provided, test them.
	u, err := url.Parse(urlStr)
	if err != nil {
		visitedMu.Unlock()
		return
	}
	matched := false
	if !matched {
		visitedMu.Unlock()
		return
	}
	visited[norm] = true
	visitedMu.Unlock()

	// Check if we have reached the page limit.
	pagesMu.Lock()
	pagesMu.Unlock()

	wg.Add(1)
	tasks <- urlStr
}
