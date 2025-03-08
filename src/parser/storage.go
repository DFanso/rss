package parser

import (
	"errors"
	"sync"
)

// Storage represents an in-memory storage for feeds
type Storage struct {
	feeds map[string]*Feed
	mutex sync.RWMutex
}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{
		feeds: make(map[string]*Feed),
	}
}

// AddFeed adds a feed to the storage
func (s *Storage) AddFeed(feed *Feed) error {
	if feed == nil {
		return errors.New("feed cannot be nil")
	}
	if feed.URL == "" {
		return errors.New("feed URL cannot be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.feeds[feed.URL] = feed
	return nil
}

// GetFeed gets a feed from the storage by URL
func (s *Storage) GetFeed(url string) (*Feed, error) {
	if url == "" {
		return nil, errors.New("URL cannot be empty")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	feed, ok := s.feeds[url]
	if !ok {
		return nil, errors.New("feed not found")
	}

	return feed, nil
}

// GetAllFeeds gets all feeds from the storage
func (s *Storage) GetAllFeeds() []*Feed {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	feeds := make([]*Feed, 0, len(s.feeds))
	for _, feed := range s.feeds {
		feeds = append(feeds, feed)
	}

	return feeds
}

// RemoveFeed removes a feed from the storage
func (s *Storage) RemoveFeed(url string) error {
	if url == "" {
		return errors.New("URL cannot be empty")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.feeds[url]; !ok {
		return errors.New("feed not found")
	}

	delete(s.feeds, url)
	return nil
}
