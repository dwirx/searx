package main

import "testing"

func TestInstallerActionArgs(t *testing.T) {
	tests := []struct {
		name           string
		action         string
		keepLightpanda bool
		want           []string
		wantErr        bool
	}{
		{
			name:    "update",
			action:  "update",
			want:    []string{"--update"},
			wantErr: false,
		},
		{
			name:           "uninstall keep lightpanda",
			action:         "uninstall",
			keepLightpanda: true,
			want:           []string{"--uninstall", "--keep-lightpanda"},
			wantErr:        false,
		},
		{
			name:    "uninstall remove lightpanda",
			action:  "uninstall",
			want:    []string{"--uninstall"},
			wantErr: false,
		},
		{
			name:    "unknown action",
			action:  "noop",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := installerActionArgs(tc.action, tc.keepLightpanda)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("installerActionArgs(%q) error = nil, want error", tc.action)
				}
				return
			}
			if err != nil {
				t.Fatalf("installerActionArgs(%q) error = %v", tc.action, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("installerActionArgs(%q) len = %d, want %d", tc.action, len(got), len(tc.want))
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("installerActionArgs(%q)[%d] = %q, want %q", tc.action, i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestInstallerShellCommand(t *testing.T) {
	cmd, err := installerShellCommand("update", false)
	if err != nil {
		t.Fatalf("installerShellCommand(update) error = %v", err)
	}
	want := "curl -sSL " + installScriptURL + " | bash -s -- --update"
	if cmd != want {
		t.Fatalf("installerShellCommand(update) = %q, want %q", cmd, want)
	}
}
