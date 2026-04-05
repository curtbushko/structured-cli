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

	"github.com/curtbushko/structured-cli/internal/application"
	"github.com/curtbushko/structured-cli/internal/domain"
)

// =============================================================================
// E2E Tests for Output Deduplication System
// These tests verify the complete deduplication pipeline from command execution
// to JSON output with deduplication.
// =============================================================================

// TestE2E_Dedup_EnabledByDefault verifies that deduplication is applied automatically
// in JSON mode when no flags are specified.
func TestE2E_Dedup_EnabledByDefault(t *testing.T) {
	// Arrange: Create handler with real deduplicator and mock parser returning duplicates
	runner := &mockRunner{stdout: "item1\nitem1\nitem2\n"}

	// Parser returns data with duplicates
	parser := &mockParser{
		matchCmd: "lint",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"errors": []any{
					map[string]any{"rule": "no-unused-vars", "file": "a.js", "line": float64(10)},
					map[string]any{"rule": "no-unused-vars", "file": "a.js", "line": float64(10)},
					map[string]any{"rule": "no-unused-vars", "file": "a.js", "line": float64(10)},
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
		[]string{"--json", "lint"},
		"", // envJSON
		"", // envDisableFilter - empty means dedup enabled
		&buf,
	)

	// Assert: Deduplication should occur - duplicates collapsed, stats present
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Should have dedupStats (reduction happened: 3 -> 1)
	assert.Contains(t, output, "dedupStats", "dedup should be enabled by default and add stats")
}

// TestE2E_Dedup_IdenticalObjectsCollapsed verifies that identical objects at the same
// array level are collapsed into a single object with a count field.
func TestE2E_Dedup_IdenticalObjectsCollapsed(t *testing.T) {
	// Arrange: Parser returns 3 identical objects
	runner := &mockRunner{stdout: "lint output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"errors": []any{
					map[string]any{"rule": "no-unused-vars", "file": "a.js", "line": float64(10)},
					map[string]any{"rule": "no-unused-vars", "file": "a.js", "line": float64(10)},
					map[string]any{"rule": "no-unused-vars", "file": "a.js", "line": float64(10)},
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "eslint"},
		"",
		"",
		&buf,
	)

	// Assert
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Check the errors array was collapsed
	errors, ok := output["errors"].([]any)
	require.True(t, ok, "errors should be an array")
	require.Len(t, errors, 1, "3 identical objects should collapse to 1")

	// Check the collapsed item has count field
	firstError, ok := errors[0].(map[string]any)
	require.True(t, ok, "error item should be a map")
	assert.Equal(t, float64(3), firstError["count"], "count should be 3")
	assert.Equal(t, "no-unused-vars", firstError["rule"], "original fields preserved")
}

// TestE2E_Dedup_DifferentObjectsPreserved verifies that unique/different objects
// are NOT collapsed and are preserved individually.
func TestE2E_Dedup_DifferentObjectsPreserved(t *testing.T) {
	// Arrange: Parser returns 3 different objects
	runner := &mockRunner{stdout: "lint output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"errors": []any{
					map[string]any{"rule": "no-unused-vars", "file": "a.js", "line": float64(10)},
					map[string]any{"rule": "no-unused-vars", "file": "b.js", "line": float64(20)},
					map[string]any{"rule": "semi", "file": "a.js", "line": float64(5)},
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "eslint"},
		"",
		"",
		&buf,
	)

	// Assert
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Check the errors array was NOT collapsed (all unique)
	errors, ok := output["errors"].([]any)
	require.True(t, ok, "errors should be an array")
	assert.Len(t, errors, 3, "3 different objects should remain as 3")

	// Check no count field was added (no deduplication happened)
	for i, err := range errors {
		errMap, ok := err.(map[string]any)
		require.True(t, ok, "error item should be a map")
		_, hasCount := errMap["count"]
		assert.False(t, hasCount, "unique item %d should not have count field", i)
	}

	// Should NOT have dedupStats (no reduction)
	_, hasDedupStats := output["dedupStats"]
	assert.False(t, hasDedupStats, "no dedupStats when all items unique")
}

