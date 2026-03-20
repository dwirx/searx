# 🚀 Multi-Engine Search & Info-Center CLI (Go)

A powerful, modular, and distraction-free CLI tool written in Go. Search across multiple engines, track real-time prediction markets, manage RSS feeds, and read articles directly in your terminal with automatic paywall and bot-protection bypass.

---

## ✨ Features

- **🔍 Multi-Engine Search**: Support for DuckDuckGo, Mojeek, Google, Brave, and SearX.
- **📊 Real-time Markets**: Integration with **Polymarket** featuring color-coded price movements (↑/↓) and categorical filtering.
- **📰 RSS Reader**: Full-featured RSS/Atom manager with global search, source filtering, and automated feed validation.
- **⚖️ Hukum Indonesia**: Integration with **Pasal.id** for full-text search across Indonesian laws and regulations.
- **📖 Smart Reader Mode**: Automatically extracts clean, Markdown-like article content for a distraction-free reading experience.
- **🛡️ Auto-Bypass**: Intelligent fallbacks using [Lightpanda](https://lightpanda.io/) (Headless Browser) and Archive.today to bypass paywalls (NYT, WSJ, Bloomberg, etc.).
- **💾 Save to Markdown**: Save any extracted article directly to a `.md` file for your personal knowledge base.
- **⏳ Visual Feedback**: Modern UI with loading spinners and ANSI colors for high terminal productivity.

---

## 🛠 Installation

### ⚡ Quick Install (Linux & macOS)
```bash
curl -sSL https://github.com/dwirx/searx/releases/latest/download/install.sh | bash
```

### 🔨 Build From Source
```bash
git clone https://github.com/dwirx/searx
cd searx
go build -o search ./cmd/search
./search setup  # Install the Lightpanda helper
```

---

## 📖 Usage Guide

### 1. 🔍 Searching the Web
```bash
search "golang concurrency patterns"          # Default (DuckDuckGo)
search -e google "latest space news"          # Use Google
search -e mojeek "privacy tools"              # Use Mojeek (fast & independent)
```

### 2. 📊 Polymarket (Prediction Markets)
Track what the world thinks is going to happen in real-time.
```bash
search -market                       # View trending breaking news
search -market -cat crypto           # Filter by Crypto
search -market -cat politics         # Filter by Politics
```
*Supported categories: `trending`, `breaking`, `new`, `politics`, `crypto`, `sports`, `finance`, `geopolitics`, `tech`, `culture`, `weather`.*

### 3. ⚖️ Hukum Indonesia (Pasal.id)
Cari pasal atau undang-undang dengan mudah.
```bash
search -pasal "upah minimum"         # Cari tentang upah minimum
search -pasal "hak cipta"            # Cari tentang hak cipta
```

### 4. 📰 RSS Feed Manager
Stay updated with your favorite news sources.
```bash
search -rss                          # Read all subscribed feeds
search -rss "artificial intelligence" # Search for "AI" across all feeds
search -rss -source bloomberg        # Filter to only read Bloomberg
```

**Manage your feeds:**
```bash
search -add-rss techcrunch=https://techcrunch.com/feed/  # Add new source
search -del-rss bbc                                      # Remove source
search list-rss                                          # View all subscriptions
search check-rss                                         # Validate and cleanup broken feeds
```
*Configuration stored at: `~/.local/share/searx/rss.yaml`*

### 5. 📖 Reader Mode (Bypass Paywalls)
```bash
search -read "https://www.nytimes.com/..."        # Extract and read
search -read "https://go.dev/blog/..." -save      # Read and save to .md
```

---

## ⚙️ Command Options

| Flag | Description |
|------|-------------|
| `-e <name>` | Select search engine (`ddg`, `google`, `brave`, `mojeek`, `hn`, `searx`, `polymarket`, `rss`) |
| `-market` | Shortcut for Polymarket markets |
| `-cat <topic>` | Category for Polymarket (politics, crypto, etc.) |
| `-rss` | Read subscribed RSS feeds |
| `-source <name>` | Filter RSS results by source name |
| `-read <url>` | Extract and read article content |
| `-save` | Save the extracted content to a `.md` file |
| `-panda` | Force use of Lightpanda headless browser |
| `-archive` | Force use of `archive.today` prefix |
| `-hn <cat>` | Hacker News Category (`top`, `new`, `best`, `ask`, `show`, `job`) |

---

## 📂 Configuration
Your RSS feeds are stored in a human-readable YAML file:
📍 `~/.local/share/searx/rss.yaml`

**Default feeds included:**
Bloomberg, BBC, CNN, Reuters, The Verge, TechCrunch, Wired, Ars Technica, OpenAI, DeepMind, Al Jazeera, and more.

---

## 🏗 Modular Architecture
- `cmd/search/`: Main CLI entry point.
- `internal/engine/`: Search and data providers (RSS, Polymarket, Search Engines).
- `internal/reader/`: Article extraction and paywall bypass logic.
- `internal/ui/`: Modern terminal interface and formatting.
- `internal/types/`: Common data structures.

---
*Created with focus on privacy, speed, and terminal productivity. Enjoy your distraction-free information stream!* 🌐✨
