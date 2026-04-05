package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for command names.
const (
	testCmdGit         = "git"
	testCmdNpm         = "npm"
	testCmdGo          = "go"
	testCmdCargo       = "cargo"
	testCmdEslint      = "eslint"
	testCmdGolangciLnt = "golangci-lint"
	testCmdRuff        = "ruff"
	testCmdTsc         = "tsc"
	testCmdPipAudit    = "pip-audit"
)

// TestDefaultPatterns_ReturnsSlice tests that DefaultPatterns returns a slice
// of MinimalPatterns for all supported commands.
func TestDefaultPatterns_ReturnsSlice(t *testing.T) {
	// When: Call DefaultPatterns()
	patterns := DefaultPatterns()

	// Then: Returns a non-empty slice of patterns
	require.NotNil(t, patterns)
	assert.Greater(t, len(patterns), 0, "DefaultPatterns should return at least one pattern")
}

// TestPatterns_GitStatusClean tests that the git status clean pattern matches
// "nothing to commit, working tree clean" output.
func TestPatterns_GitStatusClean(t *testing.T) {
	// Given: Get git status clean pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var gitStatusPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdGit && patterns[i].Subcommand == "status" {
			gitStatusPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, gitStatusPattern, "git status pattern should exist")

	// When: Test pattern against 'nothing to commit, working tree clean'
	raw := "nothing to commit, working tree clean"
	matches := gitStatusPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match 'nothing to commit, working tree clean'")
	assert.Equal(t, "clean", gitStatusPattern.Status)
}

// TestPatterns_GitStatusDirty tests that the git status pattern does not match
// output with staged files (actionable content).
func TestPatterns_GitStatusDirty(t *testing.T) {
	// Given: Get git status pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var gitStatusPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdGit && patterns[i].Subcommand == "status" {
			gitStatusPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, gitStatusPattern, "git status pattern should exist")

	// When: Test pattern against output with staged files
	raw := `On branch main
Changes to be committed:
  (use "git restore --staged <file>..." to unstage)
        modified:   README.md`
	matches := gitStatusPattern.Pattern.MatchString(raw)

	// Then: Pattern does not match (has actionable content)
	assert.False(t, matches, "pattern should not match output with staged files")
}

// TestPatterns_GitDiffEmpty tests that the git diff pattern matches empty output.
func TestPatterns_GitDiffEmpty(t *testing.T) {
	// Given: Get git diff pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var gitDiffPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdGit && patterns[i].Subcommand == "diff" {
			gitDiffPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, gitDiffPattern, "git diff pattern should exist")

	// When: Test pattern against empty string
	raw := ""
	matches := gitDiffPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match empty string")
	assert.Equal(t, "clean", gitDiffPattern.Status)
	assert.Equal(t, "no changes", gitDiffPattern.Summary)
}

// TestPatterns_GoBuildSuccess tests that the go build pattern matches empty output.
func TestPatterns_GoBuildSuccess(t *testing.T) {
	// Given: Get go build pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var goBuildPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdGo && patterns[i].Subcommand == "build" {
			goBuildPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, goBuildPattern, "go build pattern should exist")

	// When: Test pattern against empty string
	raw := ""
	matches := goBuildPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match empty string")
	assert.Equal(t, "success", goBuildPattern.Status)
	assert.Equal(t, "build succeeded", goBuildPattern.Summary)
}

// TestPatterns_NpmAuditClean tests that the npm audit pattern matches
// "found 0 vulnerabilities" output.
func TestPatterns_NpmAuditClean(t *testing.T) {
	// Given: Get npm audit pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var npmAuditPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdNpm && patterns[i].Subcommand == "audit" {
			npmAuditPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, npmAuditPattern, "npm audit pattern should exist")

	// When: Test pattern against 'found 0 vulnerabilities'
	raw := "found 0 vulnerabilities"
	matches := npmAuditPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match 'found 0 vulnerabilities'")
	assert.Equal(t, "clean", npmAuditPattern.Status)
}

// TestPatterns_ESLintClean tests that the eslint pattern matches empty output
// or '0 problems'.
func TestPatterns_ESLintClean(t *testing.T) {
	// Given: Get eslint pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var eslintPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdEslint && patterns[i].Subcommand == "" {
			eslintPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, eslintPattern, "eslint pattern should exist")

	// When: Test pattern against empty string
	raw := ""
	matchesEmpty := eslintPattern.Pattern.MatchString(raw)

	// Then: Pattern matches empty
	assert.True(t, matchesEmpty, "pattern should match empty string")
	assert.Equal(t, "clean", eslintPattern.Status)
	assert.Equal(t, "no issues", eslintPattern.Summary)
}

// TestPatterns_GolangciLintClean tests that the golangci-lint pattern matches
// empty output.
func TestPatterns_GolangciLintClean(t *testing.T) {
	// Given: Get golangci-lint pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var golangciLintPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdGolangciLnt && patterns[i].Subcommand == "run" {
			golangciLintPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, golangciLintPattern, "golangci-lint pattern should exist")

	// When: Test pattern against empty string
	raw := ""
	matches := golangciLintPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match empty string")
	assert.Equal(t, "clean", golangciLintPattern.Status)
	assert.Equal(t, "no issues", golangciLintPattern.Summary)
}