// TestE2E_Dedup_RawTextCollapsed verifies that raw text deduplication collapses
// identical lines when the Deduper receives raw text directly (Stage 1).
// Note: This tests the Deduper component directly since raw text dedup
// is separate from the JSON object dedup in the pipeline.
func TestE2E_Dedup_RawTextCollapsed(t *testing.T) {
	// Arrange: Test the deduper directly with raw text (Stage 1 dedup)
	rawText := "error: something failed\nerror: something failed\nerror: something failed\nsuccess\n"
	deduper := application.NewDeduper()

	// Act: Call Dedupe directly with raw text string
	result, stats := deduper.Dedupe(rawText)

	// Assert
	resultStr, ok := result.(string)
	require.True(t, ok, "result should be a string")

	// Check that identical lines are collapsed
	assert.Contains(t, resultStr, "repeated 3 times", "should show repeat count for 3 identical lines")
	assert.Contains(t, resultStr, "error: something failed", "original text preserved")
	assert.Contains(t, resultStr, "success", "unique line preserved")

	// Check stats
	assert.Equal(t, 5, stats.OriginalCount, "4 lines + trailing newline = 5")
	assert.Equal(t, 3, stats.DedupedCount, "collapsed to 3 unique lines")
}

// TestE2E_Dedup_DisabledWithFlag verifies that --disable-filter=dedupe disables deduplication.
func TestE2E_Dedup_DisabledWithFlag(t *testing.T) {
	// Arrange: Parser returns duplicates
	runner := &mockRunner{stdout: "lint output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"errors": []any{
					map[string]any{"rule": "no-unused-vars", "file": "a.js"},
					map[string]any{"rule": "no-unused-vars", "file": "a.js"},
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Use --disable-filter=dedupe
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "--disable-filter=dedupe", "eslint"},
		"",
		"",
		&buf,
	)

	// Assert: No deduplication
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Should NOT have dedupStats (dedup disabled)
	_, hasDedupStats := output["dedupStats"]
	assert.False(t, hasDedupStats, "no dedupStats when dedup disabled via flag")

	// All duplicates should remain
	errors, ok := output["errors"].([]any)
	require.True(t, ok, "errors should be an array")
	assert.Len(t, errors, 2, "duplicates should remain when dedup disabled")
}

// TestE2E_Dedup_DisabledWithAll verifies that --disable-filter=all disables deduplication.
func TestE2E_Dedup_DisabledWithAll(t *testing.T) {
	// Arrange: Parser returns duplicates
	runner := &mockRunner{stdout: "lint output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"errors": []any{
					map[string]any{"rule": "semi", "file": "a.js"},
					map[string]any{"rule": "semi", "file": "a.js"},
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Use --disable-filter=all
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "--disable-filter=all", "eslint"},
		"",
		"",
		&buf,
	)

	// Assert: No deduplication
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// All duplicates should remain
	errors, ok := output["errors"].([]any)
	require.True(t, ok, "errors should be an array")
	assert.Len(t, errors, 2, "duplicates should remain when all filters disabled")
}

// TestE2E_Dedup_DisabledWithEnvVar verifies that STRUCTURED_CLI_DISABLE_FILTER=dedupe
// environment variable disables deduplication.
func TestE2E_Dedup_DisabledWithEnvVar(t *testing.T) {
	// Arrange: Parser returns duplicates
	runner := &mockRunner{stdout: "lint output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"errors": []any{
					map[string]any{"rule": "indent", "file": "a.js"},
					map[string]any{"rule": "indent", "file": "a.js"},
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Use env var to disable dedupe
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "eslint"},
		"",       // envJSON
		"dedupe", // envDisableFilter - set to "dedupe"
		&buf,
	)

	// Assert: No deduplication
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// All duplicates should remain
	errors, ok := output["errors"].([]any)
	require.True(t, ok, "errors should be an array")
	assert.Len(t, errors, 2, "duplicates should remain when dedup disabled via env")
}

// TestE2E_Dedup_FirstOccurrenceSample verifies that the first occurrence of a
// deduplicated item is kept as the sample (fields preserved).
func TestE2E_Dedup_FirstOccurrenceSample(t *testing.T) {
	// Arrange: Parser returns identical objects
	runner := &mockRunner{stdout: "lint output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"errors": []any{
					map[string]any{"rule": "no-unused-vars", "file": "a.js", "line": float64(10), "column": float64(5)},
					map[string]any{"rule": "no-unused-vars", "file": "a.js", "line": float64(10), "column": float64(5)},
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "eslint"},
		"",
		"",
		&buf,
	)

	// Assert
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Check the first occurrence fields are preserved
	errors, ok := output["errors"].([]any)
	require.True(t, ok, "errors should be an array")
	require.Len(t, errors, 1, "identical objects collapsed to 1")

	firstError, ok := errors[0].(map[string]any)
	require.True(t, ok, "error should be a map")

	// All original fields should be preserved
	assert.Equal(t, "no-unused-vars", firstError["rule"])
	assert.Equal(t, "a.js", firstError["file"])
	assert.Equal(t, float64(10), firstError["line"])
	assert.Equal(t, float64(5), firstError["column"])
	assert.Equal(t, float64(2), firstError["count"], "count should be 2")
}

