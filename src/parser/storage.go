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

// FeedMetadata represents the essential information about a feed without its content
type FeedMetadata struct {
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AddedAt     time.Time `json:"added_at"`
}

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
		// Update only metadata for existing feed
		existingFeed.Title = feed.Title
		existingFeed.Description = feed.Description
		// Don't update the items - we'll fetch them fresh each time

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

	// Add new feed (only metadata is important for storage)
	s.feeds[feed.URL] = &Feed{
		URL:         feed.URL,
		Title:       feed.Title,
		Description: feed.Description,
		UpdatedAt:   time.Now(),
		// Don't store items - we'll fetch them fresh when needed
		Items: nil,
	}

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

// GetFeed gets a feed from the storage by URL and refreshes its content
func (s *Storage) GetFeed(url string) (*Feed, error) {
	if url == "" {
		return nil, errors.New("URL cannot be empty")
	}

	s.mutex.RLock()
	storedFeed, ok := s.feeds[url]
	s.mutex.RUnlock()

	if !ok {
		return nil, errors.New("feed not found")
	}

	// Always fetch fresh content for the feed
	log.Printf("Fetching fresh content for feed: %s", url)
	freshFeed, err := FetchFeed(url)
	if err != nil {
		return nil, err
	}

	// Update stored feed metadata if needed
	s.mutex.Lock()
	storedFeed.Title = freshFeed.Title
	storedFeed.Description = freshFeed.Description
	storedFeed.UpdatedAt = time.Now()
	s.mutex.Unlock()

	// Return the fresh feed with content
	return freshFeed, nil
}

// GetAllFeeds gets all feeds from the storage (without their content)
func (s *Storage) GetAllFeeds() []*Feed {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	feeds := make([]*Feed, 0, len(s.feeds))
	for _, feed := range s.feeds {
		// Create a copy without items to reduce memory usage
		feedCopy := &Feed{
			URL:         feed.URL,
			Title:       feed.Title,
			Description: feed.Description,
			UpdatedAt:   feed.UpdatedAt,
			// Don't include items - they'll be fetched when needed
			Items: nil,
		}
		feeds = append(feeds, feedCopy)
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

// SaveToFile saves feed metadata to a JSON file
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

	// Convert feeds map to metadata for more efficient storage
	metadataList := make([]FeedMetadata, 0, len(s.feeds))
	for _, feed := range s.feeds {
		metadata := FeedMetadata{
			URL:         feed.URL,
			Title:       feed.Title,
			Description: feed.Description,
			AddedAt:     feed.UpdatedAt,
		}
		metadataList = append(metadataList, metadata)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(metadataList, "", "  ")
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
	log.Printf("Saved %d feed subscriptions to %s", len(metadataList), s.filePath)
	return nil
}

// LoadFromFile loads feed metadata from a JSON file
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
	var metadataList []FeedMetadata
	if err := json.Unmarshal(data, &metadataList); err != nil {
		return err
	}

	// Add feeds to storage from metadata
	s.feeds = make(map[string]*Feed)
	for _, metadata := range metadataList {
		if metadata.URL != "" {
			s.feeds[metadata.URL] = &Feed{
				URL:         metadata.URL,
				Title:       metadata.Title,
				Description: metadata.Description,
				UpdatedAt:   metadata.AddedAt,
				// Don't load items - will fetch fresh when needed
				Items: nil,
			}
		}
	}

	log.Printf("Loaded %d feed subscriptions from %s", len(metadataList), s.filePath)
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
