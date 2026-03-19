package main

import "testing"

func TestParseSubcommandOptionsUpdate(t *testing.T) {
	opts, err := parseSubcommandOptions("update", []string{"--lightpanda-only"})
	if err != nil {
		t.Fatalf("parseSubcommandOptions(update) error = %v", err)
	}
	if !opts.LightpandaOnly {
		t.Fatalf("LightpandaOnly = false, want true")
	}
	if opts.KeepLightpanda {
		t.Fatalf("KeepLightpanda = true, want false")
	}
}

func TestParseSubcommandOptionsUninstall(t *testing.T) {
	opts, err := parseSubcommandOptions("uninstall", []string{"--keep-lightpanda"})
	if err != nil {
		t.Fatalf("parseSubcommandOptions(uninstall) error = %v", err)
	}
	if !opts.KeepLightpanda {
		t.Fatalf("KeepLightpanda = false, want true")
	}
	if opts.LightpandaOnly {
		t.Fatalf("LightpandaOnly = true, want false")
	}
}

func TestParseSubcommandOptionsRejectsUnexpectedArgs(t *testing.T) {
	_, err := parseSubcommandOptions("update", []string{"--lightpanda-only", "extra"})
	if err == nil {
		t.Fatalf("expected error for unexpected positional args")
	}
}

func TestParseSubcommandOptionsUnknownCommand(t *testing.T) {
	opts, err := parseSubcommandOptions("search", nil)
	if err != nil {
		t.Fatalf("parseSubcommandOptions(search) error = %v", err)
	}
	if opts.LightpandaOnly || opts.KeepLightpanda {
		t.Fatalf("expected zero options for unsupported command")
	}
}
