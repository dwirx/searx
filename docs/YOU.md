# You.com AI Search Guide

`searx-cli` integrates with the **You.com API** to provide three distinct modes of operation: high-speed web search, agentic AI research with citations, and clean web content extraction.

## 🔑 Setup
To use You.com features, you must add your API key to the `.env` file:
```env
YOU_API_KEY=your_api_key_here
```

## 🚀 Search Modes

### 1. Web Search (`search`)
The default mode. Optimized for Large Language Models (LLMs), providing structured snippets and metadata from the web and live news.

**Usage:**
```bash
search -you "latest space news"
# Or explicitly:
search -you -you-mode search "latest space news"
```

### 2. Agentic Research (`research`)
Performs multi-step reasoning to synthesize information from multiple sources into a single, well-cited answer. Best for complex questions.

**Usage:**
```bash
search -you -you-mode research "What are the pros and cons of using Go for microservices in 2026?"
```
*Output includes a comprehensive narrative followed by a list of source URLs used.*

### 3. Web Contents Scraper (`contents`)
Retrieves the main content of specific web pages, automatically stripping away advertisements, navigation bars, and footers. Returns clean text/Markdown.

**Usage:**
```bash
search -you -you-mode contents "https://go.dev/blog/pgo"
```
*You can also pass multiple URLs separated by commas.*

## ⚙️ Advanced Options

| Flag | Description | Default |
|------|-------------|---------|
| `-you-mode <m>` | Choose between `search`, `research`, or `contents` | `search` |
| `-you-count <n>` | Number of results to return (Search mode only) | `10` |
| `-you-country <c>` | Geographic focus (e.g., `US`, `ID`) | System |
| `-you-lang <l>` | Result language (e.g., `en`, `id`) | `en` |

---
*Note: Research mode may take longer (up to 60s) as it performs real-time synthesis.*
