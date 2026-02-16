// Package application contains the business logic and use cases for structured-cli.
// This layer orchestrates domain logic and depends on ports (interfaces) - never adapters.
package application

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Executor orchestrates command execution, parsing, and output writing.
// It is the main application service that coordinates between:
// - CommandRunner: executes the actual CLI commands
// - ParserRegistry: finds the appropriate parser for a command
// - OutputWriter: formats and writes the result
type Executor struct {
	runner   ports.CommandRunner
	registry ports.ParserRegistry
	writer   ports.OutputWriter
}

// NewExecutor creates a new Executor with the given dependencies.
// All dependencies are injected via constructor for testability.
func NewExecutor(runner ports.CommandRunner, registry ports.ParserRegistry, writer ports.OutputWriter) *Executor {
	return &Executor{
		runner:   runner,
		registry: registry,
		writer:   writer,
	}
}

// Execute runs a command and writes the result to the output writer.
// If a parser is registered for the command, it parses the output.
// Otherwise, it passes through the raw output.
//
// The flow is:
// 1. Run the command via CommandRunner
// 2. Look up a parser in the registry
// 3. If parser found: parse output, handle any parse errors
// 4. If no parser: create passthrough result with raw output
// 5. Write result via OutputWriter
func (e *Executor) Execute(ctx context.Context, cmd domain.Command, out io.Writer) error {
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

	if found {
		// Parse the output
		schema = parser.Schema()
		result, err = parser.Parse(strings.NewReader(raw))
		if err != nil {
			// Parser failed - create error result with raw output preserved
			result = domain.NewParseResultWithError(
				fmt.Errorf("parse error: %w", err),
				raw,
				exitCode,
			)
		} else {
			// Update the exit code in the result
			result.ExitCode = exitCode
			result.Raw = raw
		}
	} else {
		// No parser - create fallback result with raw output
		// This allows JSON mode to return {"raw": "...", "parsed": false, "exitCode": N}
		// while passthrough mode can extract just the raw output
		fallbackResult := domain.NewFallbackResult(raw, exitCode)
		result = domain.NewParseResult(fallbackResult, raw, exitCode)
	}

	// Write the result
	if err := e.writer.Write(out, result, schema); err != nil {
		return fmt.Errorf("write result: %w", err)
	}

	return nil
}
