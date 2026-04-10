// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"strings"
)

const (
	// JSONFlag is the flag used to enable JSON output mode.
	JSONFlag = "--json"

	// EnvJSONKey is the environment variable that controls JSON output mode.
	EnvJSONKey = "STRUCTURED_CLI_JSON"
)

// ExtractJSONFlag scans args for the --json flag, removes it, and returns
// whether it was found along with the remaining arguments.
//
// The --json flag can appear anywhere in the argument list. All occurrences
// are removed from the returned args slice.
//
// Example:
//
//	ExtractJSONFlag([]string{"git", "--json", "status"})
//	// Returns: true, []string{"git", "status"}
func ExtractJSONFlag(args []string) (jsonFound bool, remaining []string) {
	remaining = make([]string, 0, len(args))

	for _, arg := range args {
		if arg == JSONFlag {
			jsonFound = true
		} else {
			remaining = append(remaining, arg)
		}
	}

	return jsonFound, remaining
}

// ShouldOutputJSON determines whether output should be in JSON format
// based on the --json flag and STRUCTURED_CLI_JSON environment variable.
//
// Precedence (highest to lowest):
// 1. --json flag (if present, always returns true)
// 2. STRUCTURED_CLI_JSON environment variable ("true", "TRUE", "1" = JSON mode)
// 3. Default: passthrough mode (returns false)
//
// This follows the principle that explicit flags override environment variables.
func ShouldOutputJSON(flagJSON bool, envValue string) bool {
	// Flag takes highest precedence
	if flagJSON {
		return true
	}

	// Check environment variable
	envLower := strings.ToLower(envValue)
	return envLower == "true" || envValue == "1"
}
