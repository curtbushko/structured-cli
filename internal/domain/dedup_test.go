package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDedupConfig_Defaults(t *testing.T) {
	// Given: Create new DedupConfig with NewDedupConfig()
	config := NewDedupConfig()

	// When: Access Enabled field
	// Then: Enabled is true
	assert.True(t, config.Enabled)
}

func TestDedupConfig_Fields(t *testing.T) {
	// Test that DedupConfig has the required fields
	config := DedupConfig{
		Enabled: false,
	}

	assert.False(t, config.Enabled)
}

func TestDedupResult_Fields(t *testing.T) {
	// Test that DedupResult has the required fields
	result := DedupResult{
		OriginalCount: 100,
		DedupedCount:  50,
		Reduction:     "50%",
	}

	assert.Equal(t, 100, result.OriginalCount)
	assert.Equal(t, 50, result.DedupedCount)
	assert.Equal(t, "50%", result.Reduction)
}

func TestDedupResult_JSONTags(t *testing.T) {
	// Verify JSON serialization uses snake_case field names
	result := DedupResult{
		OriginalCount: 100,
		DedupedCount:  50,
		Reduction:     "50%",
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Should serialize to {"original_count":100,"deduped_count":50,"reduction":"50%"}
	expected := `{"original_count":100,"deduped_count":50,"reduction":"50%"}`
	assert.JSONEq(t, expected, string(data))
}

func TestNewDedupResult(t *testing.T) {
	// Given: Create DedupResult using constructor
	original := 100
	deduped := 50

	// When: Call NewDedupResult
	result := NewDedupResult(original, deduped)

	// Then: Result has correct fields with calculated reduction
	assert.Equal(t, original, result.OriginalCount)
	assert.Equal(t, deduped, result.DedupedCount)
	assert.Equal(t, "50%", result.Reduction)
}

func TestNewDedupResult_ZeroOriginal(t *testing.T) {
	// Given: Create DedupResult with zero original count
	// When: Call NewDedupResult
	result := NewDedupResult(0, 0)

	// Then: Reduction is 0% (avoid division by zero)
	assert.Equal(t, "0%", result.Reduction)
}

func TestDedupItem_Fields(t *testing.T) {
	// Test that DedupItem has the required fields
	item := DedupItem{
		Count:  5,
		Sample: "test sample",
	}

	assert.Equal(t, 5, item.Count)
	assert.Equal(t, "test sample", item.Sample)
}

func TestDedupItem_JSONTags(t *testing.T) {
	// Verify JSON serialization uses correct field names
	item := DedupItem{
		Count:  5,
		Sample: "test sample",
	}

	data, err := json.Marshal(item)
	require.NoError(t, err)

	// Should serialize to {"count":5,"sample":"test sample"}
	expected := `{"count":5,"sample":"test sample"}`
	assert.JSONEq(t, expected, string(data))
}

func TestNewDedupItem(t *testing.T) {
	// Given: Create DedupItem using constructor
	count := 5
	sample := "test sample"

	// When: Call NewDedupItem
	item := NewDedupItem(count, sample)

	// Then: Item has correct fields
	assert.Equal(t, count, item.Count)
	assert.Equal(t, sample, item.Sample)
}
