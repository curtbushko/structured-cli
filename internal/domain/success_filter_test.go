package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessFilterConfig_Defaults(t *testing.T) {
	// Given: Create new SuccessFilterConfig with NewSuccessFilterConfig()
	config := NewSuccessFilterConfig()

	// When: Access Enabled field
	// Then: Enabled is true by default
	assert.True(t, config.Enabled)
}

func TestSuccessFilterConfig_Fields(t *testing.T) {
	// Test that SuccessFilterConfig has the required fields
	config := SuccessFilterConfig{
		Enabled: false,
	}

	assert.False(t, config.Enabled)
}

func TestSuccessFilterResult_Fields(t *testing.T) {
	// Test that SuccessFilterResult has the required fields
	result := SuccessFilterResult{
		Total:   100,
		Passed:  80,
		Failed:  15,
		Skipped: 5,
		Removed: 80,
		Kept:    20,
	}

	assert.Equal(t, 100, result.Total)
	assert.Equal(t, 80, result.Passed)
	assert.Equal(t, 15, result.Failed)
	assert.Equal(t, 5, result.Skipped)
	assert.Equal(t, 80, result.Removed)
	assert.Equal(t, 20, result.Kept)
}

func TestSuccessFilterResult_JSONTags(t *testing.T) {
	// Verify JSON serialization uses snake_case field names
	result := SuccessFilterResult{
		Total:   100,
		Passed:  80,
		Failed:  15,
		Skipped: 5,
		Removed: 80,
		Kept:    20,
	}

	data, err := json.Marshal(result)
	require.NoError(t, err)

	// Should serialize to snake_case
	expected := `{"total":100,"passed":80,"failed":15,"skipped":5,"removed":80,"kept":20}`
	assert.JSONEq(t, expected, string(data))
}

func TestNewSuccessFilterResult(t *testing.T) {
	// Given: Create SuccessFilterResult using constructor
	// When: Call NewSuccessFilterResult with all values
	result := NewSuccessFilterResult(100, 80, 15, 5, 80, 20)

	// Then: Result has correct fields
	assert.Equal(t, 100, result.Total)
	assert.Equal(t, 80, result.Passed)
	assert.Equal(t, 15, result.Failed)
	assert.Equal(t, 5, result.Skipped)
	assert.Equal(t, 80, result.Removed)
	assert.Equal(t, 20, result.Kept)
}

func TestNewSuccessFilterResult_ZeroValues(t *testing.T) {
	// Given: Create SuccessFilterResult with zero values
	// When: Call NewSuccessFilterResult
	result := NewSuccessFilterResult(0, 0, 0, 0, 0, 0)

	// Then: All values are zero
	assert.Equal(t, 0, result.Total)
	assert.Equal(t, 0, result.Passed)
	assert.Equal(t, 0, result.Failed)
	assert.Equal(t, 0, result.Skipped)
	assert.Equal(t, 0, result.Removed)
	assert.Equal(t, 0, result.Kept)
}

func TestFilterRule_Fields(t *testing.T) {
	// Test that FilterRule has the required fields
	rule := FilterRule{
		StatusField: "status",
		PassValues:  []string{"passed", "pass", "ok"},
		FailValues:  []string{"failed", "fail", "error"},
	}

	assert.Equal(t, "status", rule.StatusField)
	assert.Equal(t, []string{"passed", "pass", "ok"}, rule.PassValues)
	assert.Equal(t, []string{"failed", "fail", "error"}, rule.FailValues)
}

func TestFilterRule_Matches_Pass(t *testing.T) {
	// Given: A FilterRule with pass values
	rule := FilterRule{
		StatusField: "status",
		PassValues:  []string{"passed", "pass", "ok"},
		FailValues:  []string{"failed", "fail", "error"},
	}

	// When: Check if "passed" matches pass values
	// Then: Returns true for pass, false for fail
	assert.True(t, rule.MatchesPass("passed"))
	assert.True(t, rule.MatchesPass("pass"))
	assert.True(t, rule.MatchesPass("ok"))
	assert.False(t, rule.MatchesPass("failed"))
	assert.False(t, rule.MatchesPass("unknown"))
}

func TestFilterRule_Matches_Fail(t *testing.T) {
	// Given: A FilterRule with fail values
	rule := FilterRule{
		StatusField: "status",
		PassValues:  []string{"passed", "pass", "ok"},
		FailValues:  []string{"failed", "fail", "error"},
	}

	// When: Check if "failed" matches fail values
	// Then: Returns true for fail, false for pass
	assert.True(t, rule.MatchesFail("failed"))
	assert.True(t, rule.MatchesFail("fail"))
	assert.True(t, rule.MatchesFail("error"))
	assert.False(t, rule.MatchesFail("passed"))
	assert.False(t, rule.MatchesFail("unknown"))
}

func TestFilterRule_EmptyValues(t *testing.T) {
	// Given: A FilterRule with empty pass/fail values
	rule := FilterRule{
		StatusField: "status",
		PassValues:  []string{},
		FailValues:  []string{},
	}

	// When: Check matches
	// Then: Nothing matches
	assert.False(t, rule.MatchesPass("anything"))
	assert.False(t, rule.MatchesFail("anything"))
}

func TestNewFilterRule(t *testing.T) {
	// Given: Create FilterRule using constructor
	statusField := "outcome"
	passValues := []string{"success"}
	failValues := []string{"failure"}

	// When: Call NewFilterRule
	rule := NewFilterRule(statusField, passValues, failValues)

	// Then: Rule has correct fields
	assert.Equal(t, statusField, rule.StatusField)
	assert.Equal(t, passValues, rule.PassValues)
	assert.Equal(t, failValues, rule.FailValues)
}
