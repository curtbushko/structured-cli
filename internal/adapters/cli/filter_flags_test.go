// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractDisableFilter_ParsesFlag tests that ExtractDisableFilter extracts
// the --disable-filter flag and returns the filter names.
func TestExtractDisableFilter_ParsesFlag(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantFilters   []string
		wantRemaining []string
	}{
		{
			name:          "extracts single filter",
			args:          []string{"git", "--disable-filter=small", "status"},
			wantFilters:   []string{"small"},
			wantRemaining: []string{"git", "status"},
		},
		{
			name:          "extracts multiple comma-separated filters",
			args:          []string{"git", "--disable-filter=small,success", "status"},
			wantFilters:   []string{"small", "success"},
			wantRemaining: []string{"git", "status"},
		},
		{
			name:          "extracts all filter",
			args:          []string{"git", "--disable-filter=all", "status"},
			wantFilters:   []string{"all"},
			wantRemaining: []string{"git", "status"},
		},
		{
			name:          "extracts three filters",
			args:          []string{"--disable-filter=small,success,dedupe", "git", "status"},
			wantFilters:   []string{"small", "success", "dedupe"},
			wantRemaining: []string{"git", "status"},
		},
		{
			name:          "no flag present",
			args:          []string{"git", "status", "--json"},
			wantFilters:   nil,
			wantRemaining: []string{"git", "status", "--json"},
		},
		{
			name:          "empty args",
			args:          []string{},
			wantFilters:   nil,
			wantRemaining: []string{},
		},
		{
			name:          "flag at end",
			args:          []string{"git", "status", "--disable-filter=small"},
			wantFilters:   []string{"small"},
			wantRemaining: []string{"git", "status"},
		},
		{
			name:          "flag at start",
			args:          []string{"--disable-filter=small", "git", "status"},
			wantFilters:   []string{"small"},
			wantRemaining: []string{"git", "status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilters, gotRemaining := ExtractDisableFilter(tt.args)

			assert.Equal(t, tt.wantFilters, gotFilters)
			assert.Equal(t, tt.wantRemaining, gotRemaining)
		})
	}
}

// TestShouldDisableFilter_ChecksFlagValue tests that ShouldDisableFilter
// correctly checks the flag value for a given filter name.
func TestShouldDisableFilter_ChecksFlagValue(t *testing.T) {
	tests := []struct {
		name       string
		filterName string
		filters    []string
		envValue   string
		want       bool
	}{
		{
			name:       "small filter in list",
			filterName: "small",
			filters:    []string{"small"},
			envValue:   "",
			want:       true,
		},
		{
			name:       "small filter not in list",
			filterName: "small",
			filters:    []string{"success"},
			envValue:   "",
			want:       false,
		},
		{
			name:       "all disables any filter",
			filterName: "small",
			filters:    []string{"all"},
			envValue:   "",
			want:       true,
		},
		{
			name:       "multiple filters contains target",
			filterName: "success",
			filters:    []string{"small", "success", "dedupe"},
			envValue:   "",
			want:       true,
		},
		{
			name:       "empty filters list",
			filterName: "small",
			filters:    nil,
			envValue:   "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldDisableFilter(tt.filterName, tt.filters, tt.envValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestShouldDisableFilter_HandlesEnvVar tests that ShouldDisableFilter
// checks the STRUCTURED_CLI_DISABLE_FILTER environment variable.
func TestShouldDisableFilter_HandlesEnvVar(t *testing.T) {
	tests := []struct {
		name       string
		filterName string
		filters    []string
		envValue   string
		want       bool
	}{
		{
			name:       "env var contains small",
			filterName: "small",
			filters:    nil,
			envValue:   "small",
			want:       true,
		},
		{
			name:       "env var contains multiple filters",
			filterName: "success",
			filters:    nil,
			envValue:   "small,success",
			want:       true,
		},
		{
			name:       "env var is all",
			filterName: "dedupe",
			filters:    nil,
			envValue:   "all",
			want:       true,
		},
		{
			name:       "env var does not contain filter",
			filterName: "small",
			filters:    nil,
			envValue:   "success,dedupe",
			want:       false,
		},
		{
			name:       "flag takes precedence over env",
			filterName: "small",
			filters:    []string{"small"},
			envValue:   "",
			want:       true,
		},
		{
			name:       "both flag and env set",
			filterName: "small",
			filters:    []string{"dedupe"},
			envValue:   "small",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldDisableFilter(tt.filterName, tt.filters, tt.envValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestExtractDisableFilter_WithEqualsSign tests edge cases with equals signs.
func TestExtractDisableFilter_WithEqualsSign(t *testing.T) {
	// Given: Args with --disable-filter=small
	args := []string{"git", "--disable-filter=small", "status"}

	// When: Extract disable filter
	filters, remaining := ExtractDisableFilter(args)

	// Then: Correctly parses the flag value
	require.Len(t, filters, 1)
	assert.Equal(t, "small", filters[0])
	assert.Equal(t, []string{"git", "status"}, remaining)
}
