// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// mockSuccessFilter implements ports.SuccessFilter for testing.
type mockSuccessFilter struct {
	shouldFilterCalled bool
	filterCalled       bool
	shouldFilterResult bool
	returnData         any
	returnStats        domain.SuccessFilterResult
	lastCmd            string
	lastSubcmds        []string
	lastData           any
}

func (m *mockSuccessFilter) ShouldFilter(cmd string, subcmds []string) bool {
	m.shouldFilterCalled = true
	m.lastCmd = cmd
	m.lastSubcmds = subcmds
	return m.shouldFilterResult
}

func (m *mockSuccessFilter) Filter(data any) (any, domain.SuccessFilterResult) {
	m.filterCalled = true
	m.lastData = data
	if m.returnData != nil {
		return m.returnData, m.returnStats
	}
	return data, m.returnStats
}

// Verify mockSuccessFilter implements ports.SuccessFilter
var _ ports.SuccessFilter = (*mockSuccessFilter)(nil)

// TestHandler_SuccessFilterDisabledPassesThrough tests that when success filter is disabled,
// data passes through unchanged.
func TestHandler_SuccessFilterDisabledPassesThrough(t *testing.T) {
	// Given: Create handler with success filter that should filter
	runner := &mockRunner{stdout: "test output"}

	parserResult := domain.NewParseResult(
		map[string]any{
			"tests": []any{
				map[string]any{"name": "test1", "status": "passed"},
				map[string]any{"name": "test2", "status": "failed"},
			},
		},
		"test output",
		0,
	)

	parser := &mockParser{
		matchCmd: "go",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	successFilter := &mockSuccessFilter{
		shouldFilterResult: true,
		returnData: map[string]any{
			"tests": []any{
				map[string]any{"name": "test2", "status": "failed"},
			},
		},
		returnStats: domain.NewSuccessFilterResult(2, 1, 1, 0, 1, 1),
	}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, nil, successFilter)

	// When: Execute command with --json --disable-filter=success
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(
		context.Background(),
		[]string{"go", "test", "--json", "--disable-filter=success"},
		"",
		&buf,
	)

	// Then: Filter should NOT be called (disabled via flag)
	require.NoError(t, err)
	assert.False(t, successFilter.filterCalled, "Filter should NOT be called when disabled via flag")
}

// TestHandler_SuccessFilterFiltersTestOutput tests that success filter filters test output.
func TestHandler_SuccessFilterFiltersTestOutput(t *testing.T) {
	// Given: Create handler with success filter for test command
	runner := &mockRunner{stdout: "test output"}

	parserResult := domain.NewParseResult(
		map[string]any{
			"tests": []any{
				map[string]any{"name": "test1", "status": "passed"},
				map[string]any{"name": "test2", "status": "passed"},
				map[string]any{"name": "test3", "status": "failed"},
			},
		},
		"test output",
		0,
	)

	parser := &mockParser{
		matchCmd: "go",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	successFilter := &mockSuccessFilter{
		shouldFilterResult: true,
		returnData: map[string]any{
			"tests": []any{
				map[string]any{"name": "test3", "status": "failed"},
			},
		},
		returnStats: domain.NewSuccessFilterResult(3, 2, 1, 0, 2, 1),
	}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, nil, successFilter)

	// When: Execute command with --json flag (no disable)
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"go", "test", "--json"}, "", &buf)

	// Then: Filter should be called and stats should appear in output
	require.NoError(t, err)
	assert.True(t, successFilter.shouldFilterCalled, "ShouldFilter should be called")
	assert.True(t, successFilter.filterCalled, "Filter should be called")

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Verify successFilterStats is present
	stats, ok := output["successFilterStats"]
	require.True(t, ok, "Output should include successFilterStats field")

	statsMap, ok := stats.(map[string]any)
	require.True(t, ok, "successFilterStats should be a map")
	assert.Equal(t, float64(3), statsMap["total"])
	assert.Equal(t, float64(2), statsMap["passed"])
	assert.Equal(t, float64(1), statsMap["failed"])
}

// TestHandler_SuccessFilterDisabledViaEnv tests that STRUCTURED_CLI_DISABLE_FILTER=success
// disables the success filter.
func TestHandler_SuccessFilterDisabledViaEnv(t *testing.T) {
	// Given: Create handler with success filter
	runner := &mockRunner{stdout: "test output"}

	parserResult := domain.NewParseResult(
		map[string]any{
			"tests": []any{
				map[string]any{"name": "test1", "status": "passed"},
			},
		},
		"test output",
		0,
	)

	parser := &mockParser{
		matchCmd: "go",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	successFilter := &mockSuccessFilter{
		shouldFilterResult: true,
		returnStats:        domain.NewSuccessFilterResult(1, 1, 0, 0, 1, 0),
	}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, nil, successFilter)

	// When: Execute command with env var set
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"go", "test", "--json"},
		"",        // envJSON
		"success", // envDisableFilter
		&buf,
	)

	// Then: Filter should NOT be called
	require.NoError(t, err)
	assert.False(t, successFilter.filterCalled, "Filter should NOT be called when disabled via env")
}

