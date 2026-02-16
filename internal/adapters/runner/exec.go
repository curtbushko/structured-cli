// Package runner provides implementations of the CommandRunner port.
// This package is in the adapters layer and implements the ports.CommandRunner
// interface using os/exec for actual command execution.
package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/curtbushko/structured-cli/internal/ports"
)

// ErrEmptyCommand is returned when an empty command string is provided.
var ErrEmptyCommand = errors.New("command cannot be empty")

// Compile-time check that ExecRunner implements ports.CommandRunner.
var _ ports.CommandRunner = (*ExecRunner)(nil)

// ExecRunner implements ports.CommandRunner using os/exec.
// It executes shell commands and captures their output streams.
type ExecRunner struct{}

// NewExecRunner creates a new ExecRunner instance.
func NewExecRunner() *ExecRunner {
	return &ExecRunner{}
}

// Run executes a command with the given arguments and returns output streams.
// The ctx parameter allows cancellation of long-running commands.
// Returns stdout and stderr as readers, the exit code, and any execution error.
//
// An error is returned only for execution failures (command not found,
// permission denied, context cancellation, etc.), not for non-zero exit codes.
func (r *ExecRunner) Run(ctx context.Context, cmd string, args []string) (stdout, stderr io.Reader, exitCode int, err error) {
	if cmd == "" {
		return nil, nil, 0, ErrEmptyCommand
	}

	// Check if context is already canceled before starting
	if err := ctx.Err(); err != nil {
		return nil, nil, 0, fmt.Errorf("context error before execution: %w", err)
	}

	// Create command with context for cancellation support
	command := exec.CommandContext(ctx, cmd, args...)

	// Capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	command.Stdout = &stdoutBuf
	command.Stderr = &stderrBuf

	// Run the command
	runErr := command.Run()

	// Get exit code
	exitCode = 0
	if runErr != nil {
		// Check if this is a context cancellation
		if ctx.Err() != nil {
			return nil, nil, 0, fmt.Errorf("command execution interrupted: %w", ctx.Err())
		}

		// Check if it's an exit error (non-zero exit code)
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
			// Non-zero exit is not an execution error, just return the exit code
			return &stdoutBuf, &stderrBuf, exitCode, nil
		}

		// Other errors (command not found, permission denied, etc.)
		return nil, nil, 0, fmt.Errorf("command execution failed: %w", runErr)
	}

	return &stdoutBuf, &stderrBuf, exitCode, nil
}
