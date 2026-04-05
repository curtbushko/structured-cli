// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/application"
	"github.com/curtbushko/structured-cli/internal/domain"
)

// TestDedupe_EnabledByDefaultInJSONMode verifies that deduplication runs automatically
// in JSON mode when --disable-filter is not specified.
func TestDedupe_EnabledByDefaultInJSONMode(t *testing.T) {
	// Arrange: Create handler with deduplicator and mock parser returning duplicates
	runner := &mockRunner{stdout: "duplicate\nduplicate\nduplicate\n"}
	parser := &mockParser{
		matchCmd: "test",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"items": []any{
					map[string]any{"name": "item1"},
					map[string]any{"name": "item1"}, // duplicate
					map[string]any{"name": "item2"},
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Execute in JSON mode without --disable-filter
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "test"},
		"", // envJSON
		"", // envDisableFilter - empty means no filters disabled
		&buf,
	)

	// Assert: Deduplication should have occurred (items reduced)
	require.NoError(t, err)
	output := buf.String()

	// Should have dedupStats in output when reduction happened
	assert.Contains(t, output, "dedupStats", "output should contain dedup statistics")
	// The duplicate items should be collapsed
	assert.Contains(t, output, "count", "duplicates should be collapsed with count")
}

// TestDedupe_DisabledWithFlag verifies that --disable-filter=dedupe disables deduplication.
func TestDedupe_DisabledWithFlag(t *testing.T) {
	// Arrange: Handler with deduplicator
	runner := &mockRunner{stdout: "duplicate\n"}
	parser := &mockParser{
		matchCmd: "test",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"items": []any{
					map[string]any{"name": "item1"},
					map[string]any{"name": "item1"}, // duplicate
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Execute with --disable-filter=dedupe flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "--disable-filter=dedupe", "test"},
		"", // envJSON
		"", // envDisableFilter
		&buf,
	)

	// Assert: No deduplication should occur
	require.NoError(t, err)
	output := buf.String()

	// Should NOT have dedupStats (dedup was disabled)
	assert.NotContains(t, output, "dedupStats", "dedupStats should not appear when dedup disabled")
}

// TestDedupe_DisabledWithEnvVar verifies that STRUCTURED_CLI_DISABLE_FILTER=dedupe
// environment variable disables deduplication.
func TestDedupe_DisabledWithEnvVar(t *testing.T) {
	// Arrange: Handler with deduplicator
	runner := &mockRunner{stdout: "duplicate\n"}
	parser := &mockParser{
		matchCmd: "test",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"items": []any{
					map[string]any{"name": "item1"},
					map[string]any{"name": "item1"}, // duplicate
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Execute with env var set to disable dedupe
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "test"},
		"",       // envJSON
		"dedupe", // envDisableFilter - set via environment
		&buf,
	)

	// Assert: No deduplication should occur
	require.NoError(t, err)
	output := buf.String()

	// Should NOT have dedupStats (dedup was disabled via env var)
	assert.NotContains(t, output, "dedupStats", "dedupStats should not appear when dedup disabled via env")
}

// TestDedupe_DisabledWithAll verifies that --disable-filter=all disables all filters
// including deduplication.
func TestDedupe_DisabledWithAll(t *testing.T) {
	// Arrange: Handler with deduplicator
	runner := &mockRunner{stdout: "duplicate\n"}
	parser := &mockParser{
		matchCmd: "test",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"items": []any{
					map[string]any{"name": "item1"},
					map[string]any{"name": "item1"}, // duplicate
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Execute with --disable-filter=all
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "--disable-filter=all", "test"},
		"", // envJSON
		"", // envDisableFilter
		&buf,
	)

	// Assert: No deduplication should occur
	require.NoError(t, err)
	output := buf.String()

	// Should NOT have dedupStats (all filters disabled)
	assert.NotContains(t, output, "dedupStats", "dedupStats should not appear when all filters disabled")
}

// TestDedupe_CommaFilters verifies that comma-separated filters work correctly,
// e.g., --disable-filter=dedupe,small disables both filters.
func TestDedupe_CommaFilters(t *testing.T) {
	// Arrange: Handler with deduplicator
	runner := &mockRunner{stdout: "duplicate\n"}
	parser := &mockParser{
		matchCmd: "test",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"items": []any{
					map[string]any{"name": "item1"},
					map[string]any{"name": "item1"}, // duplicate
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Execute with comma-separated filters
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "--disable-filter=dedupe,small", "test"},
		"", // envJSON
		"", // envDisableFilter
		&buf,
	)

	// Assert: No deduplication should occur
	require.NoError(t, err)
	output := buf.String()

	// Should NOT have dedupStats (dedupe in comma list disabled it)
	assert.NotContains(t, output, "dedupStats", "dedupStats should not appear when dedupe in comma-separated list")
}

// TestDedupe_NotInPassthroughMode verifies that deduplication is not applied
// when not in JSON mode (passthrough mode).
func TestDedupe_NotInPassthroughMode(t *testing.T) {
	// Arrange: Handler with deduplicator
	runner := &mockRunner{stdout: "line1\nline1\nline2\n"}
	registry := &mockRegistry{} // No parser registered
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Execute without --json flag (passthrough mode)
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"test"}, // No --json flag
		"",               // envJSON empty
		"",               // envDisableFilter
		&buf,
	)

	// Assert: Output should be raw passthrough (no dedup applied)
	require.NoError(t, err)
	output := buf.String()

	// Should be raw output, not JSON
	assert.Equal(t, "line1\nline1\nline2\n", output, "passthrough mode should output raw, unmodified")
	assert.NotContains(t, output, "dedupStats", "dedupStats should not appear in passthrough mode")
}

// TestDedupe_FlagExtraction verifies the ExtractDisableFilter function correctly
// parses the --disable-filter flag with dedupe value.
func TestDedupe_FlagExtraction(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantFilters   []string
		wantRemaining []string
	}{
		{
			name:          "extracts dedupe filter",
			args:          []string{"git", "--disable-filter=dedupe", "status"},
			wantFilters:   []string{"dedupe"},
			wantRemaining: []string{"git", "status"},
		},
		{
			name:          "extracts dedupe with other filters",
			args:          []string{"--disable-filter=small,dedupe,success", "git", "log"},
			wantFilters:   []string{"small", "dedupe", "success"},
			wantRemaining: []string{"git", "log"},
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

// TestDedupe_ShouldDisableFilter verifies the ShouldDisableFilter function
// correctly handles the dedupe filter name.
func TestDedupe_ShouldDisableFilter(t *testing.T) {
	tests := []struct {
		name       string
		filterName string
		filters    []string
		envValue   string
		want       bool
	}{
		{
			name:       "dedupe disabled via flag",
			filterName: FilterNameDedupe,
			filters:    []string{"dedupe"},
			envValue:   "",
			want:       true,
		},
		{
			name:       "dedupe disabled via all",
			filterName: FilterNameDedupe,
			filters:    []string{"all"},
			envValue:   "",
			want:       true,
		},
		{
			name:       "dedupe disabled via env var",
			filterName: FilterNameDedupe,
			filters:    nil,
			envValue:   "dedupe",
			want:       true,
		},
		{
			name:       "dedupe disabled via env all",
			filterName: FilterNameDedupe,
			filters:    nil,
			envValue:   "all",
			want:       true,
		},
		{
			name:       "dedupe enabled by default",
			filterName: FilterNameDedupe,
			filters:    nil,
			envValue:   "",
			want:       false,
		},
		{
			name:       "dedupe enabled when other filter disabled",
			filterName: FilterNameDedupe,
			filters:    []string{"small"},
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
