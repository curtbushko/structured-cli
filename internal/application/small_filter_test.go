package application

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Test constants for common string values.
const testGitCleanOutput = "nothing to commit, working tree clean"

// TestSmallFilter_ImplementsInterface verifies that SmallFilter implements
// the ports.SmallOutputFilter interface.
func TestSmallFilter_ImplementsInterface(t *testing.T) {
	// Given: Create SmallFilter
	filter := NewSmallFilter()

	// Then: SmallFilter implements ports.SmallOutputFilter interface
	var _ ports.SmallOutputFilter = filter
	require.NotNil(t, filter)
}

// TestNewSmallFilter_DefaultConfig tests that NewSmallFilter creates a filter
// with default configuration values.
func TestNewSmallFilter_DefaultConfig(t *testing.T) {
	// Given: Create SmallFilter with NewSmallFilter()
	filter := NewSmallFilter()

	// Then: Filter has default config with Enabled=true and TokenThreshold=25
	require.NotNil(t, filter)
	assert.True(t, filter.config.Enabled)
	assert.Equal(t, domain.MinTokenThreshold, filter.config.TokenThreshold)
}

// TestNewSmallFilterWithConfig_CustomConfig tests that NewSmallFilterWithConfig
// accepts custom configuration.
func TestNewSmallFilterWithConfig_CustomConfig(t *testing.T) {
	// Given: Create custom config
	config := domain.SmallOutputConfig{
		Enabled:        false,
		TokenThreshold: 50,
	}

	// When: Create SmallFilter with custom config
	filter := NewSmallFilterWithConfig(config)

	// Then: Filter has custom config values
	require.NotNil(t, filter)
	assert.False(t, filter.config.Enabled)
	assert.Equal(t, 50, filter.config.TokenThreshold)
}

// TestSmallFilter_RegisterPattern tests that RegisterPattern adds patterns
// to the filter.
func TestSmallFilter_RegisterPattern(t *testing.T) {
	// Given: Create SmallFilter
	filter := NewSmallFilter()

	// When: Register a pattern
	pattern := ports.MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	}
	filter.RegisterPattern(pattern)

	// Then: Filter has the registered pattern
	assert.Len(t, filter.patterns, 1)
	assert.Equal(t, "git", filter.patterns[0].Command)
	assert.Equal(t, "status", filter.patterns[0].Subcommand)
}

// TestSmallFilter_TriggersForSmallOutput tests that ShouldFilter returns true
// for small output that matches a pattern.
func TestSmallFilter_TriggersForSmallOutput(t *testing.T) {
	// Given: Create SmallFilter with default config and pattern for 'nothing to commit'
	filter := NewSmallFilter()
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	})

	// When: Call ShouldFilter with raw=testGitCleanOutput (< 25 tokens)
	raw := testGitCleanOutput
	tokenCount := domain.EstimateTokens(raw)
	result := filter.ShouldFilter(raw, tokenCount, "git", []string{"status"})

	// Then: Returns true
	assert.True(t, result)
}

// TestSmallFilter_DoesNotTriggerForLargeOutput tests that ShouldFilter returns
// false for output exceeding the token threshold.
func TestSmallFilter_DoesNotTriggerForLargeOutput(t *testing.T) {
	// Given: Create SmallFilter with default config
	filter := NewSmallFilter()
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	})

	// When: Call ShouldFilter with raw output > 100 characters (> 25 tokens)
	raw := `On branch main
Your branch is up to date with 'origin/main'.

nothing to commit, working tree clean

This is a very long output that exceeds the token threshold.
It has many lines and many words to ensure we exceed 25 tokens.
The filter should not trigger for this large output.
Additional padding lines to make absolutely sure we exceed the threshold.
More content here to really push the token count over the limit.`
	tokenCount := 100 // Force large token count
	result := filter.ShouldFilter(raw, tokenCount, "git", []string{"status"})

	// Then: Returns false
	assert.False(t, result)
}

// TestSmallFilter_ReturnsCompactStatus tests that Filter returns a
// SmallOutputResult with status and summary.
func TestSmallFilter_ReturnsCompactStatus(t *testing.T) {
	// Given: Create SmallFilter with git status pattern
	filter := NewSmallFilter()
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	})

	// When: Call Filter with testGitCleanOutput
	raw := testGitCleanOutput
	result := filter.Filter(raw)

	// Then: Returns SmallOutputResult with Status='clean' and Summary=testGitCleanOutput
	assert.Equal(t, "clean", result.Status)
	assert.Equal(t, testGitCleanOutput, result.Summary)
}

// TestSmallFilter_RespectsCommandMatching tests that ShouldFilter returns
// false when the command doesn't match any registered pattern.
func TestSmallFilter_RespectsCommandMatching(t *testing.T) {
	// Given: Create SmallFilter with git status pattern only
	filter := NewSmallFilter()
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	})

	// When: Call ShouldFilter with cmd='npm' and small output
	raw := "nothing to commit"
	tokenCount := domain.EstimateTokens(raw)
	result := filter.ShouldFilter(raw, tokenCount, "npm", []string{"install"})

	// Then: Returns false (pattern does not match npm)
	assert.False(t, result)
}

