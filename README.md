# Multi-Engine Search & Reader CLI (Go)

A powerful, modular, and distraction-free CLI search tool written in Go. Search across multiple engines, browse Hacker News, and read articles directly in your terminal with automatic paywall and bot-protection bypass.

## 🚀 Features

- **Multi-Engine Search**: Support for DuckDuckGo, Mojeek, Google (Lite), Brave, and SearX.
- **Official Hacker News API**: Browse `top`, `new`, `best`, `ask`, `show`, and `job` stories.
- **Smart Reader**: Automatically extracts clean article content (distraction-free).
- **Auto-Bypass**: 
    - Intelligent fallbacks: Standard Fetch ➔ Lightpanda Browser ➔ Archive.today.
    - Automatic paywall detection for NYT, WSJ, Bloomberg, etc.
- **Headless Browser Integration**: Uses [Lightpanda](https://lightpanda.io/) for high-performance, stealthy web fetching.
- **Save to Markdown**: Save extracted articles directly to `.md` files with sanitized filenames.

## 🛠 Installation

1. **Prerequisites**: [Go](https://go.dev/dl/) installed.
2. **Build the binary**:
   ```bash
   go build -o search cmd/search/main.go
   ```
3. **Setup Lightpanda** (Required for complex sites):
   The tool expects the `lightpanda` binary in the current directory.
   ```bash
   curl -L -o lightpanda https://github.com/lightpanda-io/browser/releases/download/v0.2.6/lightpanda-x86_64-linux && chmod +x ./lightpanda
   ```

## 📖 Usage Guide

### 1. Search
Search using the default engine (DuckDuckGo):
```bash
./search "golang generics"
```
Search using Mojeek (very fast and reliable):
```bash
./search -e mojeek "linux kernel internals"
```

### 2. Hacker News
Browse the best stories currently on HN:
```bash
./search -e hn -hn best
```

### 3. Read & Save Articles
Read an article directly in the terminal:
```bash
./search -read "https://go.dev/blog/go1.22"
```

**Save the article** to a Markdown file automatically:
```bash
./search -read "https://go.dev/blog/go1.22" -save
```
*Output: `[✔] Article saved to: go-122-is-released-the-go-programming-language.md`*

**Bypass Paywalls (NYT, etc.)**:
The tool will automatically try to use Archive.today and Lightpanda if it detects a known paywalled site.
```bash
./search -read "https://www.nytimes.com/..." -save
```

## ⚙️ Options

| Flag | Description |
|------|-------------|
| `-e` | Search engine (`ddg`, `google`, `brave`, `mojeek`, `hn`, `searx`) |
| `-read` | URL of the article to extract and read |
| `-save` | Save the extracted content to a `.md` file |
| `-panda` | Force use of Lightpanda headless browser |
| `-archive` | Force use of `archive.today` prefix |
| `-hn` | HN Category (`top`, `new`, `best`, `ask`, `show`, `job`) |
| `-i` | Custom Searx instance URL |

## 🏗 Modular Structure

- `cmd/search/`: Main entry point and flag handling.
- `internal/engine/`: Logic for different search engines.
- `internal/reader/`: Smart fetching and content extraction.
- `internal/ui/`: Formatting and CLI output.

---
*Created with focus on privacy, speed, and terminal productivity.*
