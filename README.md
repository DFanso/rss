# RSS Reader

A simple RSS reader and generator application built with Go.

## Features

- Fetch and parse RSS feeds
- Display feed content in a clean, modern UI
- Export feeds in RSS format
- Add, delete, and manage feeds
- **Persistent storage of feed subscriptions**
- **Always fetches fresh feed content** for up-to-date information
- Modern dark theme with Tailwind CSS
- Reactive UI with HTMX (no JavaScript frameworks needed)
- Mobile-responsive design
- Fast, lightweight frontend

## Requirements

- Go 1.16 or higher
- Web browser with JavaScript enabled

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/dfanso/rss.git
   cd rss
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Build the application:
   ```
   go build -o rss-reader ./cmd/rss
   ```

## Usage

Run the application:

```
./rss-reader
```

By default, the server will start on port 8080. You can specify a different port using the `-port` flag:

```
./rss-reader -port 3000
```

You can also add default feeds at startup using the `-feeds` flag:

```
./rss-reader -feeds "https://news.ycombinator.com/rss,https://www.reddit.com/.rss"
```

### Data Storage

The application stores your feed subscriptions in a JSON file for persistence between restarts. By default, subscriptions are stored in `data/feeds.json`. You can specify a different data directory using the `-data` flag:

```
./rss-reader -data /path/to/data
```

The application will automatically:
- Create the data directory if it doesn't exist
- Load saved feed subscriptions when starting
- Save feed subscriptions when they are added or removed
- **Always fetch the latest feed content** when you view a feed, ensuring you get the most up-to-date information

## Development

The project structure is as follows:

- `cmd/rss`: Main application entry point
- `src/parser`: RSS parsing and storage logic
- `src/server`: HTTP server and API endpoints with HTMX support
- `web/templates`: HTML templates with Tailwind CSS and HTMX
- `data`: Feed subscription storage (created at runtime)

### Technology Stack

- **Backend**: Go with Gin web framework
- **Frontend**: HTMX for reactive interactions
- **Styling**: Tailwind CSS for modern, responsive design
- **Storage**: JSON file-based persistence
- **Icons**: Bootstrap Icons

## API Endpoints

- `GET /`: Home page
- `GET /feeds`: List all feeds
- `POST /feeds`: Add a new feed
- `GET /feed?url=...`: Get a specific feed (always fetches fresh content)
- `DELETE /feed?url=...`: Remove a feed
- `GET /export?url=...`: Export a feed in RSS format

## License

This project is licensed under the MIT License - see the LICENSE file for details.
 
