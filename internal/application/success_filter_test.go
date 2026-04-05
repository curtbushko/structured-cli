package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// TestSuccessFilterer_ImplementsInterface verifies that SuccessFilterer implements
// the ports.SuccessFilter interface.
func TestSuccessFilterer_ImplementsInterface(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// Then: SuccessFilterer implements ports.SuccessFilter interface
	var _ ports.SuccessFilter = filterer
	require.NotNil(t, filterer)
}

// TestNewSuccessFilterer_DefaultConfig tests that NewSuccessFilterer creates a
// filterer with default configuration values.
func TestNewSuccessFilterer_DefaultConfig(t *testing.T) {
	// Given/When: Create SuccessFilterer with NewSuccessFilterer()
	filterer := NewSuccessFilterer()

	// Then: Filterer has default config with Enabled=true
	require.NotNil(t, filterer)
	assert.True(t, filterer.config.Enabled)
}

// TestNewSuccessFiltererWithConfig_CustomConfig tests that NewSuccessFiltererWithConfig
// accepts custom configuration.
func TestNewSuccessFiltererWithConfig_CustomConfig(t *testing.T) {
	// Given: Create custom config
	config := domain.SuccessFilterConfig{
		Enabled: false,
	}

	// When: Create SuccessFilterer with custom config
	filterer := NewSuccessFiltererWithConfig(config)

	// Then: Filterer has custom config values
	require.NotNil(t, filterer)
	assert.False(t, filterer.config.Enabled)
}

// TestSuccessFilterer_ShouldFilter_NpmTest tests that ShouldFilter returns true
// for 'npm test' command.
func TestSuccessFilterer_ShouldFilter_NpmTest(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Call ShouldFilter with 'npm test'
	result := filterer.ShouldFilter("npm", []string{"test"})

	// Then: Returns true
	assert.True(t, result)
}

// TestSuccessFilterer_ShouldFilter_NpmRunTest tests that ShouldFilter returns true
// for 'npm run test' command.
func TestSuccessFilterer_ShouldFilter_NpmRunTest(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Call ShouldFilter with 'npm run test'
	result := filterer.ShouldFilter("npm", []string{"run", "test"})

	// Then: Returns true
	assert.True(t, result)
}

// TestSuccessFilterer_ShouldFilter_GoTest tests that ShouldFilter returns true
// for 'go test' command.
func TestSuccessFilterer_ShouldFilter_GoTest(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Call ShouldFilter with 'go test'
	result := filterer.ShouldFilter("go", []string{"test"})

	// Then: Returns true
	assert.True(t, result)
}

// TestSuccessFilterer_ShouldFilter_Pytest tests that ShouldFilter returns true
// for 'pytest' command.
func TestSuccessFilterer_ShouldFilter_Pytest(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Call ShouldFilter with 'pytest'
	result := filterer.ShouldFilter("pytest", []string{})

	// Then: Returns true
	assert.True(t, result)
}

// TestSuccessFilterer_ShouldFilter_CargoTest tests that ShouldFilter returns true
// for 'cargo test' command.
func TestSuccessFilterer_ShouldFilter_CargoTest(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Call ShouldFilter with 'cargo test'
	result := filterer.ShouldFilter("cargo", []string{"test"})

	// Then: Returns true
	assert.True(t, result)
}

// TestSuccessFilterer_ShouldFilter_GitStatus tests that ShouldFilter returns false
// for 'git status' command.
func TestSuccessFilterer_ShouldFilter_GitStatus(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Call ShouldFilter with 'git status'
	result := filterer.ShouldFilter("git", []string{"status"})

	// Then: Returns false (not a test/lint command)
	assert.False(t, result)
}

// TestSuccessFilterer_ShouldFilter_Eslint tests that ShouldFilter returns true
// for 'eslint' command.
func TestSuccessFilterer_ShouldFilter_Eslint(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Call ShouldFilter with 'eslint'
	result := filterer.ShouldFilter("eslint", []string{})

	// Then: Returns true
	assert.True(t, result)
}

// TestSuccessFilterer_ShouldFilter_GolangciLint tests that ShouldFilter returns true
// for 'golangci-lint run' command.
func TestSuccessFilterer_ShouldFilter_GolangciLint(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Call ShouldFilter with 'golangci-lint run'
	result := filterer.ShouldFilter("golangci-lint", []string{"run"})

	// Then: Returns true
	assert.True(t, result)
}

