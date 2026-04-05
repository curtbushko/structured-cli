// Package application contains the business logic and use cases for structured-cli.
package application

import "github.com/curtbushko/structured-cli/internal/domain"

// testRunnerRules defines filter rules for test runner output.
// Each rule specifies which field contains the status and what values indicate pass/fail.
var testRunnerRules = map[string]domain.FilterRule{
	// Jest: status field with passed/failed values
	"jest": domain.NewFilterRule(
		"status",
		[]string{"passed"},
		[]string{"failed"},
	),
	// Pytest: outcome field with passed/failed/error values
	"pytest": domain.NewFilterRule(
		"outcome",
		[]string{"passed"},
		[]string{"failed", "error"},
	),
	// Vitest: status field with pass/fail values
	"vitest": domain.NewFilterRule(
		"status",
		[]string{"pass"},
		[]string{"fail"},
	),
	// Go test: action field with pass/fail values
	"gotest": domain.NewFilterRule(
		"action",
		[]string{"pass"},
		[]string{"fail"},
	),
	// Cargo test: status field with ok/FAILED values
	"cargotest": domain.NewFilterRule(
		"status",
		[]string{"ok"},
		[]string{"FAILED", "failed"},
	),
	// Mocha: state field with passed/failed values
	"mocha": domain.NewFilterRule(
		"state",
		[]string{"passed"},
		[]string{"failed"},
	),
}

// linterRules defines filter rules for linter output.
// Some linters filter by severity, others keep all issues.
var linterRules = map[string]domain.FilterRule{
	// ESLint: severity field - keep errors, filter warnings
	"eslint": domain.NewFilterRule(
		"severity",
		[]string{"warning"}, // warnings are "pass" (filtered out)
		[]string{"error"},   // errors are "fail" (kept)
	),
	// golangci-lint: keep all issues (no pass values means nothing filtered)
	"golangci-lint": domain.NewFilterRule(
		"severity",
		[]string{}, // empty means keep all
		[]string{},
	),
	// tsc: keep all errors (TypeScript only outputs errors)
	"tsc": domain.NewFilterRule(
		"",
		[]string{},
		[]string{},
	),
	// ruff: keep all violations (no severity levels)
	"ruff": domain.NewFilterRule(
		"",
		[]string{},
		[]string{},
	),
}

// commandToParser maps CLI commands to their parser rule names.
// This allows commands like "npm test" to use "jest" rules.
var commandToParser = map[string]string{
	// Test runner mappings
	"npm:test":     "jest",
	"npm:run test": "jest",
	"npx:jest":     "jest",
	"yarn:jest":    "jest",
	"pnpm:jest":    "jest",
	"pytest":       "pytest",
	"npx:vitest":   "vitest",
	"yarn:vitest":  "vitest",
	"pnpm:vitest":  "vitest",
	"go:test":      "gotest",
	"cargo:test":   "cargotest",
	"npx:mocha":    "mocha",
	"yarn:mocha":   "mocha",
	"pnpm:mocha":   "mocha",
	// Linter mappings
	"eslint":            "eslint",
	"npx:eslint":        "eslint",
	"golangci-lint:run": "golangci-lint",
	"tsc":               "tsc",
	"npx:tsc":           "tsc",
	"ruff:check":        "ruff",
}

// getRule returns the FilterRule for a given command and subcommands.
// Returns the rule and true if found, or an empty rule and false if not found.
func (f *SuccessFilterer) getRule(cmd string, subcmds []string) (domain.FilterRule, bool) {
	// Build lookup key
	key := cmd
	if len(subcmds) > 0 {
		key = cmd + ":" + subcmds[0]
	}

	// Look up parser name
	parserName, exists := commandToParser[key]
	if !exists {
		// Try with just the command (e.g., "pytest" with no subcommands)
		parserName, exists = commandToParser[cmd]
		if !exists {
			return domain.FilterRule{}, false
		}
	}

	// Check test runner rules first
	if rule, ok := testRunnerRules[parserName]; ok {
		return rule, true
	}

	// Check linter rules
	if rule, ok := linterRules[parserName]; ok {
		return rule, true
	}

	return domain.FilterRule{}, false
}
