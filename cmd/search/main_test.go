package main

import "testing"

func TestBuildReaderURL_ForceArchive(t *testing.T) {
	raw := "https://example.com/article"
	got := buildReaderURL(raw, true)
	want := "https://archive.today/https://example.com/article"
	if got != want {
		t.Fatalf("buildReaderURL() = %q, want %q", got, want)
	}
}

func TestBuildReaderURL_AutoArchiveForPaywallDomain(t *testing.T) {
	raw := "https://www.nytimes.com/2026/03/20/us/politics/example.html"
	got := buildReaderURL(raw, false)
	want := "https://archive.today/https://www.nytimes.com/2026/03/20/us/politics/example.html"
	if got != want {
		t.Fatalf("buildReaderURL() = %q, want %q", got, want)
	}
}

func TestBuildReaderURL_KeepArchiveURLAsIs(t *testing.T) {
	raw := "https://archive.today/https://www.wsj.com/articles/example"
	got := buildReaderURL(raw, false)
	if got != raw {
		t.Fatalf("buildReaderURL() = %q, want %q", got, raw)
	}
}

func TestSanitizeFileName(t *testing.T) {
	cases := map[string]string{
		"Go 1.22 is Released!":      "go-1-22-is-released",
		"***":                       "article",
		"  Already-slugged  title ": "already-slugged-title",
	}

	for input, want := range cases {
		got := sanitizeFileName(input)
		if got != want {
			t.Fatalf("sanitizeFileName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSelectEngine(t *testing.T) {
	tests := []struct {
		name      string
		engine    string
		hnCat     string
		instance  string
		wantName  string
		wantError bool
	}{
		{
			name:     "default ddg",
			engine:   "ddg",
			wantName: "DuckDuckGo (Scraper)",
		},
		{
			name:     "hn category",
			engine:   "hn",
			hnCat:    "best",
			wantName: "Hacker News API (best)",
		},
		{
			name:     "searx custom",
			engine:   "searx",
			instance: "https://searx.be",
			wantName: "Searx (https://searx.be)",
		},
		{
			name:      "unknown",
			engine:    "bing",
			wantError: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			e, err := selectEngine(tc.engine, tc.hnCat, tc.instance)
			if tc.wantError {
				if err == nil {
					t.Fatalf("selectEngine() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("selectEngine() unexpected error: %v", err)
			}
			if e.Name() != tc.wantName {
				t.Fatalf("selectEngine().Name() = %q, want %q", e.Name(), tc.wantName)
			}
		})
	}
}