// TestSuccessFilterer_ShouldFilter_Disabled tests that ShouldFilter returns false
// when filterer is disabled.
func TestSuccessFilterer_ShouldFilter_Disabled(t *testing.T) {
	// Given: Create disabled SuccessFilterer
	config := domain.SuccessFilterConfig{Enabled: false}
	filterer := NewSuccessFiltererWithConfig(config)

	// When: Call ShouldFilter with 'go test'
	result := filterer.ShouldFilter("go", []string{"test"})

	// Then: Returns false (disabled)
	assert.False(t, result)
}

// TestSuccessFilterer_Filter_JestOutput tests that Filter removes passed tests
// from jest output.
func TestSuccessFilterer_Filter_JestOutput(t *testing.T) {
	// Given: Create SuccessFilterer and jest-style output with passed/failed tests
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test1", "status": "passed"},
			map[string]any{"name": "test2", "status": "failed"},
			map[string]any{"name": "test3", "status": "passed"},
			map[string]any{"name": "test4", "status": "skipped"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed tests are removed, failures preserved
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Len(t, tests, 2) // Only failed and skipped
	assert.Equal(t, 4, stats.Total)
	assert.Equal(t, 2, stats.Passed)
	assert.Equal(t, 1, stats.Failed)
	assert.Equal(t, 1, stats.Skipped)
	assert.Equal(t, 2, stats.Removed)
	assert.Equal(t, 2, stats.Kept)
}

// TestSuccessFilterer_Filter_PytestOutput tests that Filter removes passed tests
// from pytest output.
func TestSuccessFilterer_Filter_PytestOutput(t *testing.T) {
	// Given: Create SuccessFilterer and pytest-style output
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test1", "outcome": "passed"},
			map[string]any{"name": "test2", "outcome": "failed"},
			map[string]any{"name": "test3", "outcome": "passed"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed tests are removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Len(t, tests, 1) // Only failed
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 2, stats.Passed)
	assert.Equal(t, 1, stats.Failed)
}

// TestSuccessFilterer_Filter_GoTestOutput tests that Filter removes passed tests
// from go test output.
func TestSuccessFilterer_Filter_GoTestOutput(t *testing.T) {
	// Given: Create SuccessFilterer and go test-style output
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "TestOne", "action": "pass"},
			map[string]any{"name": "TestTwo", "action": "fail"},
			map[string]any{"name": "TestThree", "action": "pass"},
			map[string]any{"name": "TestFour", "action": "skip"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed tests are removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Len(t, tests, 2) // fail and skip
	assert.Equal(t, 4, stats.Total)
	assert.Equal(t, 2, stats.Passed)
	assert.Equal(t, 1, stats.Failed)
	assert.Equal(t, 1, stats.Skipped)
}

// TestSuccessFilterer_Filter_VitestOutput tests that Filter removes passed tests
// from vitest output.
func TestSuccessFilterer_Filter_VitestOutput(t *testing.T) {
	// Given: Create SuccessFilterer and vitest-style output
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test1", "state": "pass"},
			map[string]any{"name": "test2", "state": "fail"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed tests are removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Len(t, tests, 1) // Only failed
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 1, stats.Passed)
	assert.Equal(t, 1, stats.Failed)
}

// TestSuccessFilterer_Filter_CargoTestOutput tests that Filter removes passed tests
// from cargo test output.
func TestSuccessFilterer_Filter_CargoTestOutput(t *testing.T) {
	// Given: Create SuccessFilterer and cargo test-style output
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test1", "status": "ok"},
			map[string]any{"name": "test2", "status": "FAILED"},
			map[string]any{"name": "test3", "status": "ok"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed tests are removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Len(t, tests, 1) // Only failed
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 2, stats.Passed)
	assert.Equal(t, 1, stats.Failed)
}

// TestSuccessFilterer_Filter_MochaOutput tests that Filter removes passed tests
// from mocha output.
func TestSuccessFilterer_Filter_MochaOutput(t *testing.T) {
	// Given: Create SuccessFilterer and mocha-style output
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"title": "test1", "state": "passed"},
			map[string]any{"title": "test2", "state": "failed"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed tests are removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Len(t, tests, 1) // Only failed
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 1, stats.Passed)
	assert.Equal(t, 1, stats.Failed)
}

