package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

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
func (s *Server) addFeed(c *gin.Context) {
	url := c.PostForm("url")
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL parameter is required"})
		return
	}

	// Log the URL we're trying to add
	gin.DefaultWriter.Write([]byte("Adding feed URL: " + url + "\n"))

	feed, err := parser.FetchFeed(url)
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

	// Return HTML fragment for HTMX
	feedItemHTML := fmt.Sprintf(`
	<li class="feed-item hover:bg-dark-hover cursor-pointer transition-all duration-200 group border-b border-dark-border last:border-b-0"
		hx-get="/feed?url=%s"
		hx-target="#feed-content"
		hx-indicator="#loading-indicator"
		onclick="setActiveFeed(this)">
		<div class="px-4 lg:px-6 py-3 lg:py-4">
			<div class="flex items-start justify-between">
				<div class="flex items-start space-x-3 flex-1 min-w-0">
					<i class="bi bi-rss text-blue-400 flex-shrink-0 mt-0.5"></i>
					<div class="flex-1 min-w-0">
						<h3 class="text-dark-text font-medium text-sm leading-tight mb-1 line-clamp-2">%s</h3>
						<p class="text-dark-text-secondary text-xs truncate">%s</p>
					</div>
				</div>
				<div class="feed-actions flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity ml-2 flex-shrink-0">
					<a href="/export?url=%s" 
					   target="_blank" 
					   title="View RSS Feed"
					   onclick="event.stopPropagation()"
					   class="p-2 text-dark-text-secondary hover:text-blue-400 transition-colors rounded hover:bg-dark-hover">
						<i class="bi bi-box-arrow-up-right text-sm"></i>
					</a>
					<button hx-delete="/feed?url=%s"
							hx-target="closest li"
							hx-swap="outerHTML"
							hx-confirm="Are you sure you want to remove this feed subscription?"
							title="Delete Feed"
							onclick="event.stopPropagation()"
							class="p-2 text-dark-text-secondary hover:text-red-400 transition-colors rounded hover:bg-dark-hover">
						<i class="bi bi-trash text-sm"></i>
					</button>
				</div>
			</div>
		</div>
	</li>`, template.URLQueryEscaper(feed.URL), template.HTMLEscaper(feed.Title), template.URLQueryEscaper(feed.URL), template.URLQueryEscaper(feed.URL), template.URLQueryEscaper(feed.URL))

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, feedItemHTML)
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

	// Return HTML fragment for HTMX
	var feedContentHTML strings.Builder

	if len(feed.Items) > 0 {
		feedContentHTML.WriteString(`<div class="max-w-4xl mx-auto space-y-6">`)

		for _, item := range feed.Items {
			date := item.PublishedAt.Format("January 2, 2006 at 3:04 PM")
			content := item.Description
			if content == "" {
				content = item.Content
			}

			// Process content for dark mode
			content = processContentForDarkMode(content)

			feedContentHTML.WriteString(fmt.Sprintf(`
			<article class="bg-dark-card border border-dark-border rounded-lg p-6 mb-6 hover:shadow-lg transition-all duration-200 hover:border-blue-500">
				<h3 class="text-xl font-semibold text-dark-text mb-3 leading-tight">
					<a href="%s" target="_blank" rel="noopener noreferrer" 
					   class="hover:text-blue-400 transition-colors group flex items-start p-1 -m-1 rounded">
						<span class="flex-1">%s</span>
						<i class="bi bi-box-arrow-up-right ml-2 text-base opacity-60 group-hover:opacity-100 group-hover:text-blue-400 flex-shrink-0 mt-1 transition-all"></i>
					</a>
				</h3>
				<div class="flex items-center text-dark-text-secondary text-sm mb-4">
					<i class="bi bi-calendar mr-2 text-blue-400"></i>
					<span>%s</span>
				</div>
				<div class="feed-content text-dark-text prose-sm">
					%s
				</div>
			</article>`,
				template.HTMLEscaper(item.Link),
				template.HTMLEscaper(item.Title),
				date,
				content,
			))
		}

		feedContentHTML.WriteString(`</div>`)
	} else {
		feedContentHTML.WriteString(`
		<div class="flex flex-col items-center justify-center h-96 text-center text-dark-text-secondary">
			<i class="bi bi-info-circle text-6xl mb-4 text-yellow-400 opacity-50"></i>
			<p class="text-lg mb-2">No items found in this feed</p>
			<p class="text-sm opacity-75">This feed might be empty or temporarily unavailable</p>
		</div>`)
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, feedContentHTML.String())
}

// processContentForDarkMode processes HTML content to ensure visibility in dark mode
func processContentForDarkMode(content string) string {
	// Basic processing to improve dark mode compatibility
	content = strings.ReplaceAll(content, `style="color: white"`, `style="color: #e0e0e0"`)
	content = strings.ReplaceAll(content, `style="color: #ffffff"`, `style="color: #e0e0e0"`)
	content = strings.ReplaceAll(content, `style="background-color: white"`, `style="background-color: transparent"`)
	content = strings.ReplaceAll(content, `style="background-color: #ffffff"`, `style="background-color: transparent"`)

	// Remove problematic inline styles
	content = strings.ReplaceAll(content, `color: white;`, `color: #e0e0e0;`)
	content = strings.ReplaceAll(content, `color: #ffffff;`, `color: #e0e0e0;`)
	content = strings.ReplaceAll(content, `background-color: white;`, `background-color: transparent;`)
	content = strings.ReplaceAll(content, `background-color: #ffffff;`, `background-color: transparent;`)

	return content
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

	// Return empty content for HTMX to remove the element
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
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
