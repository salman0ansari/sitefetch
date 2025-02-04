package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
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
	rootCmd := &cobra.Command{
		Use:   "sitefetch [url]",
		Short: "Fetch a site and extract its readable content as Markdown",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			outfile, _ := cmd.Flags().GetString("outfile")
			concurrency, _ := cmd.Flags().GetInt("concurrency")
			matchFlag, _ := cmd.Flags().GetString("match")
			contentSelector, _ := cmd.Flags().GetString("content-selector")
			limit, _ := cmd.Flags().GetInt("limit")
			silent, _ := cmd.Flags().GetBool("silent")

			siteURL := args[0]
			var matches []string
			if matchFlag != "" {
				parts := strings.Split(matchFlag, ",")
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
				ContentSelector: contentSelector,
				Limit:           limit,
				Silent:          silent,
			}
			logger := &Logger{silent: silent}

			tasks = make(chan string, 100)
			for i := 0; i < concurrency; i++ {
				go worker(logger, opts)
			}

			logger.Info("Started fetching ", siteURL, " with a concurrency of ", concurrency)

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
				// fmt.Println(output)
			}
		},
	}

	// Define command-line flags using Cobra.
	rootCmd.Flags().String("outfile", "", "Write the fetched site to a text file")
	rootCmd.Flags().Int("concurrency", 3, "Number of concurrent requests")
	rootCmd.Flags().String("match", "", "Only fetch matched pages (comma separated)")
	rootCmd.Flags().String("content-selector", "", "The CSS selector to find content")
	rootCmd.Flags().Int("limit", 0, "Limit the result to this number of pages")
	rootCmd.Flags().Bool("silent", false, "Do not print any logs")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}
