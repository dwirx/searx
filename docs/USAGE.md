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
```
Supported categories: `trending`, `breaking`, `new`, `politics`, `crypto`, `sports`, `finance`, `geopolitics`, `tech`, `culture`, `weather`.

## ⚖️ Hukum Indonesia (Pasal.id)
Cari pasal atau undang-undang RI.

**Contoh:**
```bash
search -pasal "upah minimum"
search -pasal -law-type UU "cipta kerja"
```
Lihat [docs/PASAL.md](PASAL.md) untuk panduan lengkap.

## 📰 RSS Feeds
Manage and read your favorite news feeds directly from the CLI.

**Read all feeds:**
```bash
search -rss
```

**Search within feeds:**
```bash
search -rss "technology"
```

**Filter by source:**
```bash
search -rss -source bloomberg
```

**Manage feeds:**
```bash
search -add-rss name=url
search -del-rss name
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

## 🛠 Commands
- `search setup`: Install Lightpanda browser (Linux).
- `search check-rss`: Validate and cleanup broken feeds.
- `search list-rss`: List all subscriptions.
- `search update`: Update Search CLI and Lightpanda.
- `search version`: Show versions.
