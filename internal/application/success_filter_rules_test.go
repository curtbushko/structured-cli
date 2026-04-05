package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test Runner Rules Tests

// TestGetRule_JestReturnsCorrectRule tests that getRule returns the correct
// rule for jest command with status field and pass/fail values.
func TestGetRule_JestReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for npx jest
	rule, found := filterer.getRule("npx", []string{"jest"})

	// Then: Returns jest rule
	require.True(t, found)
	assert.Equal(t, "status", rule.StatusField)
	assert.Contains(t, rule.PassValues, "passed")
	assert.Contains(t, rule.FailValues, "failed")
}

// TestGetRule_PytestReturnsCorrectRule tests that getRule returns the correct
// rule for pytest command with outcome field.
func TestGetRule_PytestReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for pytest
	rule, found := filterer.getRule("pytest", []string{})

	// Then: Returns pytest rule
	require.True(t, found)
	assert.Equal(t, "outcome", rule.StatusField)
	assert.Contains(t, rule.PassValues, "passed")
	assert.Contains(t, rule.FailValues, "failed")
}

// TestGetRule_VitestReturnsCorrectRule tests that getRule returns the correct
// rule for vitest command with status field using pass/fail values.
func TestGetRule_VitestReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for npx vitest
	rule, found := filterer.getRule("npx", []string{"vitest"})

	// Then: Returns vitest rule
	require.True(t, found)
	assert.Equal(t, "status", rule.StatusField)
	assert.Contains(t, rule.PassValues, "pass")
	assert.Contains(t, rule.FailValues, "fail")
}

// TestGetRule_GoTestReturnsCorrectRule tests that getRule returns the correct
// rule for go test command with action field.
func TestGetRule_GoTestReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for go test
	rule, found := filterer.getRule("go", []string{"test"})

	// Then: Returns go test rule
	require.True(t, found)
	assert.Equal(t, "action", rule.StatusField)
	assert.Contains(t, rule.PassValues, "pass")
	assert.Contains(t, rule.FailValues, "fail")
}

// TestGetRule_CargoTestReturnsCorrectRule tests that getRule returns the correct
// rule for cargo test command with status field using ok/FAILED values.
func TestGetRule_CargoTestReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for cargo test
	rule, found := filterer.getRule("cargo", []string{"test"})

	// Then: Returns cargo test rule
	require.True(t, found)
	assert.Equal(t, "status", rule.StatusField)
	assert.Contains(t, rule.PassValues, "ok")
	assert.Contains(t, rule.FailValues, "FAILED")
}

// TestGetRule_MochaReturnsCorrectRule tests that getRule returns the correct
// rule for mocha command with state field using passed/failed values.
func TestGetRule_MochaReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for npx mocha
	rule, found := filterer.getRule("npx", []string{"mocha"})

	// Then: Returns mocha rule
	require.True(t, found)
	assert.Equal(t, "state", rule.StatusField)
	assert.Contains(t, rule.PassValues, "passed")
	assert.Contains(t, rule.FailValues, "failed")
}

// Linter Rules Tests

// TestGetRule_EslintReturnsCorrectRule tests that getRule returns the correct
// rule for eslint command with severity field filtering.
func TestGetRule_EslintReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for eslint
	rule, found := filterer.getRule("eslint", []string{})

	// Then: Returns eslint rule
	require.True(t, found)
	assert.Equal(t, "severity", rule.StatusField)
	assert.Contains(t, rule.PassValues, "warning")
	assert.Contains(t, rule.FailValues, "error")
}

// TestGetRule_GolangciLintReturnsCorrectRule tests that getRule returns the correct
// rule for golangci-lint run command.
func TestGetRule_GolangciLintReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for golangci-lint run
	rule, found := filterer.getRule("golangci-lint", []string{"run"})

	// Then: Returns golangci-lint rule (keeps all issues)
	require.True(t, found)
	// golangci-lint keeps all issues by default
	assert.Empty(t, rule.PassValues)
}

// TestGetRule_TscReturnsCorrectRule tests that getRule returns the correct
// rule for tsc command (keeps all errors).
func TestGetRule_TscReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for tsc
	rule, found := filterer.getRule("tsc", []string{})

	// Then: Returns tsc rule (keeps all)
	require.True(t, found)
	assert.Empty(t, rule.PassValues)
}

// TestGetRule_RuffReturnsCorrectRule tests that getRule returns the correct
// rule for ruff command (keeps all violations).
func TestGetRule_RuffReturnsCorrectRule(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for ruff check
	rule, found := filterer.getRule("ruff", []string{"check"})

	// Then: Returns ruff rule (keeps all)
	require.True(t, found)
	assert.Empty(t, rule.PassValues)
}

// Command Mapping Tests

// TestGetRule_NpmTestMapsToJest tests that npm test maps to jest rule.
func TestGetRule_NpmTestMapsToJest(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for npm test
	rule, found := filterer.getRule("npm", []string{"test"})

	// Then: Returns jest-style rule
	require.True(t, found)
	assert.Equal(t, "status", rule.StatusField)
	assert.Contains(t, rule.PassValues, "passed")
}

// TestGetRule_UnknownCommandReturnsFalse tests that unknown command returns false.
func TestGetRule_UnknownCommandReturnsFalse(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for unknown command
	_, found := filterer.getRule("unknown", []string{"command"})

	// Then: Returns false
	assert.False(t, found)
}