// TestE2E_Dedup_DedupStatsPresent verifies that dedupStats field is added to output
// when deduplication actually reduces the item count.
func TestE2E_Dedup_DedupStatsPresent(t *testing.T) {
	// Arrange: Parser returns 100 items that dedupe to 3
	items := make([]any, 100)
	for i := 0; i < 100; i++ {
		ruleNum := i % 3 // Creates 3 unique items repeated ~33 times each
		items[i] = map[string]any{
			"rule": map[int]string{0: "rule-a", 1: "rule-b", 2: "rule-c"}[ruleNum],
			"file": "test.js",
		}
	}

	runner := &mockRunner{stdout: "lint output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result:   domain.NewParseResult(map[string]any{"errors": items}, "", 0),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "eslint"},
		"",
		"",
		&buf,
	)

	// Assert
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Check dedupStats is present
	dedupStats, ok := output["dedupStats"].(map[string]any)
	require.True(t, ok, "dedupStats should be present")

	assert.Equal(t, float64(100), dedupStats["original_count"])
	assert.Equal(t, float64(3), dedupStats["deduped_count"])
	assert.Equal(t, "97%", dedupStats["reduction"])
}

// TestE2E_Dedup_NestedLevelsSeparate verifies that objects at different nesting levels
// are NOT compared/deduplicated against each other.
func TestE2E_Dedup_NestedLevelsSeparate(t *testing.T) {
	// Arrange: Parser returns nested structure with similar objects at different levels
	runner := &mockRunner{stdout: "output"}
	parser := &mockParser{
		matchCmd: "tool",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"level1": []any{
					map[string]any{"name": "item", "value": float64(1)},
					map[string]any{"name": "item", "value": float64(1)},
				},
				"nested": map[string]any{
					"level2": []any{
						map[string]any{"name": "item", "value": float64(1)},
						map[string]any{"name": "item", "value": float64(1)},
					},
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "tool"},
		"",
		"",
		&buf,
	)

	// Assert
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// level1 should be collapsed (2 identical -> 1 with count)
	level1, ok := output["level1"].([]any)
	require.True(t, ok, "level1 should be an array")
	assert.Len(t, level1, 1, "level1 duplicates should collapse")

	// level2 inside nested should ALSO be collapsed independently
	nested, ok := output["nested"].(map[string]any)
	require.True(t, ok, "nested should be a map")
	level2, ok := nested["level2"].([]any)
	require.True(t, ok, "level2 should be an array")
	assert.Len(t, level2, 1, "level2 duplicates should collapse independently")

	// Both should have count=2, proving they were not compared across levels
	l1Item, ok := level1[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(2), l1Item["count"])

	l2Item, ok := level2[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(2), l2Item["count"])
}

// TestE2E_Dedup_EmptyArrays verifies that empty arrays are handled gracefully
// without error and without adding dedupStats.
func TestE2E_Dedup_EmptyArrays(t *testing.T) {
	// Arrange: Parser returns empty arrays
	runner := &mockRunner{stdout: "output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"errors":   []any{},
				"warnings": []any{},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "eslint"},
		"",
		"",
		&buf,
	)

	// Assert
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// Empty arrays should remain empty
	errors, ok := output["errors"].([]any)
	require.True(t, ok)
	assert.Empty(t, errors)

	warnings, ok := output["warnings"].([]any)
	require.True(t, ok)
	assert.Empty(t, warnings)

	// No dedupStats (nothing to dedupe)
	_, hasDedupStats := output["dedupStats"]
	assert.False(t, hasDedupStats, "no dedupStats for empty arrays")
}

// TestE2E_Dedup_PassthroughUnaffected verifies that passthrough mode (no --json flag)
// is not affected by deduplication - raw output passes through unchanged.
func TestE2E_Dedup_PassthroughUnaffected(t *testing.T) {
	// Arrange
	rawOutput := "error: something failed\nerror: something failed\nerror: something failed\n"
	runner := &mockRunner{stdout: rawOutput}
	registry := &mockRegistry{}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act: Execute WITHOUT --json flag (passthrough mode)
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"some-command"}, // No --json flag
		"",                       // envJSON empty
		"",                       // envDisableFilter
		&buf,
	)

	// Assert: Raw output should pass through unchanged
	require.NoError(t, err)
	assert.Equal(t, rawOutput, buf.String(), "passthrough mode should not modify output")
}

