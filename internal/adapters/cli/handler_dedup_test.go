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

// mockDeduplicator implements ports.Deduplicator for testing.
type mockDeduplicator struct {
	dedupeCalled    bool
	lastData        any
	returnData      any
	returnResult    domain.DedupResult
	shouldTransform bool // if true, returns different data
}

func (m *mockDeduplicator) Dedupe(data any) (any, domain.DedupResult) {
	m.dedupeCalled = true
	m.lastData = data
	if m.shouldTransform && m.returnData != nil {
		return m.returnData, m.returnResult
	}
	return data, m.returnResult
}

// TestHandler_DedupInPipeline tests that deduplication is applied after parsing in JSON mode.
func TestHandler_DedupInPipeline(t *testing.T) {
	// Given: Create handler with deduplicator and a parser returning structured data
	runner := &mockRunner{stdout: "line1\nline1\nline2\nline2\nline2\n"}

	// Create a parser that returns structured data
	parserResult := domain.NewParseResult(
		map[string]any{
			"lines": []any{"line1", "line1", "line2", "line2", "line2"},
		},
		"line1\nline1\nline2\nline2\nline2\n",
		0,
	)

	parser := &mockParser{
		matchCmd: "echo",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	deduper := &mockDeduplicator{
		shouldTransform: true,
		returnData: map[string]any{
			"lines": []any{"line1 (repeated 2 times)", "line2 (repeated 3 times)"},
		},
		returnResult: domain.NewDedupResult(5, 2),
	}

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// When: Execute command with --json flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"echo", "test", "--json"}, "", &buf)

	// Then: Deduplicator was called
	require.NoError(t, err)
	assert.True(t, deduper.dedupeCalled, "Dedupe should be called in JSON mode")
}

// TestHandler_DedupDisabled tests that deduplication is not applied when --disable-filter=dedupe.
func TestHandler_DedupDisabled(t *testing.T) {
	// Given: Create handler with deduplicator
	runner := &mockRunner{stdout: "line1\nline1\nline2\n"}
	registry := &mockRegistry{}
	deduper := &mockDeduplicator{
		shouldTransform: true,
		returnData:      map[string]any{"deduped": true},
		returnResult:    domain.NewDedupResult(3, 2),
	}

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// When: Execute command with --json --disable-filter=dedupe
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(
		context.Background(),
		[]string{"echo", "test", "--json", "--disable-filter=dedupe"},
		"",
		&buf,
	)

	// Then: Deduplicator should NOT be called
	require.NoError(t, err)
	assert.False(t, deduper.dedupeCalled, "Dedupe should NOT be called when disabled via flag")
}

// TestHandler_DedupDisabledViaEnv tests that deduplication is not applied when
// STRUCTURED_CLI_DISABLE_FILTER=dedupe is set.
func TestHandler_DedupDisabledViaEnv(t *testing.T) {
	// Given: Create handler with deduplicator
	runner := &mockRunner{stdout: "line1\nline1\nline2\n"}
	registry := &mockRegistry{}
	deduper := &mockDeduplicator{
		shouldTransform: true,
		returnData:      map[string]any{"deduped": true},
		returnResult:    domain.NewDedupResult(3, 2),
	}

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// When: Execute command with env var set
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"echo", "test", "--json"},
		"",       // envJSON
		"dedupe", // envDisableFilter
		&buf,
	)

	// Then: Deduplicator should NOT be called
	require.NoError(t, err)
	assert.False(t, deduper.dedupeCalled, "Dedupe should NOT be called when disabled via env")
}

// TestHandler_DedupStatsInOutput tests that dedupStats field is added to JSON output
// when deduplication is applied and there was a reduction.
func TestHandler_DedupStatsInOutput(t *testing.T) {
	// Given: Create handler with deduplicator that produces a reduction
	runner := &mockRunner{stdout: "line1\nline1\nline2\n"}

	// Create a parser that returns structured data
	parserResult := domain.NewParseResult(
		map[string]any{
			"lines": []any{"line1", "line1", "line2"},
		},
		"line1\nline1\nline2\n",
		0,
	)

	parser := &mockParser{
		matchCmd: "echo",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	deduper := &mockDeduplicator{
		shouldTransform: true,
		returnData: map[string]any{
			"lines": []any{
				map[string]any{"sample": "line1", "count": 2},
				"line2",
			},
		},
		returnResult: domain.NewDedupResult(3, 2),
	}

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// When: Execute command with --json flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"echo", "test", "--json"}, "", &buf)

	// Then: Output includes dedupStats field
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Verify dedupStats is present
	dedupStats, ok := output["dedupStats"]
	require.True(t, ok, "Output should include dedupStats field")

	statsMap, ok := dedupStats.(map[string]any)
	require.True(t, ok, "dedupStats should be a map")

	assert.Equal(t, float64(3), statsMap["original_count"])
	assert.Equal(t, float64(2), statsMap["deduped_count"])
	assert.Equal(t, "33%", statsMap["reduction"])
}

// TestHandler_DedupPassthrough tests that passthrough mode is unaffected by deduplication.
func TestHandler_DedupPassthrough(t *testing.T) {
	// Given: Create handler with deduplicator
	runner := &mockRunner{stdout: "line1\nline1\nline2\n"}
	registry := &mockRegistry{}
	deduper := &mockDeduplicator{
		shouldTransform: true,
		returnData:      map[string]any{"deduped": true},
		returnResult:    domain.NewDedupResult(3, 2),
	}

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// When: Execute command WITHOUT --json flag (passthrough mode)
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"echo", "test"}, "", &buf)

	// Then: Deduplicator should NOT be called in passthrough mode
	require.NoError(t, err)
	assert.False(t, deduper.dedupeCalled, "Dedupe should NOT be called in passthrough mode")

	// Output should be raw passthrough
	assert.Equal(t, "line1\nline1\nline2\n", buf.String())
}

