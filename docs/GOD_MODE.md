# God Mode Guide (Multi-Engine Aggregator)

**God Mode** (`-god`) is the ultimate search mode in `searx-cli`. It aggregates results from 7 different search engines simultaneously while providing high-level protection against bot detection using the **Lightpanda** headless browser.

## 🌟 Why Use God Mode?
Most search engines (Google, Bing, Brave) aggressively block standard CLI/HTTP requests. God Mode bypasses these restrictions by:
1.  **Paralellization**: Searching 7 sources at once (**Google, Brave, DuckDuckGo, Yahoo, Mojeek, Bing, Ask.com**).
2.  **Anti-Bot Protection**: Using **Lightpanda** to fetch search result pages, which mimics real browser behavior and handles JavaScript challenges.
3.  **URL Deduplication**: Automatically merging results from all engines and removing duplicate links to provide a clean, unique list.

## 🚀 Usage
Simply use the `-god` flag before your query:
```bash
search -god "golang concurrency patterns"
```

## 🛠 Prerequisites
God Mode requires **Lightpanda** to be installed on your system (currently supported on Linux).

**Setup Lightpanda:**
```bash
search setup
```

## ⚙️ How it Works
When you run `-god`, the CLI:
1.  Spawns 7 concurrent goroutines.
2.  Each goroutine executes a `lightpanda fetch` command for its assigned engine.
3.  Parses the returned HTML using specialized CSS selectors for each site.
4.  Normalizes URLs (removes trailing slashes, etc.).
5.  Aggregates the final list and presents it in the terminal UI.

---
*Created for maximum coverage and reliability. Get the truth from all sources at once.* 🌐⚡