// TestHandler_SuccessFilterDisableBoth tests that --disable-filter=success,dedupe disables both.
func TestHandler_SuccessFilterDisableBoth(t *testing.T) {
	// Given: Create handler with both success filter and deduplicator
	runner := &mockRunner{stdout: "test output"}

	parserResult := domain.NewParseResult(
		map[string]any{
			"tests": []any{
				map[string]any{"name": "test1", "status": "passed"},
			},
		},
		"test output",
		0,
	)

	parser := &mockParser{
		matchCmd: "go",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	successFilter := &mockSuccessFilter{
		shouldFilterResult: true,
		returnStats:        domain.NewSuccessFilterResult(1, 1, 0, 0, 1, 0),
	}

	deduper := &mockDeduplicator{
		shouldTransform: true,
		returnResult:    domain.NewDedupResult(2, 1),
	}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, deduper, successFilter)

	// When: Execute command with --disable-filter=success,dedupe
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(
		context.Background(),
		[]string{"go", "test", "--json", "--disable-filter=success,dedupe"},
		"",
		&buf,
	)

	// Then: Neither filter should be called
	require.NoError(t, err)
	assert.False(t, successFilter.filterCalled, "Success filter should NOT be called")
	assert.False(t, deduper.dedupeCalled, "Dedup should NOT be called")
}

// TestHandler_SuccessFilterStatsInOutput tests that filter stats appear in JSON output.
func TestHandler_SuccessFilterStatsInOutput(t *testing.T) {
	// Given: Create handler with success filter that reduces output
	runner := &mockRunner{stdout: "test output"}

	parserResult := domain.NewParseResult(
		map[string]any{
			"tests": []any{
				map[string]any{"name": "test1", "status": "passed"},
				map[string]any{"name": "test2", "status": "failed"},
			},
		},
		"test output",
		0,
	)

	parser := &mockParser{
		matchCmd: "go",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	successFilter := &mockSuccessFilter{
		shouldFilterResult: true,
		returnData: map[string]any{
			"tests": []any{
				map[string]any{"name": "test2", "status": "failed"},
			},
		},
		returnStats: domain.NewSuccessFilterResult(2, 1, 1, 0, 1, 1),
	}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, nil, successFilter)

	// When: Execute command with --json flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"go", "test", "--json"}, "", &buf)

	// Then: Output includes successFilterStats
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	stats, ok := output["successFilterStats"]
	require.True(t, ok, "Output should include successFilterStats")

	statsMap, ok := stats.(map[string]any)
	require.True(t, ok, "successFilterStats should be a map")

	assert.Equal(t, float64(2), statsMap["total"])
	assert.Equal(t, float64(1), statsMap["passed"])
	assert.Equal(t, float64(1), statsMap["failed"])
	assert.Equal(t, float64(0), statsMap["skipped"])
	assert.Equal(t, float64(1), statsMap["removed"])
	assert.Equal(t, float64(1), statsMap["kept"])
}

// TestHandler_SuccessFilterNoRemovalNoStats tests that stats are not added when no items removed.
func TestHandler_SuccessFilterNoRemovalNoStats(t *testing.T) {
	// Given: Create handler with success filter that removes nothing
	runner := &mockRunner{stdout: "test output"}

	parserResult := domain.NewParseResult(
		map[string]any{
			"tests": []any{
				map[string]any{"name": "test1", "status": "failed"},
			},
		},
		"test output",
		0,
	)

	parser := &mockParser{
		matchCmd: "go",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	successFilter := &mockSuccessFilter{
		shouldFilterResult: true,
		returnData: map[string]any{
			"tests": []any{
				map[string]any{"name": "test1", "status": "failed"},
			},
		},
		returnStats: domain.NewSuccessFilterResult(1, 0, 1, 0, 0, 1), // Nothing removed
	}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, nil, successFilter)

	// When: Execute command with --json flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"go", "test", "--json"}, "", &buf)

	// Then: Output should NOT include successFilterStats (no removal)
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	_, ok := output["successFilterStats"]
	assert.False(t, ok, "Output should NOT include successFilterStats when nothing removed")
}