// TestHandler_DedupWithSmallFilter tests that both filters work together without conflict.
func TestHandler_DedupWithSmallFilter(t *testing.T) {
	// Given: Create handler with both small filter and deduplicator
	runner := &mockRunner{stdout: "nothing to commit, working tree clean"}
	registry := &mockRegistry{}

	smallFilter := &mockSmallFilter{
		shouldFilterResult: true,
		filterResult: domain.SmallOutputResult{
			Status:  "clean",
			Summary: "nothing to commit, working tree clean",
		},
	}

	deduper := &mockDeduplicator{
		shouldTransform: false,
		returnResult:    domain.NewDedupResult(0, 0),
	}

	h := NewHandlerWithDeduplicator(runner, registry, nil, smallFilter, deduper)

	// When: Execute 'git status --json'
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status", "--json"}, "", &buf)

	// Then: Small filter is applied (takes precedence over dedup for terse output)
	require.NoError(t, err)
	assert.True(t, smallFilter.shouldFilterCalled, "SmallFilter.ShouldFilter should be called")

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Equal(t, "clean", output["status"])
}

// TestHandler_DedupNoReductionNoStats tests that dedupStats is not added when
// there's no actual reduction (all items unique).
func TestHandler_DedupNoReductionNoStats(t *testing.T) {
	// Given: Create handler with deduplicator that produces no reduction
	runner := &mockRunner{stdout: "line1\nline2\nline3\n"}

	// Create a parser that returns structured data
	parserResult := domain.NewParseResult(
		map[string]any{
			"lines": []any{"line1", "line2", "line3"},
		},
		"line1\nline2\nline3\n",
		0,
	)

	parser := &mockParser{
		matchCmd: "echo",
		matchSub: []string{"test"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	deduper := &mockDeduplicator{
		shouldTransform: false,
		returnResult:    domain.NewDedupResult(3, 3), // No reduction
	}

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// When: Execute command with --json flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"echo", "test", "--json"}, "", &buf)

	// Then: Output should NOT include dedupStats (no reduction)
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// dedupStats should NOT be present when there's no reduction
	_, ok := output["dedupStats"]
	assert.False(t, ok, "Output should NOT include dedupStats when there's no reduction")
}

// TestHandler_DedupAllFiltersDisabled tests that --disable-filter=all disables dedup.
func TestHandler_DedupAllFiltersDisabled(t *testing.T) {
	// Given: Create handler with deduplicator
	runner := &mockRunner{stdout: "line1\nline1\nline2\n"}
	registry := &mockRegistry{}
	deduper := &mockDeduplicator{
		shouldTransform: true,
		returnData:      map[string]any{"deduped": true},
		returnResult:    domain.NewDedupResult(3, 2),
	}

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// When: Execute command with --disable-filter=all
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(
		context.Background(),
		[]string{"echo", "test", "--json", "--disable-filter=all"},
		"",
		&buf,
	)

	// Then: Deduplicator should NOT be called
	require.NoError(t, err)
	assert.False(t, deduper.dedupeCalled, "Dedupe should NOT be called when all filters disabled")
}

// TestHandler_DedupNilDeduplicator tests that nil deduplicator is handled gracefully.
func TestHandler_DedupNilDeduplicator(t *testing.T) {
	// Given: Create handler WITHOUT deduplicator (nil)
	runner := &mockRunner{stdout: "line1\nline1\nline2\n"}
	registry := &mockRegistry{}

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, nil)

	// When: Execute command with --json flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"echo", "test", "--json"}, "", &buf)

	// Then: Should succeed without panic
	require.NoError(t, err)

	// Output should be raw wrapped JSON
	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Contains(t, output, "raw")
}

// Helper function to verify deduplicator is accessible via interface.
var _ ports.Deduplicator = (*mockDeduplicator)(nil)
