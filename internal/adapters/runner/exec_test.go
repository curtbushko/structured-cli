// Package runner provides implementations of the CommandRunner port.
package runner

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestExecRunner_Run_SimpleCommand(t *testing.T) {
	// Arrange
	runner := NewExecRunner()
	ctx := context.Background()

	// Act
	stdout, stderr, exitCode, err := runner.Run(ctx, "echo", []string{"hello"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	stderrBytes, err := io.ReadAll(stderr)
	if err != nil {
		t.Fatalf("failed to read stderr: %v", err)
	}

	if !strings.Contains(string(stdoutBytes), "hello") {
		t.Errorf("stdout = %q, want to contain 'hello'", string(stdoutBytes))
	}

	if len(stderrBytes) != 0 {
		t.Errorf("stderr = %q, want empty", string(stderrBytes))
	}

	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
}

func TestExecRunner_Run_WithArgs(t *testing.T) {
	// Arrange
	runner := NewExecRunner()
	ctx := context.Background()

	// Act - use printf which handles args predictably across platforms
	stdout, _, exitCode, err := runner.Run(ctx, "printf", []string{"%s %s", "arg1", "arg2"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	output := string(stdoutBytes)
	if !strings.Contains(output, "arg1") || !strings.Contains(output, "arg2") {
		t.Errorf("stdout = %q, want to contain 'arg1' and 'arg2'", output)
	}

	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
}

func TestExecRunner_Run_Failure(t *testing.T) {
	// Arrange
	runner := NewExecRunner()
	ctx := context.Background()

	// Act - 'false' command exits with code 1
	_, _, exitCode, err := runner.Run(ctx, "false", nil)

	// Assert
	// Error should be nil - non-zero exit is not an execution error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if exitCode == 0 {
		t.Errorf("exitCode = %d, want non-zero", exitCode)
	}
}

func TestExecRunner_Run_WithStderr(t *testing.T) {
	// Arrange
	runner := NewExecRunner()
	ctx := context.Background()

	// Act - use sh to write to stderr
	_, stderr, exitCode, err := runner.Run(ctx, "sh", []string{"-c", "echo error_output >&2"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stderrBytes, err := io.ReadAll(stderr)
	if err != nil {
		t.Fatalf("failed to read stderr: %v", err)
	}

	if !strings.Contains(string(stderrBytes), "error_output") {
		t.Errorf("stderr = %q, want to contain 'error_output'", string(stderrBytes))
	}

	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
}

func TestExecRunner_Run_ContextCancelled(t *testing.T) {
	// Arrange
	runner := NewExecRunner()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Act - try to run a command with canceled context
	_, _, _, err := runner.Run(ctx, "sleep", []string{"10"})

	// Assert - should return an error due to context cancellation
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestExecRunner_Run_ContextTimeout(t *testing.T) {
	// Arrange
	runner := NewExecRunner()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Act - run a long command that should be killed by timeout
	_, _, _, err := runner.Run(ctx, "sleep", []string{"10"})

	// Assert - should return an error due to context timeout
	if err == nil {
		t.Fatal("expected error due to context timeout, got nil")
	}
}

func TestExecRunner_Run_CommandNotFound(t *testing.T) {
	// Arrange
	runner := NewExecRunner()
	ctx := context.Background()

	// Act - try to run a non-existent command
	_, _, _, err := runner.Run(ctx, "nonexistent_command_xyz_123", nil)

	// Assert - should return an error for command not found
	if err == nil {
		t.Fatal("expected error for command not found, got nil")
	}
}

func TestExecRunner_Run_EmptyCommand(t *testing.T) {
	// Arrange
	runner := NewExecRunner()
	ctx := context.Background()

	// Act - try to run with empty command
	_, _, _, err := runner.Run(ctx, "", nil)

	// Assert - should return an error for empty command
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}
}

func TestExecRunner_ImplementsCommandRunner(t *testing.T) {
	// This test verifies that ExecRunner implements the CommandRunner interface
	// at compile time. If it doesn't, this won't compile.
	var _ interface {
		Run(ctx context.Context, cmd string, args []string) (stdout, stderr io.Reader, exitCode int, err error)
	} = (*ExecRunner)(nil)
}

// TestExecRunner_Run_ReadersAreIndependent verifies that stdout and stderr
// can be read independently without blocking each other.
func TestExecRunner_Run_ReadersAreIndependent(t *testing.T) {
	// Arrange
	runner := NewExecRunner()
	ctx := context.Background()

	// Act - command that writes to both stdout and stderr
	stdout, stderr, exitCode, err := runner.Run(ctx, "sh", []string{"-c", "echo stdout_msg; echo stderr_msg >&2"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read stdout first
	var stdoutBuf bytes.Buffer
	if _, err := io.Copy(&stdoutBuf, stdout); err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}

	// Then read stderr
	var stderrBuf bytes.Buffer
	if _, err := io.Copy(&stderrBuf, stderr); err != nil {
		t.Fatalf("failed to read stderr: %v", err)
	}

	if !strings.Contains(stdoutBuf.String(), "stdout_msg") {
		t.Errorf("stdout = %q, want to contain 'stdout_msg'", stdoutBuf.String())
	}

	if !strings.Contains(stderrBuf.String(), "stderr_msg") {
		t.Errorf("stderr = %q, want to contain 'stderr_msg'", stderrBuf.String())
	}

	if exitCode != 0 {
		t.Errorf("exitCode = %d, want 0", exitCode)
	}
}
