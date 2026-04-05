package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// TestDeduper_ImplementsInterface verifies that Deduper implements
// the ports.Deduplicator interface.
func TestDeduper_ImplementsInterface(t *testing.T) {
	// Given: Create Deduper
	deduper := NewDeduper()

	// Then: Deduper implements ports.Deduplicator interface
	var _ ports.Deduplicator = deduper
	require.NotNil(t, deduper)
}

// TestNewDeduper tests that NewDeduper creates a deduper with default config.
func TestNewDeduper(t *testing.T) {
	// Given/When: Create Deduper with NewDeduper()
	deduper := NewDeduper()

	// Then: Deduper has default config with Enabled=true
	require.NotNil(t, deduper)
	assert.True(t, deduper.config.Enabled)
}

// TestNewDeduperWithConfig tests that NewDeduperWithConfig accepts custom config.
func TestNewDeduperWithConfig(t *testing.T) {
	// Given: Create custom config
	config := domain.DedupConfig{
		Enabled: false,
	}

	// When: Create Deduper with custom config
	deduper := NewDeduperWithConfig(config)

	// Then: Deduper has custom config values
	require.NotNil(t, deduper)
	assert.False(t, deduper.config.Enabled)
}

// TestDeduper_RawTextDedup tests that deduplication collapses identical lines with count.
func TestDeduper_RawTextDedup(t *testing.T) {
	// Given: Create Deduper and raw text with duplicate lines
	deduper := NewDeduper()
	input := "line1\nline1\nline1\nline2\nline2"

	// When: Dedupe the raw text
	result, stats := deduper.Dedupe(input)

	// Then: Identical lines are collapsed with count
	resultStr, ok := result.(string)
	require.True(t, ok, "result should be a string")
	assert.Contains(t, resultStr, "line1 (repeated 3 times)")
	assert.Contains(t, resultStr, "line2 (repeated 2 times)")
	assert.Equal(t, 5, stats.OriginalCount)
	assert.Equal(t, 2, stats.DedupedCount)
}

// TestDeduper_RawTextDedup_UniqueLines tests that unique lines are preserved.
func TestDeduper_RawTextDedup_UniqueLines(t *testing.T) {
	// Given: Create Deduper and raw text with unique lines
	deduper := NewDeduper()
	input := "line1\nline2\nline3"

	// When: Dedupe the raw text
	result, stats := deduper.Dedupe(input)

	// Then: Unique lines are preserved without count
	resultStr, ok := result.(string)
	require.True(t, ok, "result should be a string")
	assert.Contains(t, resultStr, "line1")
	assert.Contains(t, resultStr, "line2")
	assert.Contains(t, resultStr, "line3")
	assert.NotContains(t, resultStr, "repeated")
	assert.Equal(t, 3, stats.OriginalCount)
	assert.Equal(t, 3, stats.DedupedCount)
}

// TestDeduper_JSONArrayDedup tests that identical objects at same level are collapsed.
func TestDeduper_JSONArrayDedup(t *testing.T) {
	// Given: Create Deduper and array with duplicate objects
	deduper := NewDeduper()
	input := []any{
		map[string]any{"name": "alice", "age": 30},
		map[string]any{"name": "alice", "age": 30},
		map[string]any{"name": "alice", "age": 30},
		map[string]any{"name": "bob", "age": 25},
	}

	// When: Dedupe the array
	result, stats := deduper.Dedupe(input)

	// Then: Identical objects are collapsed with count
	resultArr, ok := result.([]any)
	require.True(t, ok, "result should be an array")
	assert.Len(t, resultArr, 2) // alice (grouped) + bob
	assert.Equal(t, 4, stats.OriginalCount)
	assert.Equal(t, 2, stats.DedupedCount)
}

// TestDeduper_JSONArrayDedup_DifferentObjects tests that different objects are preserved.
func TestDeduper_JSONArrayDedup_DifferentObjects(t *testing.T) {
	// Given: Create Deduper and array with all different objects
	deduper := NewDeduper()
	input := []any{
		map[string]any{"name": "alice", "age": 30},
		map[string]any{"name": "bob", "age": 25},
		map[string]any{"name": "charlie", "age": 35},
	}

	// When: Dedupe the array
	result, stats := deduper.Dedupe(input)

	// Then: All objects are preserved individually
	resultArr, ok := result.([]any)
	require.True(t, ok, "result should be an array")
	assert.Len(t, resultArr, 3)
	assert.Equal(t, 3, stats.OriginalCount)
	assert.Equal(t, 3, stats.DedupedCount)
}

