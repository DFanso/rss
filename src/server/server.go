package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/user/rss/src/parser"
)

// Server represents the RSS server
type Server struct {
	router  *gin.Engine
	storage *parser.Storage
}

// NewServer creates a new server instance
func NewServer(storage *parser.Storage) *Server {
	router := gin.Default()
	server := &Server{
		router:  router,
		storage: storage,
	}

	// Set up routes - using query parameters instead of path parameters for URLs
	router.GET("/", server.homePage)
	router.GET("/feeds", server.listFeeds)
	router.POST("/feeds", server.addFeed)
	router.GET("/feed", server.getFeed)       // Changed to /feed?url=...
	router.DELETE("/feed", server.removeFeed) // Changed to /feed?url=...
	router.GET("/export", server.exportFeed)  // Changed to /export?url=...

	// Serve static files
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	return server
}

// Start starts the server
func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

// homePage handles the home page
func (s *Server) homePage(c *gin.Context) {
	feeds := s.storage.GetAllFeeds()
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "RSS Reader",
		"feeds": feeds,
	})
}

// listFeeds handles listing all feeds
func (s *Server) listFeeds(c *gin.Context) {
	feeds := s.storage.GetAllFeeds()
	c.JSON(http.StatusOK, feeds)
}

// addFeed handles adding a new feed
type addFeedRequest struct {
	URL string `json:"url" binding:"required"`
}

func (s *Server) addFeed(c *gin.Context) {
	var req addFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log the URL we're trying to add
	gin.DefaultWriter.Write([]byte("Adding feed URL: " + req.URL + "\n"))

	feed, err := parser.FetchFeed(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = s.storage.AddFeed(feed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log storage contents after adding
	feeds := s.storage.GetAllFeeds()
	gin.DefaultWriter.Write([]byte(fmt.Sprintf("Storage now contains %d feeds\n", len(feeds))))
	for _, f := range feeds {
		gin.DefaultWriter.Write([]byte(fmt.Sprintf("  - %s\n", f.URL)))
	}

	c.JSON(http.StatusCreated, feed)
}

// getFeed handles getting a feed by URL
func (s *Server) getFeed(c *gin.Context) {
	// Use query parameter instead of path parameter
	feedURL := c.Query("url")
	if feedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	// Log the URL we're looking for
	gin.DefaultWriter.Write([]byte("Getting feed URL: " + feedURL + "\n"))

	feed, err := s.storage.GetFeed(feedURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, feed)
}

// removeFeed handles removing a feed
func (s *Server) removeFeed(c *gin.Context) {
	// Use query parameter instead of path parameter
	feedURL := c.Query("url")
	if feedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	err := s.storage.RemoveFeed(feedURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// exportFeed exports a feed in RSS format
func (s *Server) exportFeed(c *gin.Context) {
	// Use query parameter instead of path parameter
	feedURL := c.Query("url")
	if feedURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	feed, err := s.storage.GetFeed(feedURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Create a new RSS feed
	rssFeed := &feeds.Feed{
		Title:       feed.Title,
		Link:        &feeds.Link{Href: feed.URL},
		Description: feed.Description,
		Created:     feed.UpdatedAt,
	}

	rssFeed.Items = make([]*feeds.Item, 0, len(feed.Items))
	for _, item := range feed.Items {
		rssFeed.Items = append(rssFeed.Items, &feeds.Item{
			Title:       item.Title,
			Link:        &feeds.Link{Href: item.Link},
			Description: item.Description,
			Content:     item.Content,
			Created:     item.PublishedAt,
			Id:          item.GUID,
		})
	}

	rss, err := rssFeed.ToRss()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/rss+xml")
	c.String(http.StatusOK, rss)
}
