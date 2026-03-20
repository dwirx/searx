package engine

import (
	"testing"
)

func TestEngines(t *testing.T) {
	engines := []SearchEngine{
		&DuckDuckGoEngine{},
		&BraveEngine{},
		&ExaEngine{},
		&FirecrawlEngine{},
	}

	for _, e := range engines {
		t.Run(e.Name(), func(t *testing.T) {
			results, err := e.Search("golang")
			if err != nil {
				t.Logf("Warning: %s search failed (might be blocking): %v", e.Name(), err)
				return
			}
			if len(results) > 0 {
				t.Logf("%s returned %d results", e.Name(), len(results))
				if results[0].Title == "" {
					t.Errorf("%s first result has empty title", e.Name())
				}
				if results[0].URL == "" {
					t.Errorf("%s first result has empty URL", e.Name())
				}
			} else {
				t.Logf("No results from %s", e.Name())
			}
		})
	}
}
