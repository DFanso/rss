# RSS Reader

A simple RSS reader and generator application built with Go.

## Features

- Fetch and parse RSS feeds
- Display feed content in a clean, modern UI
- Export feeds in RSS format
- Add, delete, and manage feeds

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

## Development

The project structure is as follows:

- `cmd/rss`: Main application entry point
- `src/parser`: RSS parsing and storage logic
- `src/server`: HTTP server and API endpoints
- `web/templates`: HTML templates
- `web/static`: Static assets (CSS, JavaScript)

## API Endpoints

- `GET /feeds`: List all feeds
- `POST /feeds`: Add a new feed
- `GET /feeds/:url`: Get a specific feed
- `DELETE /feeds/:url`: Remove a feed
- `GET /export/:url`: Export a feed in RSS format

## License

This project is licensed under the MIT License - see the LICENSE file for details.
 
