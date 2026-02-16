// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"testing"
)

func TestExtractJSONFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantJSON bool
		wantArgs []string
	}{
		{
			name:     "extracts --json flag from middle",
			args:     []string{"git", "--json", "status"},
			wantJSON: true,
			wantArgs: []string{"git", "status"},
		},
		{
			name:     "extracts --json flag from end",
			args:     []string{"git", "status", "--json"},
			wantJSON: true,
			wantArgs: []string{"git", "status"},
		},
		{
			name:     "extracts --json flag from start",
			args:     []string{"--json", "git", "status"},
			wantJSON: true,
			wantArgs: []string{"git", "status"},
		},
		{
			name:     "no --json flag present",
			args:     []string{"git", "status", "--short"},
			wantJSON: false,
			wantArgs: []string{"git", "status", "--short"},
		},
		{
			name:     "empty args",
			args:     []string{},
			wantJSON: false,
			wantArgs: []string{},
		},
		{
			name:     "only --json flag",
			args:     []string{"--json"},
			wantJSON: true,
			wantArgs: []string{},
		},
		{
			name:     "multiple --json flags",
			args:     []string{"--json", "git", "--json", "status"},
			wantJSON: true,
			wantArgs: []string{"git", "status"},
		},
		{
			name:     "--json in different position with other flags",
			args:     []string{"git", "log", "--json", "--oneline", "-n", "5"},
			wantJSON: true,
			wantArgs: []string{"git", "log", "--oneline", "-n", "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJSON, gotArgs := ExtractJSONFlag(tt.args)

			if gotJSON != tt.wantJSON {
				t.Errorf("ExtractJSONFlag() json = %v, want %v", gotJSON, tt.wantJSON)
			}

			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("ExtractJSONFlag() args len = %d, want %d", len(gotArgs), len(tt.wantArgs))
				return
			}

			for i, arg := range gotArgs {
				if arg != tt.wantArgs[i] {
					t.Errorf("ExtractJSONFlag() args[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestShouldOutputJSON(t *testing.T) {
	tests := []struct {
		name     string
		flagJSON bool
		envValue string
		want     bool
	}{
		{
			name:     "flag true overrides env false",
			flagJSON: true,
			envValue: "false",
			want:     true,
		},
		{
			name:     "flag true with no env",
			flagJSON: true,
			envValue: "",
			want:     true,
		},
		{
			name:     "flag false with env true",
			flagJSON: false,
			envValue: "true",
			want:     true,
		},
		{
			name:     "flag false with env false",
			flagJSON: false,
			envValue: "false",
			want:     false,
		},
		{
			name:     "flag false with no env (default passthrough)",
			flagJSON: false,
			envValue: "",
			want:     false,
		},
		{
			name:     "env TRUE (case insensitive)",
			flagJSON: false,
			envValue: "TRUE",
			want:     true,
		},
		{
			name:     "env 1 treated as true",
			flagJSON: false,
			envValue: "1",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldOutputJSON(tt.flagJSON, tt.envValue)
			if got != tt.want {
				t.Errorf("ShouldOutputJSON(%v, %q) = %v, want %v", tt.flagJSON, tt.envValue, got, tt.want)
			}
		})
	}
}
