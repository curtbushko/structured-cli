// Package application contains the business logic and use cases for structured-cli.
// This layer orchestrates domain logic and depends on ports (interfaces) - never adapters.
package application

import (
	"regexp"

	"github.com/curtbushko/structured-cli/internal/ports"
)

// minimalPatternDef defines a pattern for detecting minimal/clean command output.
// It extends ports.MinimalPattern with additional metadata for status and summary.
type minimalPatternDef struct {
	// Command is the base command (e.g., "git", "npm", "docker").
	Command string

	// Subcommand is the subcommand being targeted (e.g., "status", "install").
	// Empty string matches any subcommand.
	Subcommand string

	// Pattern is the compiled regex pattern to match against raw output.
	Pattern *regexp.Regexp

	// Status is the status string to return when this pattern matches.
	Status string

	// Summary is the summary string to return when this pattern matches.
	// If empty, the raw output will be used as the summary.
	Summary string
}

// DefaultPatterns returns all minimal output patterns for supported commands.
// These patterns detect known "clean/empty" outputs that can be simplified
// to a compact status JSON response.
//
// Categories:
// - Git: status (clean), stash list (empty), diff (empty)
// - NPM: audit (0 vulnerabilities), outdated (empty)
// - Python: pip-audit (no vulnerabilities)
// - Go: build (empty)
// - Cargo: build (Finished without errors)
// - TypeScript: tsc (empty)
// - Linters: eslint (empty/0 problems), golangci-lint (empty), ruff (empty)
func DefaultPatterns() []minimalPatternDef {
	return []minimalPatternDef{
		// Git patterns
		{
			Command:    "git",
			Subcommand: "status",
			Pattern:    regexp.MustCompile(`(?:nothing to commit|working tree clean)`),
			Status:     StatusClean,
			Summary:    "", // Use raw output as summary
		},
		{
			Command:    "git",
			Subcommand: "stash",
			Pattern:    regexp.MustCompile(`^$`), // Empty output = no stashes
			Status:     StatusEmpty,
			Summary:    "no stashes",
		},
		{
			Command:    "git",
			Subcommand: "diff",
			Pattern:    regexp.MustCompile(`^$`), // Empty diff = no changes
			Status:     StatusClean,
			Summary:    "no changes",
		},

		// NPM patterns
		{
			Command:    "npm",
			Subcommand: "audit",
			Pattern:    regexp.MustCompile(`found 0 vulnerabilities`),
			Status:     StatusClean,
			Summary:    "", // Use raw output as summary
		},
		{
			Command:    "npm",
			Subcommand: "outdated",
			Pattern:    regexp.MustCompile(`^$`), // Empty output = all up to date
			Status:     StatusClean,
			Summary:    "all packages up to date",
		},

		// Python patterns
		{
			Command:    "pip-audit",
			Subcommand: "",
			Pattern:    regexp.MustCompile(`(?i)no known vulnerabilities`),
			Status:     StatusClean,
			Summary:    "", // Use raw output as summary
		},

		// Go patterns
		{
			Command:    "go",
			Subcommand: "build",
			Pattern:    regexp.MustCompile(`^$`), // Empty output = success
			Status:     StatusSuccess,
			Summary:    "build succeeded",
		},

		// Cargo patterns
		{
			Command:    "cargo",
			Subcommand: "build",
			Pattern:    regexp.MustCompile(`Finished.*target`),
			Status:     StatusSuccess,
			Summary:    "", // Use raw output as summary
		},

		// TypeScript patterns
		{
			Command:    "tsc",
			Subcommand: "",
			Pattern:    regexp.MustCompile(`^$`), // Empty output = no errors
			Status:     StatusSuccess,
			Summary:    "no errors",
		},

		// Linter patterns
		{
			Command:    "eslint",
			Subcommand: "",
			Pattern:    regexp.MustCompile(`^$`), // Empty output = no issues
			Status:     StatusClean,
			Summary:    "no issues",
		},
		{
			Command:    "golangci-lint",
			Subcommand: "run",
			Pattern:    regexp.MustCompile(`^$`), // Empty output = no issues
			Status:     StatusClean,
			Summary:    "no issues",
		},
		{
			Command:    "ruff",
			Subcommand: "check",
			Pattern:    regexp.MustCompile(`^$`), // Empty output = no issues
			Status:     StatusClean,
			Summary:    "no issues",
		},
	}
}

// ToMinimalPatterns converts a slice of minimalPatternDef to ports.MinimalPattern.
// This allows the internal pattern definitions to be used with the SmallFilter.
func ToMinimalPatterns(defs []minimalPatternDef) []ports.MinimalPattern {
	patterns := make([]ports.MinimalPattern, len(defs))
	for i, def := range defs {
		patterns[i] = ports.MinimalPattern{
			Command:    def.Command,
			Subcommand: def.Subcommand,
			Pattern:    def.Pattern,
		}
	}
	return patterns
}
