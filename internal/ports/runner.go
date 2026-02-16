// Package ports defines the interfaces (contracts) for the structured-cli.
// This layer only imports from domain - never from adapters or services.
// Adapters implement these interfaces; services depend on them.
package ports

import (
	"context"
	"io"
)

// CommandRunner executes shell commands and returns their output streams.
// Implementations handle the actual process execution (e.g., os/exec).
//
// The interface returns io.Reader for stdout/stderr to enable streaming
// and composability with other io.Reader consumers.
type CommandRunner interface {
	// Run executes a command with the given arguments and returns output streams.
	// The ctx parameter allows cancellation of long-running commands.
	// Returns stdout and stderr as readers, the exit code, and any execution error.
	//
	// An error is returned only for execution failures (command not found,
	// permission denied, etc.), not for non-zero exit codes.
	Run(ctx context.Context, cmd string, args []string) (stdout, stderr io.Reader, exitCode int, err error)
}
