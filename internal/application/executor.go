// Package application contains the business logic and use cases for structured-cli.
// This layer orchestrates domain logic and depends on ports (interfaces) - never adapters.
package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Executor orchestrates command execution, parsing, and output writing.
// It is the main application service that coordinates between:
// - CommandRunner: executes the actual CLI commands
// - ParserRegistry: finds the appropriate parser for a command
// - OutputWriter: formats and writes the result
// - TrackingRecorder: records command execution metrics (optional)
type Executor struct {
	runner   ports.CommandRunner
	registry ports.ParserRegistry
	writer   ports.OutputWriter
	tracker  ports.TrackingRecorder
}

// NewExecutor creates a new Executor with the given dependencies.
// All dependencies are injected via constructor for testability.
// Tracking is disabled (nil tracker).
func NewExecutor(runner ports.CommandRunner, registry ports.ParserRegistry, writer ports.OutputWriter) *Executor {
	return &Executor{
		runner:   runner,
		registry: registry,
		writer:   writer,
		tracker:  nil,
	}
}

// NewExecutorWithTracker creates a new Executor with tracking support.
// If tracker is nil, tracking is disabled.
func NewExecutorWithTracker(
	runner ports.CommandRunner,
	registry ports.ParserRegistry,
	writer ports.OutputWriter,
	tracker ports.TrackingRecorder,
) *Executor {
	return &Executor{
		runner:   runner,
		registry: registry,
		writer:   writer,
		tracker:  tracker,
	}
}

// Execute runs a command and writes the result to the output writer.
// If a parser is registered for the command, it parses the output.
// Otherwise, it passes through the raw output.
//
// The flow is:
// 1. Run the command via CommandRunner (timed)
// 2. Look up a parser in the registry
// 3. If parser found: parse output, handle any parse errors
// 4. If no parser: create passthrough result with raw output
// 5. Track the execution (if tracker is configured)
// 6. Write result via OutputWriter
func (e *Executor) Execute(ctx context.Context, cmd domain.Command, out io.Writer) error {
	// Start timing the execution
	startTime := time.Now()

	// Run the command
	stdout, _, exitCode, err := e.runner.Run(ctx, cmd.Name, append(cmd.Subcommands, cmd.Args...))
	if err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	// Read stdout into a string for raw output preservation
	rawBytes, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("read stdout: %w", err)
	}
	raw := string(rawBytes)

	// Look up parser for this command
	parser, found := e.registry.Find(cmd.Name, cmd.Subcommands)

	var result domain.ParseResult
	var schema domain.Schema
	var parseErr error

	if found {
		// Parse the output
		schema = parser.Schema()
		result, parseErr = parser.Parse(strings.NewReader(raw))
		if parseErr != nil {
			// Parser failed - create error result with raw output preserved
			result = domain.NewParseResultWithError(
				fmt.Errorf("parse error: %w", parseErr),
				raw,
				exitCode,
			)
		} else {
			// Update the exit code in the result
			result.ExitCode = exitCode
			result.Raw = raw
			// Sync exit code and success fields in the result data
			syncExitCodeAndSuccess(result.Data, exitCode)
		}
	} else {
		// No parser - create fallback result with raw output
		// This allows JSON mode to return {"raw": "...", "parsed": false, "exitCode": N}
		// while passthrough mode can extract just the raw output
		fallbackResult := domain.NewFallbackResult(raw, exitCode)
		result = domain.NewParseResult(fallbackResult, raw, exitCode)
	}

	// Calculate execution time
	execTime := time.Since(startTime)

	// Track the execution (errors are logged but do not fail command execution)
	e.trackExecution(ctx, cmd, raw, result, execTime, parseErr)

	// Write the result
	if err := e.writer.Write(out, result, schema); err != nil {
		return fmt.Errorf("write result: %w", err)
	}

	return nil
}

// trackExecution records the command execution metrics if tracking is enabled.
// Any tracking errors are ignored to avoid failing command execution.
func (e *Executor) trackExecution(
	ctx context.Context,
	cmd domain.Command,
	raw string,
	result domain.ParseResult,
	execTime time.Duration,
	parseErr error,
) {
	if e.tracker == nil {
		return
	}

	// If there was a parse error, record it as a failure
	if parseErr != nil {
		e.recordParseFailure(ctx, cmd, parseErr)
		return
	}

	// Calculate token metrics
	rawTokens := domain.EstimateTokens(raw)
	parsedJSON := e.serializeResult(result.Data)
	parsedTokens := domain.EstimateTokens(parsedJSON)

	// Get current working directory for project context
	project, _ := os.Getwd()

	// Create and record the command record
	record := domain.NewCommandRecord(
		cmd.Name,
		cmd.Subcommands,
		rawTokens,
		parsedTokens,
		execTime,
		project,
	)

	// Record errors are logged but don't fail execution
	_ = e.tracker.Record(ctx, record)
}

// recordParseFailure records a parse failure for tracking.
func (e *Executor) recordParseFailure(ctx context.Context, cmd domain.Command, parseErr error) {
	// Build full command string
	cmdParts := append([]string{cmd.Name}, cmd.Subcommands...)
	fullCmd := strings.Join(cmdParts, " ")

	failure := domain.NewParseFailure(fullCmd, parseErr.Error(), true)
	_ = e.tracker.RecordFailure(ctx, failure)
}

// serializeResult converts result data to JSON for token counting.
func (e *Executor) serializeResult(data any) string {
	if data == nil {
		return ""
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

// syncExitCodeAndSuccess updates the ExitCode and Success fields in the result data
// based on the actual exit code from command execution. This handles cases where
// the parser couldn't detect failure from the output (e.g., silent exit with @exit 1).
func syncExitCodeAndSuccess(data any, exitCode int) {
	if data == nil {
		return
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	// Update ExitCode field if it exists
	exitCodeField := v.FieldByName("ExitCode")
	if exitCodeField.IsValid() && exitCodeField.CanSet() && exitCodeField.Kind() == reflect.Int {
		exitCodeField.SetInt(int64(exitCode))
	}

	// Update Success field based on exit code if it exists
	successField := v.FieldByName("Success")
	if successField.IsValid() && successField.CanSet() && successField.Kind() == reflect.Bool {
		// Only mark as failure if exit code is non-zero
		// Don't override an already-false Success (parser may have detected error)
		if exitCode != 0 {
			successField.SetBool(false)
		}
	}
}
