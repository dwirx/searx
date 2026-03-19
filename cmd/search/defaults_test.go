package main

import "testing"

func TestDefaultEngine(t *testing.T) {
	if got := getDefaultEngine(); got != "ddg" {
		t.Fatalf("default engine = %q, want %q", got, "ddg")
	}
}
