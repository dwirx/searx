package util

import "testing"

func TestNormalizeLightpandaVersion(t *testing.T) {
	cases := []struct {
		in       string
		expected string
		ok       bool
	}{
		{in: "v0.2.6", expected: "0.2.6", ok: true},
		{in: "0.2.6", expected: "0.2.6", ok: true},
		{in: "lightpanda 0.2.6", expected: "0.2.6", ok: true},
		{in: "Lightpanda version: v1.4.0", expected: "1.4.0", ok: true},
		{in: "edd0c5c8", expected: "", ok: false},
		{in: "unknown", expected: "", ok: false},
	}

	for _, tc := range cases {
		got, ok := normalizeLightpandaVersion(tc.in)
		if ok != tc.ok {
			t.Fatalf("normalizeLightpandaVersion(%q) ok=%v, want %v", tc.in, ok, tc.ok)
		}
		if got != tc.expected {
			t.Fatalf("normalizeLightpandaVersion(%q)=%q, want %q", tc.in, got, tc.expected)
		}
	}
}

func TestShouldAutoUpdateLightpanda(t *testing.T) {
	cases := []struct {
		name        string
		local       string
		latest      string
		recordedTag string
		want        bool
	}{
		{
			name:        "not installed must download",
			local:       "Not installed",
			latest:      "v0.2.6",
			recordedTag: "",
			want:        true,
		},
		{
			name:        "recorded tag match skip update",
			local:       "edd0c5c8",
			latest:      "v0.2.6",
			recordedTag: "v0.2.6",
			want:        false,
		},
		{
			name:        "semver match skip update",
			local:       "lightpanda 0.2.6",
			latest:      "v0.2.6",
			recordedTag: "",
			want:        false,
		},
		{
			name:        "unknown local version skip auto update",
			local:       "edd0c5c8",
			latest:      "v0.2.6",
			recordedTag: "",
			want:        false,
		},
		{
			name:        "known outdated version should update",
			local:       "v0.2.5",
			latest:      "v0.2.6",
			recordedTag: "",
			want:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldAutoUpdateLightpanda(tc.local, tc.latest, tc.recordedTag)
			if got != tc.want {
				t.Fatalf("shouldAutoUpdateLightpanda(%q, %q, %q)=%v, want %v", tc.local, tc.latest, tc.recordedTag, got, tc.want)
			}
		})
	}
}
