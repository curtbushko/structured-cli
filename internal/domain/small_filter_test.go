package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinTokenThreshold_Value(t *testing.T) {
	// Given: Access MIN_TOKEN_THRESHOLD constant
	// When: Check value
	// Then: Value equals 25
	assert.Equal(t, 25, MinTokenThreshold)
}

func TestSmallOutputConfig_Defaults(t *testing.T) {
	// Given: Create new SmallOutputConfig with NewSmallOutputConfig()
	config := NewSmallOutputConfig()

	// When: Access TokenThreshold and Enabled fields
	// Then: TokenThreshold equals MIN_TOKEN_THRESHOLD (25), Enabled is true
	assert.True(t, config.Enabled)
	assert.Equal(t, MinTokenThreshold, config.TokenThreshold)
}

func TestSmallOutputConfig_Fields(t *testing.T) {
	// Test that SmallOutputConfig has the required fields
	config := SmallOutputConfig{
		Enabled:        false,
		TokenThreshold: 50,
	}

	assert.False(t, config.Enabled)
	assert.Equal(t, 50, config.TokenThreshold)
}

func TestSmallOutputResult_Minimal(t *testing.T) {
	// Given: Create SmallOutputResult with Status='clean' and Summary='nothing to commit'
	result := NewSmallOutputResult("clean", "nothing to commit")

	// When: Serialize to JSON
	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Then: JSON has 'status' and 'summary' fields with correct values
	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "clean", parsed["status"])
	assert.Equal(t, "nothing to commit", parsed["summary"])
}

func TestSmallOutputResult_Fields(t *testing.T) {
	// Test that SmallOutputResult has the required fields
	result := SmallOutputResult{
		Status:  "modified",
		Summary: "3 files changed",
	}

	assert.Equal(t, "modified", result.Status)
	assert.Equal(t, "3 files changed", result.Summary)
}

func TestNewSmallOutputResult(t *testing.T) {
	// Given: Create SmallOutputResult using constructor
	status := "clean"
	summary := "nothing to commit"

	// When: Call NewSmallOutputResult
	result := NewSmallOutputResult(status, summary)

	// Then: Result has correct Status and Summary
	assert.Equal(t, status, result.Status)
	assert.Equal(t, summary, result.Summary)
}

func TestSmallOutputResult_JSONTags(t *testing.T) {
	// Verify JSON serialization uses correct field names
	result := NewSmallOutputResult("test-status", "test-summary")

	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Should serialize to {"status":"test-status","summary":"test-summary"}
	expected := `{"status":"test-status","summary":"test-summary"}`
	assert.JSONEq(t, expected, string(data))
}