// TestSmallFilter_EmptyOutputTriggers tests that ShouldFilter returns true
// for empty output when a pattern matches empty output.
func TestSmallFilter_EmptyOutputTriggers(t *testing.T) {
	// Given: Create SmallFilter with go build pattern (empty=success)
	filter := NewSmallFilter()
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "go",
		Subcommand: "build",
		Pattern:    regexp.MustCompile(`^$`), // Matches empty output
	})

	// When: Call ShouldFilter with empty string for go build
	raw := ""
	tokenCount := 0
	result := filter.ShouldFilter(raw, tokenCount, "go", []string{"build"})

	// Then: Returns true
	assert.True(t, result)
}

// TestSmallFilter_DisabledViaConfig tests that ShouldFilter returns false
// when the filter is disabled.
func TestSmallFilter_DisabledViaConfig(t *testing.T) {
	// Given: Create SmallFilter with Enabled=false
	config := domain.SmallOutputConfig{
		Enabled:        false,
		TokenThreshold: domain.MinTokenThreshold,
	}
	filter := NewSmallFilterWithConfig(config)
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	})

	// When: Call ShouldFilter with small output
	raw := "nothing to commit"
	tokenCount := domain.EstimateTokens(raw)
	result := filter.ShouldFilter(raw, tokenCount, "git", []string{"status"})

	// Then: Returns false
	assert.False(t, result)
}

// TestSmallFilter_SubcommandMatching tests that ShouldFilter correctly matches
// subcommands.
func TestSmallFilter_SubcommandMatching(t *testing.T) {
	// Given: Create SmallFilter with git diff pattern
	filter := NewSmallFilter()
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "git",
		Subcommand: "diff",
		Pattern:    regexp.MustCompile(`^$`), // Empty diff = no changes
	})

	// When: Call ShouldFilter with empty output for git diff
	raw := ""
	tokenCount := 0
	result := filter.ShouldFilter(raw, tokenCount, "git", []string{"diff"})

	// Then: Returns true
	assert.True(t, result)

	// When: Call ShouldFilter with same output for git status (wrong subcommand)
	result = filter.ShouldFilter(raw, tokenCount, "git", []string{"status"})

	// Then: Returns false
	assert.False(t, result)
}

// TestSmallFilter_NoSubcommands tests that ShouldFilter handles commands
// without subcommands.
func TestSmallFilter_NoSubcommands(t *testing.T) {
	// Given: Create SmallFilter with pattern that has no subcommand
	filter := NewSmallFilter()
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "ls",
		Subcommand: "", // No subcommand
		Pattern:    regexp.MustCompile(`.+`),
	})

	// When: Call ShouldFilter with no subcommands
	raw := "file1.txt"
	tokenCount := 1
	result := filter.ShouldFilter(raw, tokenCount, "ls", nil)

	// Then: Returns true
	assert.True(t, result)
}

// TestSmallFilter_FilterExtractsStatus tests that Filter extracts appropriate
// status from raw output based on patterns.
func TestSmallFilter_FilterExtractsStatus(t *testing.T) {
	tests := []struct {
		name           string
		raw            string
		expectedStatus string
	}{
		{
			name:           "clean git status",
			raw:            testGitCleanOutput,
			expectedStatus: "clean",
		},
		{
			name:           "empty output",
			raw:            "",
			expectedStatus: "empty",
		},
		{
			name:           "success output",
			raw:            "Build succeeded",
			expectedStatus: "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewSmallFilter()

			result := filter.Filter(tt.raw)

			assert.Equal(t, tt.expectedStatus, result.Status)
		})
	}
}

// TestSmallFilter_MultiplePatterns tests that ShouldFilter works with
// multiple registered patterns.
func TestSmallFilter_MultiplePatterns(t *testing.T) {
	// Given: Create SmallFilter with multiple patterns
	filter := NewSmallFilter()
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "git",
		Subcommand: "status",
		Pattern:    regexp.MustCompile(`nothing to commit`),
	})
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "go",
		Subcommand: "build",
		Pattern:    regexp.MustCompile(`^$`),
	})
	filter.RegisterPattern(ports.MinimalPattern{
		Command:    "npm",
		Subcommand: "audit",
		Pattern:    regexp.MustCompile(`found 0 vulnerabilities`),
	})

	// When/Then: Each pattern matches its command
	assert.True(t, filter.ShouldFilter("nothing to commit", 3, "git", []string{"status"}))
	assert.True(t, filter.ShouldFilter("", 0, "go", []string{"build"}))
	assert.True(t, filter.ShouldFilter("found 0 vulnerabilities", 3, "npm", []string{"audit"}))

	// And: Non-matching commands return false
	assert.False(t, filter.ShouldFilter("nothing to commit", 3, "npm", []string{"install"}))
}
