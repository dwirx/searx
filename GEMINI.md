# GEMINI.md

## Project Overview

**searx-cli** is a powerful, modular, and distraction-free CLI search tool written in Go. It allows users to search across multiple engines (DuckDuckGo, Google, Brave, Mojeek, Hacker News, SearX) and read articles directly in the terminal, bypassing paywalls and bot protection using intelligent fallbacks like Lightpanda (a headless browser) and Archive.today.

### Main Technologies
- **Language**: Go (v1.25.5+)
- **HTML Parsing**: `github.com/PuerkitoBio/goquery`
- **Headless Browser**: [Lightpanda](https://lightpanda.io/) (used for stealthy web fetching and paywall bypass)
- **External Services**: Archive.today (for fallback reading)

### Architecture
- `cmd/search/`: Contains the main entry point (`main.go`), flag handling, and command-line interface logic.
- `internal/engine/`: Implements the `SearchEngine` interface for various search providers.
- `internal/reader/`: Handles article fetching and content extraction, including logic for standard HTTP requests, Lightpanda-based fetching, and Archive.today integration.
- `internal/ui/`: Manages formatting and displaying search results to the terminal.
- `internal/util/`: Provides utility functions for setup, environment configuration, and Lightpanda binary management.

## Building and Running

### Build from Source
To build the project, run:
```bash
go build -o search ./cmd/search
```

### Running the CLI
Run the compiled binary:
```bash
./search "your search query"
```
Or run directly using `go run`:
```bash
go run ./cmd/search/main.go "your search query"
```

### Key Commands
- `search <query>`: Search using the default engine (DuckDuckGo).
- `search -e <engine> <query>`: Search using a specific engine (e.g., `mojeek`, `google`, `hn`, `polymarket`, `pasal`).
- `search -market [-cat <topic>]`: Shortcut for Polymarket markets with real-time price tracking and color coding.
- `search -pasal <query>`: Shortcut for Indonesian Laws search (pasal.id).
- `search -rss [-source <name>] [<query>]`: Read and search through subscribed RSS feeds.
    - Specific Source: Use `-source bloomberg` to only read a specific feed.
    - Configuration: `~/.local/share/searx/rss.yaml`
    - Manage: `-add-rss name=url`, `-del-rss name`, `list-rss`.
- `search check-rss`: Validate all subscribed feeds and automatically remove broken ones.
- `search list-rss`: List all currently subscribed RSS feeds.
- `search -read <url> [-save]`: Extract, read, and optionally save content to Markdown.
- `search setup`: Download and install the Lightpanda browser (Linux only).
- `search update`: Update the Search CLI and Lightpanda.
- `search version`: Display version information for the CLI and Lightpanda.

## Configuration

RSS feeds are stored in YAML format in the user's local share directory:
`~/.local/share/searx/rss.yaml`

Default feeds included: Bloomberg, BBC, CNN, Reuters, The Verge, TechCrunch, Guardian.


## Testing

The project uses Go's built-in testing framework. To run all tests, execute:
```bash
go test ./...
```
Tests are located alongside the source code in their respective directories (e.g., `cmd/search/`, `internal/engine/`, `internal/util/`).

## Development Conventions

- **Modularity**: New search engines should implement the `SearchEngine` interface in `internal/engine/`.
- **Error Handling**: Use intelligent fallbacks for network-related errors (e.g., falling back to Mojeek if DuckDuckGo blocks the request).
- **Environment Variables**:
    - `SEARX_LIGHTPANDA_PATH`: Custom path to the Lightpanda binary.
    - `SEARX_INSTALL_DIR`: Installation directory for the binary.
    - `SEARX_SKIP_SETUP`: Set to `1` to skip automatic setup during installation.
- **Content Extraction**: The `internal/reader/` package uses aggressive selectors to strip junk (ads, nav, footers) and extract clean Markdown-like content.