// TestE2E_Dedup_UniqueOutputsNatural verifies that naturally unique outputs
// (like ls entries, container IDs) are NOT deduplicated.
func TestE2E_Dedup_UniqueOutputsNatural(t *testing.T) {
	// Arrange: Parser returns unique ls-like entries
	runner := &mockRunner{stdout: "ls output"}
	parser := &mockParser{
		matchCmd: "ls",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"entries": []any{
					map[string]any{"name": "file1.txt", "size": float64(1024)},
					map[string]any{"name": "file2.txt", "size": float64(2048)},
					map[string]any{"name": "file3.txt", "size": float64(512)},
					map[string]any{"name": "dir1", "size": float64(4096)},
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "ls"},
		"",
		"",
		&buf,
	)

	// Assert: All unique entries preserved
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	entries, ok := output["entries"].([]any)
	require.True(t, ok)
	assert.Len(t, entries, 4, "all unique entries should be preserved")

	// No count fields added
	for i, entry := range entries {
		entryMap, ok := entry.(map[string]any)
		require.True(t, ok)
		_, hasCount := entryMap["count"]
		assert.False(t, hasCount, "unique entry %d should not have count field", i)
	}

	// No dedupStats (no reduction)
	_, hasDedupStats := output["dedupStats"]
	assert.False(t, hasDedupStats, "no dedupStats when all items unique")
}

// TestE2E_Dedup_RepetitiveOutputsCollapse verifies that repetitive outputs
// (like lint errors, log lines) are properly collapsed.
func TestE2E_Dedup_RepetitiveOutputsCollapse(t *testing.T) {
	// Arrange: Parser returns many repetitive lint errors
	errors := make([]any, 50)
	for i := 0; i < 50; i++ {
		errors[i] = map[string]any{
			"rule":    "no-console",
			"message": "Unexpected console statement",
			"file":    "app.js",
		}
	}

	runner := &mockRunner{stdout: "lint output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result:   domain.NewParseResult(map[string]any{"errors": errors}, "", 0),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "eslint"},
		"",
		"",
		&buf,
	)

	// Assert
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// 50 identical errors should collapse to 1
	outputErrors, ok := output["errors"].([]any)
	require.True(t, ok)
	assert.Len(t, outputErrors, 1, "50 identical errors should collapse to 1")

	// Check count is 50
	firstError, ok := outputErrors[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(50), firstError["count"])

	// Check dedupStats
	dedupStats, ok := output["dedupStats"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, float64(50), dedupStats["original_count"])
	assert.Equal(t, float64(1), dedupStats["deduped_count"])
	assert.Equal(t, "98%", dedupStats["reduction"])
}

// TestE2E_Dedup_MixedUniqueAndDuplicate verifies handling of arrays with both
// unique and duplicate items.
func TestE2E_Dedup_MixedUniqueAndDuplicate(t *testing.T) {
	// Arrange: Mix of unique and duplicate items
	runner := &mockRunner{stdout: "lint output"}
	parser := &mockParser{
		matchCmd: "eslint",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"errors": []any{
					map[string]any{"rule": "no-console", "file": "a.js"}, // unique
					map[string]any{"rule": "semi", "file": "b.js"},       // duplicate x3
					map[string]any{"rule": "semi", "file": "b.js"},
					map[string]any{"rule": "semi", "file": "b.js"},
					map[string]any{"rule": "indent", "file": "c.js"}, // unique
				},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "eslint"},
		"",
		"",
		&buf,
	)

	// Assert
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// 5 items -> 3 unique (1 + 3->1 + 1)
	errors, ok := output["errors"].([]any)
	require.True(t, ok)
	assert.Len(t, errors, 3, "5 items with 3 duplicates should become 3")

	// Find the duplicated item and check its count
	foundDuplicate := false
	for _, err := range errors {
		errMap, ok := err.(map[string]any)
		require.True(t, ok)
		if errMap["rule"] == "semi" {
			assert.Equal(t, float64(3), errMap["count"])
			foundDuplicate = true
		}
	}
	assert.True(t, foundDuplicate, "should find the collapsed semi error")
}

// TestE2E_Dedup_NonObjectArrayItems verifies handling of arrays with non-object items
// (strings, numbers).
func TestE2E_Dedup_NonObjectArrayItems(t *testing.T) {
	// Arrange: Array of strings with duplicates
	runner := &mockRunner{stdout: "output"}
	parser := &mockParser{
		matchCmd: "tool",
		matchSub: []string{},
		result: domain.NewParseResult(
			map[string]any{
				"tags": []any{"error", "error", "warning", "error", "info"},
			},
			"",
			0,
		),
	}
	registry := &mockRegistry{parsers: []mockParser{*parser}}
	deduper := application.NewDeduper()

	h := NewHandlerWithDeduplicator(runner, registry, nil, nil, deduper)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"--json", "tool"},
		"",
		"",
		&buf,
	)

	// Assert
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	// 5 items with 3 "error" -> 3 unique
	tags, ok := output["tags"].([]any)
	require.True(t, ok)
	assert.Len(t, tags, 3, "5 strings with duplicates should become 3 unique")
}
