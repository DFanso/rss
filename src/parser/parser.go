package parser

import (
	"errors"
	"time"

	"github.com/mmcdole/gofeed"
)

// Feed represents an RSS feed
type Feed struct {
	URL         string
	Title       string
	Description string
	UpdatedAt   time.Time
	Items       []FeedItem
}

// FeedItem represents a single item in an RSS feed
type FeedItem struct {
	Title       string
	Description string
	Content     string
	Link        string
	PublishedAt time.Time
	GUID        string
}

// FetchFeed fetches and parses an RSS feed from the given URL
func FetchFeed(url string) (*Feed, error) {
	if url == "" {
		return nil, errors.New("URL cannot be empty")
	}

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}

	result := &Feed{
		URL:         url,
		Title:       feed.Title,
		Description: feed.Description,
		UpdatedAt:   time.Now(),
		Items:       make([]FeedItem, 0, len(feed.Items)),
	}

	for _, item := range feed.Items {
		feedItem := FeedItem{
			Title:       item.Title,
			Description: item.Description,
			Content:     item.Content,
			Link:        item.Link,
			GUID:        item.GUID,
		}

		if item.PublishedParsed != nil {
			feedItem.PublishedAt = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			feedItem.PublishedAt = *item.UpdatedParsed
		}

		result.Items = append(result.Items, feedItem)
	}

	return result, nil
}
