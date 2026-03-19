package main

import (
	"errors"
	"testing"
)

func TestShouldFallbackByStatusCode(t *testing.T) {
	if !shouldFallback("ddg", errors.New("status code 202")) {
		t.Fatalf("expected ddg to fallback on status code 202")
	}
	if !shouldFallback("searx", errors.New("status code 403")) {
		t.Fatalf("expected searx to fallback on status code 403")
	}
	if shouldFallback("brave", errors.New("status code 403")) {
		t.Fatalf("did not expect brave to fallback")
	}
	if shouldFallback("ddg", nil) {
		t.Fatalf("did not expect fallback for nil error")
	}
}

func TestFallbackEngineNames(t *testing.T) {
	cases := []struct {
		primary string
		want    []string
	}{
		{primary: "ddg", want: []string{"mojeek", "google", "brave"}},
		{primary: "searx", want: []string{"ddg", "mojeek", "google"}},
		{primary: "google", want: nil},
	}

	for _, tc := range cases {
		got := fallbackEngineNames(tc.primary)
		if len(got) != len(tc.want) {
			t.Fatalf("fallbackEngineNames(%q) len=%d, want %d", tc.primary, len(got), len(tc.want))
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("fallbackEngineNames(%q)[%d]=%q, want %q", tc.primary, i, got[i], tc.want[i])
			}
		}
	}
}
