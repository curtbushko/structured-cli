// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestExtractStatsFlag_Present(t *testing.T) {
	// Arrange
	args := []string{"git", "--stats", "status"}

	// Act
	statsFound, remaining := ExtractStatsFlag(args)

	// Assert
	assert.True(t, statsFound, "should find --stats flag")
	assert.Equal(t, []string{"git", "status"}, remaining, "should remove --stats from args")
}

func TestExtractStatsFlag_Absent(t *testing.T) {
	// Arrange
	args := []string{"git", "status"}

	// Act
	statsFound, remaining := ExtractStatsFlag(args)

	// Assert
	assert.False(t, statsFound, "should not find --stats flag")
	assert.Equal(t, []string{"git", "status"}, remaining, "args should be unchanged")
}

func TestExtractStatsFlag_EmptyArgs(t *testing.T) {
	// Arrange
	args := []string{}

	// Act
	statsFound, remaining := ExtractStatsFlag(args)

	// Assert
	assert.False(t, statsFound)
	assert.Equal(t, []string{}, remaining)
}

func TestExtractStatsFlag_MultipleOccurrences(t *testing.T) {
	// Arrange
	args := []string{"--stats", "git", "--stats", "status"}

	// Act
	statsFound, remaining := ExtractStatsFlag(args)

	// Assert
	assert.True(t, statsFound)
	assert.Equal(t, []string{"git", "status"}, remaining)
}

func TestExtractThemeFlag_Present(t *testing.T) {
	// Arrange
	args := []string{"git", "--theme=dark", "status"}

	// Act
	themeName, remaining := ExtractThemeFlag(args)

	// Assert
	assert.Equal(t, "dark", themeName, "should extract theme name")
	assert.Equal(t, []string{"git", "status"}, remaining, "should remove --theme from args")
}

func TestExtractThemeFlag_Absent(t *testing.T) {
	// Arrange
	args := []string{"git", "status"}

	// Act
	themeName, remaining := ExtractThemeFlag(args)

	// Assert
	assert.Equal(t, "", themeName, "should return empty string when no --theme flag")
	assert.Equal(t, []string{"git", "status"}, remaining, "args should be unchanged")
}

func TestExtractThemeFlag_EmptyValue(t *testing.T) {
	// Arrange
	args := []string{"git", "--theme=", "status"}

	// Act
	themeName, remaining := ExtractThemeFlag(args)

	// Assert
	assert.Equal(t, "", themeName, "should return empty string for empty theme value")
	assert.Equal(t, []string{"git", "status"}, remaining)
}

func TestExtractThemeFlag_LightTheme(t *testing.T) {
	// Arrange
	args := []string{"--theme=light", "git", "status"}

	// Act
	themeName, remaining := ExtractThemeFlag(args)

	// Assert
	assert.Equal(t, "light", themeName)
	assert.Equal(t, []string{"git", "status"}, remaining)
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
