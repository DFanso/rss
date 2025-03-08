package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/user/rss/src/parser"
	"github.com/user/rss/src/server"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "HTTP server port")
	defaultFeeds := flag.String("feeds", "", "Comma-separated list of default RSS feed URLs")
	dataDir := flag.String("data", "data", "Directory to store data files")
	flag.Parse()

	// Ensure data directory exists
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Create a new storage with JSON persistence
	feedsFile := filepath.Join(*dataDir, "feeds.json")
	storageConfig := parser.StorageConfig{
		FilePath: feedsFile,
		AutoSave: true,
	}
	storage := parser.NewStorage(storageConfig)

	// Add default feeds if specified and if storage is empty
	if *defaultFeeds != "" && len(storage.GetAllFeeds()) == 0 {
		log.Println("Adding default feeds")
		urls := parser.SplitURLs(*defaultFeeds)
		for _, url := range urls {
			log.Printf("Fetching default feed: %s", url)
			feed, err := parser.FetchFeed(url)
			if err != nil {
				log.Printf("Error fetching feed %s: %v", url, err)
				continue
			}
			err = storage.AddFeed(feed)
			if err != nil {
				log.Printf("Error adding feed %s: %v", url, err)
			}
		}
	}

	// Create and start the HTTP server
	srv := server.NewServer(storage)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting RSS server on http://localhost%s", addr)
	log.Printf("Feeds will be saved to %s", feedsFile)

	// Start the server in a goroutine
	go func() {
		if err := srv.Start(addr); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server shutting down...")

	// Save any pending changes
	if err := storage.SaveIfNeeded(); err != nil {
		log.Printf("Error saving feeds: %v", err)
	}
}
