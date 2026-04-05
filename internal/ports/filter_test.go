package ports

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// TestMinimalPattern_Fields tests that MinimalPattern has Command, Subcommand,
// and Pattern fields with correct types.
func TestMinimalPattern_Fields(t *testing.T) {
	// Given: Create MinimalPattern with Command, Subcommand, Pattern regex
	pattern := regexp.MustCompile(`nothing to commit`)
	mp := MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    pattern,
	}

	// When: Access fields
	// Then: All fields are accessible and have correct types
	assert.Equal(t, "git", mp.Command)
	assert.Equal(t, "status", mp.Subcommand)
	assert.NotNil(t, mp.Pattern)
	assert.Equal(t, pattern, mp.Pattern)
}

// TestMinimalPattern_MatchesCleanGitStatus tests that MinimalPattern.Matches()
// returns true for clean git status output.
func TestMinimalPattern_MatchesCleanGitStatus(t *testing.T) {
	// Given: Create MinimalPattern for git status with 'nothing to commit' pattern
	mp := MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	}

	// When: Test pattern against clean git status output
	cleanOutput := `On branch main
Your branch is up to date with 'origin/main'.

nothing to commit, working tree clean`

	// Then: Pattern matches and returns true
	assert.True(t, mp.Matches(cleanOutput))
}

// TestMinimalPattern_MatchesReturnsFalse tests that MinimalPattern.Matches()
// returns false when pattern does not match.
func TestMinimalPattern_MatchesReturnsFalse(t *testing.T) {
	// Given: Create MinimalPattern for git status with 'nothing to commit' pattern
	mp := MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	}

	// When: Test pattern against modified git status output
	modifiedOutput := `On branch main
Changes not staged for commit:
  modified:   README.md`

	// Then: Pattern does not match and returns false
	assert.False(t, mp.Matches(modifiedOutput))
}

// TestMinimalPattern_MatchesEmptyOutput tests that MinimalPattern.Matches()
// returns false for empty output.
func TestMinimalPattern_MatchesEmptyOutput(t *testing.T) {
	// Given: Create MinimalPattern with a pattern
	mp := MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	}

	// When: Test pattern against empty output
	// Then: Pattern does not match and returns false
	assert.False(t, mp.Matches(""))
}

// mockSmallOutputFilter is a mock implementation of SmallOutputFilter for testing.
type mockSmallOutputFilter struct {
	shouldFilterResult bool
	filterResult       domain.SmallOutputResult
}

func (m *mockSmallOutputFilter) ShouldFilter(raw string, tokenCount int, cmd string, subcmds []string) bool {
	return m.shouldFilterResult
}

func (m *mockSmallOutputFilter) Filter(raw string) domain.SmallOutputResult {
	return m.filterResult
}

// TestSmallOutputFilter_Interface tests that SmallOutputFilter interface
// can be implemented and its methods are callable with correct signatures.
func TestSmallOutputFilter_Interface(t *testing.T) {
	// Given: Create mock implementation of SmallOutputFilter
	mock := &mockSmallOutputFilter{
		shouldFilterResult: true,
		filterResult:       domain.NewSmallOutputResult("clean", "nothing to commit"),
	}

	// Verify it implements the interface
	var filter SmallOutputFilter = mock
	require.NotNil(t, filter)

	// When: Call ShouldFilter method
	shouldFilter := filter.ShouldFilter("test output", 10, "git", []string{"status"})

	// Then: Method is callable and returns expected value
	assert.True(t, shouldFilter)

	// When: Call Filter method
	result := filter.Filter("nothing to commit, working tree clean")

	// Then: Method is callable and returns expected value
	assert.Equal(t, "clean", result.Status)
	assert.Equal(t, "nothing to commit", result.Summary)
}

// TestSmallOutputFilter_ShouldFilterSignature tests that ShouldFilter has
// the correct signature with raw string, token count, cmd, and subcmds.
func TestSmallOutputFilter_ShouldFilterSignature(t *testing.T) {
	// Given: Create mock implementation
	mock := &mockSmallOutputFilter{
		shouldFilterResult: false,
	}

	var filter SmallOutputFilter = mock

	// When: Call with different parameters
	testCases := []struct {
		raw        string
		tokenCount int
		cmd        string
		subcmds    []string
	}{
		{"", 0, "", nil},
		{"output", 5, "git", []string{"status"}},
		{"long output here", 100, "npm", []string{"install", "--save"}},
	}

	// Then: All calls work without panic
	for _, tc := range testCases {
		result := filter.ShouldFilter(tc.raw, tc.tokenCount, tc.cmd, tc.subcmds)
		assert.False(t, result)
	}
}

// TestSmallOutputFilter_FilterReturnsSmallOutputResult tests that Filter
// returns domain.SmallOutputResult.
func TestSmallOutputFilter_FilterReturnsSmallOutputResult(t *testing.T) {
	// Given: Create mock implementation with specific result
	expectedResult := domain.NewSmallOutputResult("modified", "3 files changed")
	mock := &mockSmallOutputFilter{
		filterResult: expectedResult,
	}

	var filter SmallOutputFilter = mock

	// When: Call Filter
	result := filter.Filter("some output")

	// Then: Result is SmallOutputResult with expected values
	assert.Equal(t, "modified", result.Status)
	assert.Equal(t, "3 files changed", result.Summary)
}