// TestPatterns_CargoBuildFinished tests that the cargo build pattern matches
// "Finished" output without warnings/errors.
func TestPatterns_CargoBuildFinished(t *testing.T) {
	// Given: Get cargo build pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var cargoBuildPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdCargo && patterns[i].Subcommand == "build" {
			cargoBuildPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, cargoBuildPattern, "cargo build pattern should exist")

	// When: Test pattern against 'Finished release [optimized] target(s)'
	raw := "   Compiling myproject v0.1.0\n    Finished release [optimized] target(s) in 1.23s"
	matches := cargoBuildPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match 'Finished' output")
	assert.Equal(t, "success", cargoBuildPattern.Status)
}

// TestPatterns_GitStashListEmpty tests that the git stash list pattern matches
// empty output.
func TestPatterns_GitStashListEmpty(t *testing.T) {
	// Given: Get git stash list pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var found bool
	for i := range patterns {
		if patterns[i].Command == testCmdGit && patterns[i].Subcommand == "stash" {
			// When: Test pattern against empty string
			raw := ""
			matches := patterns[i].Pattern.MatchString(raw)

			// Then: Pattern matches
			assert.True(t, matches, "pattern should match empty string")
			assert.Equal(t, "empty", patterns[i].Status)
			found = true
			break
		}
	}
	require.True(t, found, "git stash pattern should exist")
}

// TestPatterns_NpmOutdatedEmpty tests that the npm outdated pattern matches
// empty output.
func TestPatterns_NpmOutdatedEmpty(t *testing.T) {
	// Given: Get npm outdated pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var npmOutdatedPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdNpm && patterns[i].Subcommand == "outdated" {
			npmOutdatedPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, npmOutdatedPattern, "npm outdated pattern should exist")

	// When: Test pattern against empty string
	raw := ""
	matches := npmOutdatedPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match empty string")
	assert.Equal(t, "clean", npmOutdatedPattern.Status)
	assert.Equal(t, "all packages up to date", npmOutdatedPattern.Summary)
}

// TestPatterns_PipAuditClean tests that the pip-audit pattern matches
// "No known vulnerabilities" output.
func TestPatterns_PipAuditClean(t *testing.T) {
	// Given: Get pip-audit pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var pipAuditPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdPipAudit && patterns[i].Subcommand == "" {
			pipAuditPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, pipAuditPattern, "pip-audit pattern should exist")

	// When: Test pattern against 'No known vulnerabilities found'
	raw := "No known vulnerabilities found"
	matches := pipAuditPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match 'No known vulnerabilities found'")
	assert.Equal(t, "clean", pipAuditPattern.Status)
}

// TestPatterns_TscEmpty tests that the tsc pattern matches empty output.
func TestPatterns_TscEmpty(t *testing.T) {
	// Given: Get tsc pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var tscPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdTsc && patterns[i].Subcommand == "" {
			tscPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, tscPattern, "tsc pattern should exist")

	// When: Test pattern against empty string
	raw := ""
	matches := tscPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match empty string")
	assert.Equal(t, "success", tscPattern.Status)
	assert.Equal(t, "no errors", tscPattern.Summary)
}

// TestPatterns_RuffEmpty tests that the ruff pattern matches empty output.
func TestPatterns_RuffEmpty(t *testing.T) {
	// Given: Get ruff pattern from DefaultPatterns()
	patterns := DefaultPatterns()
	var ruffPattern *minimalPatternDef
	for i := range patterns {
		if patterns[i].Command == testCmdRuff && patterns[i].Subcommand == "check" {
			ruffPattern = &patterns[i]
			break
		}
	}
	require.NotNil(t, ruffPattern, "ruff pattern should exist")

	// When: Test pattern against empty string
	raw := ""
	matches := ruffPattern.Pattern.MatchString(raw)

	// Then: Pattern matches
	assert.True(t, matches, "pattern should match empty string")
	assert.Equal(t, "clean", ruffPattern.Status)
	assert.Equal(t, "no issues", ruffPattern.Summary)
}

// TestPatterns_ToMinimalPatterns tests that minimalPatternDef can be converted
// to ports.MinimalPattern.
func TestPatterns_ToMinimalPatterns(t *testing.T) {
	// Given: Get patterns from DefaultPatterns()
	defs := DefaultPatterns()
	require.NotEmpty(t, defs)

	// When: Convert to ports.MinimalPattern slice
	patterns := ToMinimalPatterns(defs)

	// Then: Returns matching slice length
	assert.Len(t, patterns, len(defs))

	// And: First pattern has correct command/subcommand/pattern
	assert.Equal(t, defs[0].Command, patterns[0].Command)
	assert.Equal(t, defs[0].Subcommand, patterns[0].Subcommand)
	assert.Equal(t, defs[0].Pattern, patterns[0].Pattern)
}
