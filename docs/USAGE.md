# USAGE Guide

## 🔍 Search
Search using the default engine (DuckDuckGo):
```bash
search "your query"
```

Use a specific engine:
```bash
search -e google "your query"
search -e mojeek "your query"
```

## 📊 Polymarket (Real-time Markets)
Polymarket features real-time price tracking with color-coded 24h changes.

**Shortcut:**
```bash
search -market
```

**Specific Category:**
```bash
search -market -cat crypto
search -market -cat politics
search -market -cat sports
```
Supported categories: `trending`, `breaking`, `new`, `politics`, `crypto`, `sports`, `finance`, `geopolitics`, `tech`, `culture`, `weather`.

## 📰 RSS Feeds
Manage and read your favorite news feeds directly from the CLI.

**Configuration File:**
Your feeds are stored in YAML format at:
`~/.local/share/searx/rss.yaml`

**Read all feeds:**
```bash
search -rss
```

**Search within feeds:**
```bash
search -rss "technology"
```

**Manage feeds:**
```bash
search -add-rss techcrunch=https://techcrunch.com/feed/
search -del-rss bbc
search list-rss
search check-rss
```

## 📖 Reader Mode
Extract clean, distraction-free content from any URL.

**Read article:**
```bash
search -read "https://example.com/article"
```

**Read and save to Markdown:**
```bash
search -read "https://example.com/article" -save
```

**Force paywall bypass (Archive.today):**
```bash
search -read "https://nytimes.com/..." -archive
```

## 🛠 Commands
- `search setup`: Install Lightpanda browser (Linux).
- `search update`: Update Search CLI and Lightpanda.
- `search version`: Show versions.
