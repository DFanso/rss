package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/rss/src/parser"
	"github.com/user/rss/src/server"
)

func main() {
	// Parse command line flags
	port := flag.Int("port", 8080, "HTTP server port")
	defaultFeeds := flag.String("feeds", "", "Comma-separated list of default RSS feed URLs")
	flag.Parse()

	// Create a new storage
	storage := parser.NewStorage()

	// Add default feeds if specified
	if *defaultFeeds != "" {
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

	// Start the server in a goroutine
	go func() {
		if err := srv.Start(addr); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Server shutting down...")
}