// TestHandler_SuccessFilterNilFilter tests that nil success filter is handled gracefully.
func TestHandler_SuccessFilterNilFilter(t *testing.T) {
	// Given: Create handler WITHOUT success filter (nil)
	runner := &mockRunner{stdout: "test output"}
	registry := &mockRegistry{}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, nil, nil)

	// When: Execute command with --json flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"echo", "test", "--json"}, "", &buf)

	// Then: Should succeed without panic
	require.NoError(t, err)
}

// TestHandler_SuccessFilterPassthrough tests that passthrough mode is unaffected.
func TestHandler_SuccessFilterPassthrough(t *testing.T) {
	// Given: Create handler with success filter
	runner := &mockRunner{stdout: "test output"}
	registry := &mockRegistry{}
	successFilter := &mockSuccessFilter{
		shouldFilterResult: true,
		returnStats:        domain.NewSuccessFilterResult(1, 1, 0, 0, 1, 0),
	}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, nil, successFilter)

	// When: Execute command WITHOUT --json flag (passthrough mode)
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"echo", "test"}, "", &buf)

	// Then: Filter should NOT be called in passthrough mode
	require.NoError(t, err)
	assert.False(t, successFilter.shouldFilterCalled, "ShouldFilter should NOT be called in passthrough")
	assert.False(t, successFilter.filterCalled, "Filter should NOT be called in passthrough")

	// Output should be raw passthrough
	assert.Equal(t, "test output", buf.String())
}

// TestHandler_SuccessFilterNotApplicable tests that filter is not applied for non-test commands.
func TestHandler_SuccessFilterNotApplicable(t *testing.T) {
	// Given: Create handler with success filter that returns false for ShouldFilter
	runner := &mockRunner{stdout: "file list output"}

	parserResult := domain.NewParseResult(
		map[string]any{
			"files": []any{"file1.txt", "file2.txt"},
		},
		"file list output",
		0,
	)

	parser := &mockParser{
		matchCmd: "ls",
		matchSub: []string{},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	successFilter := &mockSuccessFilter{
		shouldFilterResult: false, // Not applicable to this command
	}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, nil, successFilter)

	// When: Execute command with --json flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"ls", "--json"}, "", &buf)

	// Then: Filter should NOT be called (not applicable)
	require.NoError(t, err)
	assert.True(t, successFilter.shouldFilterCalled, "ShouldFilter should be called to check applicability")
	assert.False(t, successFilter.filterCalled, "Filter should NOT be called when not applicable")
}

// TestHandler_SuccessFilterChainWithDedup tests that success filter chains with dedup correctly.
func TestHandler_SuccessFilterChainWithDedup(t *testing.T) {
	// Given: Create handler with both success filter and deduplicator
	runner := &mockRunner{stdout: "test output"}

	parserResult := domain.NewParseResult(
		map[string]any{
			"tests": []any{
				map[string]any{"name": "test1", "status": "passed"},
				map[string]any{"name": "test2", "status": "failed"},
				map[string]any{"name": "test3", "status": "failed"},
			},
		},
		"test output",
		0,
	)

	parser := &mockParser{
		matchCmd: "go",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	// Success filter removes passed tests
	successFilter := &mockSuccessFilter{
		shouldFilterResult: true,
		returnData: map[string]any{
			"tests": []any{
				map[string]any{"name": "test2", "status": "failed"},
				map[string]any{"name": "test3", "status": "failed"},
			},
		},
		returnStats: domain.NewSuccessFilterResult(3, 1, 2, 0, 1, 2),
	}

	// Deduplicator further reduces duplicates
	deduper := &mockDeduplicator{
		shouldTransform: true,
		returnData: map[string]any{
			"tests": []any{
				map[string]any{"name": "test2", "status": "failed", "count": 2},
			},
		},
		returnResult: domain.NewDedupResult(2, 1),
	}

	h := NewHandlerWithSuccessFilter(runner, registry, nil, nil, deduper, successFilter)

	// When: Execute command with --json flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"go", "test", "--json"}, "", &buf)

	// Then: Both filters should be called in order
	require.NoError(t, err)
	assert.True(t, successFilter.filterCalled, "Success filter should be called")
	assert.True(t, deduper.dedupeCalled, "Dedup should be called after success filter")

	// Output should have both stats
	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	_, hasSuccessStats := output["successFilterStats"]
	_, hasDedupStats := output["dedupStats"]
	assert.True(t, hasSuccessStats, "Should have successFilterStats")
	assert.True(t, hasDedupStats, "Should have dedupStats")
}