// TestSuccessFilterer_Filter_PreservesFailures tests that Filter preserves all
// failures even if they have different field names.
func TestSuccessFilterer_Filter_PreservesFailures(t *testing.T) {
	// Given: Create SuccessFilterer and data with various failure statuses
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test1", "status": "failed"},
			map[string]any{"name": "test2", "status": "error"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: All failures preserved
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Len(t, tests, 2)
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 0, stats.Passed)
	assert.Equal(t, 2, stats.Failed)
}

// TestSuccessFilterer_Filter_DisabledReturnsUnchanged tests that disabled filter
// returns data unchanged.
func TestSuccessFilterer_Filter_DisabledReturnsUnchanged(t *testing.T) {
	// Given: Create disabled SuccessFilterer
	config := domain.SuccessFilterConfig{Enabled: false}
	filterer := NewSuccessFiltererWithConfig(config)
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test1", "status": "passed"},
			map[string]any{"name": "test2", "status": "failed"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Data unchanged
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Len(t, tests, 2) // All preserved
	assert.Equal(t, 0, stats.Total)
	assert.Equal(t, 0, stats.Removed)
}

// TestSuccessFilterer_Filter_NilData tests that Filter handles nil data gracefully.
func TestSuccessFilterer_Filter_NilData(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Filter nil data
	result, stats := filterer.Filter(nil)

	// Then: Returns nil with zero stats
	assert.Nil(t, result)
	assert.Equal(t, 0, stats.Total)
}

// TestSuccessFilterer_Filter_EmptyArray tests that Filter handles empty arrays.
func TestSuccessFilterer_Filter_EmptyArray(t *testing.T) {
	// Given: Create SuccessFilterer and data with empty tests array
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Empty array preserved with zero stats
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Empty(t, tests)
	assert.Equal(t, 0, stats.Total)
}

// TestSuccessFilterer_Filter_NestedResults tests that Filter handles nested results array.
func TestSuccessFilterer_Filter_NestedResults(t *testing.T) {
	// Given: Create SuccessFilterer and data with nested results
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"results": []any{
			map[string]any{"name": "suite1", "status": "passed"},
			map[string]any{"name": "suite2", "status": "failed"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed items removed from results
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	results, ok := resultMap["results"].([]any)
	require.True(t, ok, "results should be an array")
	assert.Len(t, results, 1)
	assert.Equal(t, 2, stats.Total)
}

// TestSuccessFilterer_Filter_TopLevelArray tests that Filter handles top-level array.
func TestSuccessFilterer_Filter_TopLevelArray(t *testing.T) {
	// Given: Create SuccessFilterer and top-level array data
	filterer := NewSuccessFilterer()
	data := []any{
		map[string]any{"name": "test1", "status": "passed"},
		map[string]any{"name": "test2", "status": "failed"},
		map[string]any{"name": "test3", "status": "passed"},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed items removed
	resultArr, ok := result.([]any)
	require.True(t, ok, "result should be an array")
	assert.Len(t, resultArr, 1) // Only failed
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 2, stats.Passed)
	assert.Equal(t, 1, stats.Failed)
}

// TestSuccessFilterer_Filter_NoStatusField tests that Filter handles items without
// status field.
func TestSuccessFilterer_Filter_NoStatusField(t *testing.T) {
	// Given: Create SuccessFilterer and data with no status field
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test1", "duration": 100},
			map[string]any{"name": "test2", "duration": 200},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: All items preserved (no status to filter by)
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok, "tests should be an array")
	assert.Len(t, tests, 2)
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 2, stats.Kept)
}

// TestSuccessFilterer_Filter_LintIssues tests that Filter handles eslint-style issues array.
func TestSuccessFilterer_Filter_LintIssues(t *testing.T) {
	// Given: Create SuccessFilterer and eslint-style output with issues
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"issues": []any{
			map[string]any{"message": "issue1", "severity": float64(2)}, // error
			map[string]any{"message": "issue2", "severity": float64(1)}, // warning
			map[string]any{"message": "issue3", "severity": float64(2)}, // error
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Issues array preserved (linters don't filter by default)
	resultMap, ok := result.(map[string]any)
	require.True(t, ok, "result should be a map")
	issues, ok := resultMap["issues"].([]any)
	require.True(t, ok, "issues should be an array")
	assert.Len(t, issues, 3) // All preserved
	assert.Equal(t, 3, stats.Total)
}