// TestDeduper_JSONArrayDedup_NestedArrays tests that nested arrays are processed independently.
func TestDeduper_JSONArrayDedup_NestedArrays(t *testing.T) {
	// Given: Create Deduper and object with nested array containing duplicates
	deduper := NewDeduper()
	input := map[string]any{
		"items": []any{
			map[string]any{"id": 1},
			map[string]any{"id": 1},
			map[string]any{"id": 2},
		},
	}

	// When: Dedupe the structure
	result, stats := deduper.Dedupe(input)

	// Then: Nested array is deduplicated independently
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	items, ok := resultMap["items"].([]any)
	require.True(t, ok, "items should be an array")
	assert.Len(t, items, 2) // id:1 (grouped) + id:2
	assert.Equal(t, 3, stats.OriginalCount)
	assert.Equal(t, 2, stats.DedupedCount)
}

// TestDeduper_JSONArrayDedup_SameLevel tests that only same-level objects are compared.
func TestDeduper_JSONArrayDedup_SameLevel(t *testing.T) {
	// Given: Create Deduper and structure with same object at different nesting levels
	deduper := NewDeduper()
	input := []any{
		map[string]any{"id": 1},
		map[string]any{
			"nested": []any{
				map[string]any{"id": 1}, // Same as parent level, but shouldn't be deduplicated
			},
		},
	}

	// When: Dedupe the structure
	result, _ := deduper.Dedupe(input)

	// Then: Objects at different levels are NOT deduplicated against each other
	resultArr, ok := result.([]any)
	require.True(t, ok, "result should be an array")
	assert.Len(t, resultArr, 2) // Both items preserved at this level
}

// TestDeduper_DeepEquality tests that objects are compared by all fields.
func TestDeduper_DeepEquality(t *testing.T) {
	// Given: Create Deduper and objects that differ in one field
	deduper := NewDeduper()
	input := []any{
		map[string]any{"name": "alice", "age": 30},
		map[string]any{"name": "alice", "age": 31}, // Different age
	}

	// When: Dedupe the array
	result, stats := deduper.Dedupe(input)

	// Then: Objects are NOT collapsed (different by one field)
	resultArr, ok := result.([]any)
	require.True(t, ok, "result should be an array")
	assert.Len(t, resultArr, 2)
	assert.Equal(t, 2, stats.OriginalCount)
	assert.Equal(t, 2, stats.DedupedCount)
}

// TestDeduper_PreservesFirstOccurrence tests that first item becomes sample.
func TestDeduper_PreservesFirstOccurrence(t *testing.T) {
	// Given: Create Deduper and array with duplicate objects
	deduper := NewDeduper()
	input := []any{
		map[string]any{"name": "alice", "order": 1},
		map[string]any{"name": "alice", "order": 1},
	}

	// When: Dedupe the array
	result, _ := deduper.Dedupe(input)

	// Then: First occurrence is kept as sample
	resultArr, ok := result.([]any)
	require.True(t, ok, "result should be an array")
	require.Len(t, resultArr, 1)

	// The grouped item should have count field
	item, ok := resultArr[0].(map[string]any)
	require.True(t, ok, "item should be a map")
	assert.Equal(t, 2, item["count"])
}

// TestDeduper_AddsCountField tests that count is added to grouped items.
func TestDeduper_AddsCountField(t *testing.T) {
	// Given: Create Deduper and array with 5 identical objects
	deduper := NewDeduper()
	input := []any{
		map[string]any{"id": 1},
		map[string]any{"id": 1},
		map[string]any{"id": 1},
		map[string]any{"id": 1},
		map[string]any{"id": 1},
	}

	// When: Dedupe the array
	result, _ := deduper.Dedupe(input)

	// Then: Grouped item has count=5
	resultArr, ok := result.([]any)
	require.True(t, ok, "result should be an array")
	require.Len(t, resultArr, 1)

	item, ok := resultArr[0].(map[string]any)
	require.True(t, ok, "item should be a map")
	assert.Equal(t, 5, item["count"])
}

