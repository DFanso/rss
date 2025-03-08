package parser

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Storage represents a storage for feeds with JSON file persistence
type Storage struct {
	feeds      map[string]*Feed
	mutex      sync.RWMutex
	filePath   string
	autoSave   bool
	lastSave   time.Time
	saveNeeded bool
}

// StorageConfig holds configuration for the storage
type StorageConfig struct {
	FilePath string
	AutoSave bool
}

// DefaultStorageConfig returns a default configuration
func DefaultStorageConfig() StorageConfig {
	return StorageConfig{
		FilePath: "feeds.json",
		AutoSave: true,
	}
}

// NewStorage creates a new storage instance
func NewStorage(config StorageConfig) *Storage {
	s := &Storage{
		feeds:    make(map[string]*Feed),
		filePath: config.FilePath,
		autoSave: config.AutoSave,
	}

	// Create directory for the file if it doesn't exist
	dir := filepath.Dir(config.FilePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("Failed to create directory for feeds file: %v", err)
		}
	}

	// Load feeds from file if it exists
	if err := s.LoadFromFile(); err != nil {
		log.Printf("Warning: Failed to load feeds from file: %v", err)
	}
	return s
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

	// Check if feed with this URL already exists
	if existingFeed, ok := s.feeds[feed.URL]; ok {
		// Update existing feed instead of adding a new one
		existingFeed.Title = feed.Title
		existingFeed.Description = feed.Description
		existingFeed.UpdatedAt = time.Now()
		existingFeed.Items = feed.Items

		s.saveNeeded = true

		if s.autoSave {
			go func() {
				if err := s.SaveToFile(); err != nil {
					log.Printf("Error saving feeds: %v", err)
				}
			}()
		}
		return nil
	}

	// Add new feed
	s.feeds[feed.URL] = feed
	s.saveNeeded = true

	if s.autoSave {
		// Save in a goroutine to avoid blocking
		go func() {
			if err := s.SaveToFile(); err != nil {
				log.Printf("Error saving feeds: %v", err)
			}
		}()
	}
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
	s.saveNeeded = true

	if s.autoSave {
		// Save in a goroutine to avoid blocking
		go func() {
			if err := s.SaveToFile(); err != nil {
				log.Printf("Error saving feeds: %v", err)
			}
		}()
	}
	return nil
}

// SaveToFile saves all feeds to a JSON file
func (s *Storage) SaveToFile() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Create directory if it doesn't exist
	dir := filepath.Dir(s.filePath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Convert feeds map to a slice for easier serialization
	feedsSlice := make([]*Feed, 0, len(s.feeds))
	for _, feed := range s.feeds {
		feedsSlice = append(feedsSlice, feed)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(feedsSlice, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first, then rename
	tempFile := s.filePath + ".tmp"
	if err := ioutil.WriteFile(tempFile, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tempFile, s.filePath); err != nil {
		// If rename fails, try direct write
		if err2 := ioutil.WriteFile(s.filePath, data, 0644); err2 != nil {
			return err2
		}
	}

	s.lastSave = time.Now()
	s.saveNeeded = false
	log.Printf("Saved %d feeds to %s", len(feedsSlice), s.filePath)
	return nil
}

// LoadFromFile loads feeds from a JSON file
func (s *Storage) LoadFromFile() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if file exists
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		log.Printf("Feeds file %s doesn't exist, starting with empty storage", s.filePath)
		return nil
	}

	// Read file
	data, err := ioutil.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	// If file is empty, return without error
	if len(data) == 0 {
		log.Printf("Feeds file %s is empty, starting with empty storage", s.filePath)
		return nil
	}

	// Unmarshal JSON
	var feedsSlice []*Feed
	if err := json.Unmarshal(data, &feedsSlice); err != nil {
		return err
	}

	// Add feeds to storage
	s.feeds = make(map[string]*Feed)
	for _, feed := range feedsSlice {
		if feed != nil && feed.URL != "" {
			s.feeds[feed.URL] = feed
		}
	}

	log.Printf("Loaded %d feeds from %s", len(feedsSlice), s.filePath)
	return nil
}

// HasChanges returns true if there are unsaved changes
func (s *Storage) HasChanges() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.saveNeeded
}

// SaveIfNeeded saves feeds to file if there are unsaved changes
func (s *Storage) SaveIfNeeded() error {
	if s.HasChanges() {
		return s.SaveToFile()
	}
	return nil
}