// TestGetRule_GitStatusReturnsFalse tests that non-test/lint commands return false.
func TestGetRule_GitStatusReturnsFalse(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for git status
	_, found := filterer.getRule("git", []string{"status"})

	// Then: Returns false
	assert.False(t, found)
}

// Rule Behavior Tests

// TestFilterRule_JestFiltersPassedTests tests that jest rule correctly filters
// passed tests from testResults array.
func TestFilterRule_JestFiltersPassedTests(t *testing.T) {
	// Given: Create SuccessFilterer and jest-style data
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"suites": []any{
			map[string]any{
				"name": "suite1",
				"tests": []any{
					map[string]any{"name": "test1", "status": "passed"},
					map[string]any{"name": "test2", "status": "failed"},
					map[string]any{"name": "test3", "status": "passed"},
				},
			},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed tests are removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	suites, ok := resultMap["suites"].([]any)
	require.True(t, ok)
	suite, ok := suites[0].(map[string]any)
	require.True(t, ok)
	tests, ok := suite["tests"].([]any)
	require.True(t, ok)
	assert.Len(t, tests, 1) // Only failed
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 2, stats.Passed)
	assert.Equal(t, 1, stats.Failed)
}

// TestFilterRule_PytestFiltersPassedOutcome tests that pytest rule correctly filters
// tests with outcome=passed.
func TestFilterRule_PytestFiltersPassedOutcome(t *testing.T) {
	// Given: Create SuccessFilterer and pytest-style data
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test_one", "outcome": "passed"},
			map[string]any{"name": "test_two", "outcome": "failed"},
			map[string]any{"name": "test_three", "outcome": "error"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Passed tests are removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok)
	assert.Len(t, tests, 2) // failed + error
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 1, stats.Passed)
	assert.Equal(t, 2, stats.Failed)
}

// TestFilterRule_VitestFiltersPassState tests that vitest rule correctly filters
// tests with status=pass.
func TestFilterRule_VitestFiltersPassState(t *testing.T) {
	// Given: Create SuccessFilterer and vitest-style data
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test1", "status": "pass"},
			map[string]any{"name": "test2", "status": "fail"},
			map[string]any{"name": "test3", "status": "skip"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Pass status is removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok)
	assert.Len(t, tests, 2) // fail + skip
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 1, stats.Passed)
}

// TestFilterRule_GoTestFiltersPassAction tests that go test rule correctly filters
// tests with action=pass.
func TestFilterRule_GoTestFiltersPassAction(t *testing.T) {
	// Given: Create SuccessFilterer and go test-style data
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "TestOne", "action": "pass"},
			map[string]any{"name": "TestTwo", "action": "fail"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Pass action is removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok)
	assert.Len(t, tests, 1) // Only fail
	assert.Equal(t, 2, stats.Total)
	assert.Equal(t, 1, stats.Passed)
}

// TestFilterRule_CargoTestFiltersOkStatus tests that cargo test rule correctly filters
// tests with status=ok.
func TestFilterRule_CargoTestFiltersOkStatus(t *testing.T) {
	// Given: Create SuccessFilterer and cargo test-style data
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"name": "test1", "status": "ok"},
			map[string]any{"name": "test2", "status": "FAILED"},
			map[string]any{"name": "test3", "status": "ignored"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: ok status is removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok)
	assert.Len(t, tests, 2) // FAILED + ignored
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 1, stats.Passed)
}

// TestFilterRule_MochaFiltersPassedState tests that mocha rule correctly filters
// tests with state=passed.
func TestFilterRule_MochaFiltersPassedState(t *testing.T) {
	// Given: Create SuccessFilterer and mocha-style data
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"tests": []any{
			map[string]any{"title": "test1", "state": "passed"},
			map[string]any{"title": "test2", "state": "failed"},
			map[string]any{"title": "test3", "state": "pending"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: passed state is removed
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	tests, ok := resultMap["tests"].([]any)
	require.True(t, ok)
	assert.Len(t, tests, 2) // failed + pending
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 1, stats.Passed)
}

// TestFilterRule_EslintKeepsErrors tests that eslint rule filters out warnings
// and keeps only errors (severity=error).
func TestFilterRule_EslintKeepsErrors(t *testing.T) {
	// Given: Create SuccessFilterer and eslint-style data
	filterer := NewSuccessFilterer()
	data := map[string]any{
		"issues": []any{
			map[string]any{"message": "error1", "severity": "error"},
			map[string]any{"message": "warning1", "severity": "warning"},
			map[string]any{"message": "error2", "severity": "error"},
		},
	}

	// When: Filter the data
	result, stats := filterer.Filter(data)

	// Then: Warnings are filtered, errors kept
	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	issues, ok := resultMap["issues"].([]any)
	require.True(t, ok)
	assert.Len(t, issues, 2) // Only errors
	assert.Equal(t, 3, stats.Total)
	assert.Equal(t, 1, stats.Passed) // warning is "passed"
	assert.Equal(t, 2, stats.Failed) // errors are "failed"
}

// TestFilterRule_GolangciLintRuleDefinition tests that golangci-lint rule is defined
// to keep all issues (empty pass values means nothing is filtered).
func TestFilterRule_GolangciLintRuleDefinition(t *testing.T) {
	// Given: Create SuccessFilterer
	filterer := NewSuccessFilterer()

	// When: Get rule for golangci-lint run
	rule, found := filterer.getRule("golangci-lint", []string{"run"})

	// Then: Rule has empty pass values, meaning keep all
	require.True(t, found)
	assert.Empty(t, rule.PassValues)
	assert.Empty(t, rule.FailValues)
}
