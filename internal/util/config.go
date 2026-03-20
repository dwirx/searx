package util

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	RSSFeeds map[string]string `yaml:"rss_feeds"`
}

func GetConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "searx")
}

func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "rss.yaml")
}

func DefaultConfig() *Config {
	return &Config{
		RSSFeeds: map[string]string{
			// --- WORLD NEWS & GEOPOLITICS ---
			"google_news_id": "https://news.google.com/rss?hl=en-ID&gl=ID&ceid=ID:en",
			"ap_top_news":    "https://news.google.com/rss/search?q=when:24h+source:Associated_Press&hl=en-US&gl=US&ceid=US:en",
			"reuters_world":  "https://news.google.com/rss/search?q=when:24h+source:Reuters&hl=en-US&gl=US&ceid=US:en",
			"bbc_world":      "https://feeds.bbci.co.uk/news/world/rss.xml",
			"guardian_world": "https://www.theguardian.com/world/rss",
			"aljazeera":      "https://www.aljazeera.com/xml/rss/all.xml",

			// --- MARKETS & FINANCE ---
			"bloomberg":     "https://finance.yahoo.com/news/bloomberg/",
			"yahoo_finance": "https://finance.yahoo.com/news/rssindex",
			"marketwatch":   "https://www.marketwatch.com/rss/topstories",
			"investing":     "https://www.investing.com/rss/news.rss",

			// --- AI & FUTURE TECH ---
			"openai":         "https://openai.com/news/rss.xml",
			"deepmind":       "https://deepmind.google/blog/rss.xml",
			"ai_news":        "https://www.artificialintelligence-news.com/feed/",
			"venturebeat_ai": "https://venturebeat.com/category/ai/feed/",

			// --- TECHNOLOGY ---
			"theverge":    "https://www.theverge.com/rss/index.xml",
			"techcrunch":  "https://techcrunch.com/feed/",
			"wired":       "https://www.wired.com/feed/rss",
			"arstechnica": "https://feeds.arstechnica.com/arstechnica/index",
			"hacker_news": "https://news.ycombinator.com/rss",
			
			// --- POPULAR BLOGS ---
			"simonwillison":   "https://simonwillison.net/atom/everything/",
			"jeffgeerling":    "https://www.jeffgeerling.com/blog.xml",
			"krebsonsecurity": "https://krebsonsecurity.com/feed/",
		},
	}
}

func LoadConfig() (*Config, error) {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		cfg := DefaultConfig()
		_ = SaveConfig(cfg)
		return cfg, nil
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("Warning: Failed to parse %s, using defaults. Error: %v\n", path, err)
		return DefaultConfig(), err
	}
	
	if cfg.RSSFeeds == nil {
		cfg.RSSFeeds = make(map[string]string)
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	path := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