// TestDeduper_EmptyArray tests that empty arrays are handled gracefully.
func TestDeduper_EmptyArray(t *testing.T) {
	// Given: Create Deduper and empty array
	deduper := NewDeduper()
	input := []any{}

	// When: Dedupe the empty array
	result, stats := deduper.Dedupe(input)

	// Then: Empty array is returned with zero stats
	resultArr, ok := result.([]any)
	require.True(t, ok, "result should be an array")
	assert.Empty(t, resultArr)
	assert.Equal(t, 0, stats.OriginalCount)
	assert.Equal(t, 0, stats.DedupedCount)
}

// TestDeduper_NilData tests that nil data is handled gracefully.
func TestDeduper_NilData(t *testing.T) {
	// Given: Create Deduper
	deduper := NewDeduper()

	// When: Dedupe nil
	result, stats := deduper.Dedupe(nil)

	// Then: Nil is returned with zero stats
	assert.Nil(t, result)
	assert.Equal(t, 0, stats.OriginalCount)
	assert.Equal(t, 0, stats.DedupedCount)
}

// TestDeduper_NonArrayData tests that non-array data passes through unchanged.
func TestDeduper_NonArrayData(t *testing.T) {
	// Given: Create Deduper and non-array data
	deduper := NewDeduper()
	input := map[string]any{
		"name":  "test",
		"count": 42,
	}

	// When: Dedupe non-array data
	result, stats := deduper.Dedupe(input)

	// Then: Data passes through unchanged with zero stats
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	assert.Equal(t, "test", resultMap["name"])
	assert.Equal(t, 42, resultMap["count"])
	assert.Equal(t, 0, stats.OriginalCount)
	assert.Equal(t, 0, stats.DedupedCount)
}

// TestDeduper_Disabled tests that disabled deduper returns input unchanged.
func TestDeduper_Disabled(t *testing.T) {
	// Given: Create Deduper with Enabled=false
	config := domain.DedupConfig{
		Enabled: false,
	}
	deduper := NewDeduperWithConfig(config)

	// When: Dedupe array with duplicates
	input := []any{
		map[string]any{"id": 1},
		map[string]any{"id": 1},
		map[string]any{"id": 1},
	}
	result, stats := deduper.Dedupe(input)

	// Then: Input is returned unchanged with zero stats
	resultArr, ok := result.([]any)
	require.True(t, ok, "result should be an array")
	assert.Len(t, resultArr, 3) // All items preserved
	assert.Equal(t, 0, stats.OriginalCount)
	assert.Equal(t, 0, stats.DedupedCount)
}

// TestDeduper_Stats tests that DedupResult stats are correct.
func TestDeduper_Stats(t *testing.T) {
	tests := []struct {
		name           string
		input          any
		expectedOrig   int
		expectedDedup  int
		expectedReduce string
	}{
		{
			name: "50% reduction",
			input: []any{
				map[string]any{"id": 1},
				map[string]any{"id": 1},
				map[string]any{"id": 2},
				map[string]any{"id": 2},
			},
			expectedOrig:   4,
			expectedDedup:  2,
			expectedReduce: "50%",
		},
		{
			name: "75% reduction",
			input: []any{
				map[string]any{"id": 1},
				map[string]any{"id": 1},
				map[string]any{"id": 1},
				map[string]any{"id": 1},
			},
			expectedOrig:   4,
			expectedDedup:  1,
			expectedReduce: "75%",
		},
		{
			name: "0% reduction (all unique)",
			input: []any{
				map[string]any{"id": 1},
				map[string]any{"id": 2},
				map[string]any{"id": 3},
			},
			expectedOrig:   3,
			expectedDedup:  3,
			expectedReduce: "0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deduper := NewDeduper()

			_, stats := deduper.Dedupe(tt.input)

			assert.Equal(t, tt.expectedOrig, stats.OriginalCount)
			assert.Equal(t, tt.expectedDedup, stats.DedupedCount)
			assert.Equal(t, tt.expectedReduce, stats.Reduction)
		})
	}
}
