package parser

import (
	"strings"
)

// SplitURLs splits a comma-separated string of URLs into a slice
func SplitURLs(urlsStr string) []string {
	if urlsStr == "" {
		return []string{}
	}

	// Split the string by commas
	urls := strings.Split(urlsStr, ",")

	// Trim whitespace from each URL
	for i, url := range urls {
		urls[i] = strings.TrimSpace(url)
	}

	// Filter out empty URLs
	filtered := make([]string, 0, len(urls))
	for _, url := range urls {
		if url != "" {
			filtered = append(filtered, url)
		}
	}

	return filtered
}
