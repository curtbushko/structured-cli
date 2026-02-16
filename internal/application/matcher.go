package application

import (
	"errors"
	"strings"
)

// ErrEmptyArgs is returned when Match is called with an empty args slice.
var ErrEmptyArgs = errors.New("empty argument list")

// CommandMatcher parses command-line arguments to extract the command,
// subcommands, and remaining arguments (flags and positional args).
//
// The matching logic follows these rules:
//  1. The first argument is always the command name
//  2. Subsequent arguments that don't start with "-" are treated as subcommands
//  3. Once an argument starting with "-" is encountered, all remaining arguments
//     are treated as "remaining" (flags and their values)
//
// Examples:
//   - ["git", "log", "--oneline"] -> cmd="git", subcommands=["log"], remaining=["--oneline"]
//   - ["docker", "compose", "up", "-d"] -> cmd="docker", subcommands=["compose", "up"], remaining=["-d"]
//   - ["ls", "-la"] -> cmd="ls", subcommands=[], remaining=["-la"]
type CommandMatcher struct{}

// NewCommandMatcher creates a new CommandMatcher.
func NewCommandMatcher() *CommandMatcher {
	return &CommandMatcher{}
}

// Match parses the argument list and returns:
// - cmd: the base command name (e.g., "git", "docker")
// - subcommands: the subcommand chain (e.g., ["status"], ["compose", "up"])
// - remaining: flags and their values (e.g., ["--oneline", "-n", "5"])
// - error: ErrEmptyArgs if args is empty
//
// The function stops collecting subcommands when it encounters an argument
// starting with "-", as this indicates the start of flags/options.
func (m *CommandMatcher) Match(args []string) (cmd string, subcommands []string, remaining []string, err error) {
	if len(args) == 0 {
		return "", nil, nil, ErrEmptyArgs
	}

	cmd = args[0]

	// Process remaining arguments
	for i := 1; i < len(args); i++ {
		arg := args[i]
		// If we hit a flag (starts with -), everything from here is "remaining"
		if strings.HasPrefix(arg, "-") {
			remaining = args[i:]
			break
		}
		subcommands = append(subcommands, arg)
	}

	// Normalize empty slices to nil for consistent comparison
	if len(subcommands) == 0 {
		subcommands = nil
	}
	if len(remaining) == 0 {
		remaining = nil
	}

	return cmd, subcommands, remaining, nil
}
