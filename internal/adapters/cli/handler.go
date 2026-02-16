// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Handler manages CLI interactions using cobra.
// It coordinates between user input, command execution, and output formatting.
type Handler struct {
	runner   ports.CommandRunner
	registry ports.ParserRegistry
	rootCmd  *cobra.Command
}

// NewHandler creates a new CLI Handler with the given dependencies.
// The runner executes commands; the registry finds parsers for command output.
func NewHandler(runner ports.CommandRunner, registry ports.ParserRegistry) *Handler {
	h := &Handler{
		runner:   runner,
		registry: registry,
	}
	h.rootCmd = h.buildRootCommand()
	return h
}

// Runner returns the CommandRunner used by this handler.
func (h *Handler) Runner() ports.CommandRunner {
	return h.runner
}

// Registry returns the ParserRegistry used by this handler.
func (h *Handler) Registry() ports.ParserRegistry {
	return h.registry
}

// RootCommand returns the cobra root command.
func (h *Handler) RootCommand() *cobra.Command {
	return h.rootCmd
}

// buildRootCommand creates the cobra root command configuration.
func (h *Handler) buildRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "structured-cli [command] [args...]",
		Short: "A CLI wrapper that transforms command output to structured JSON",
		Long: `structured-cli intercepts CLI commands (git, npm, docker, etc.),
executes them, and optionally outputs structured JSON.

By default, output passes through unchanged (passthrough mode).
Use --json flag or set STRUCTURED_CLI_JSON=true for JSON output.`,
		// Disable automatic parsing of flags after the command
		DisableFlagParsing: true,
		SilenceUsage:       true,
		SilenceErrors:      true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle --help and -h specially since we disable flag parsing
			for _, arg := range args {
				if arg == "--help" || arg == "-h" {
					return cmd.Help()
				}
			}
			envJSON := os.Getenv(EnvJSONKey)
			return h.ExecuteWithArgs(cmd.Context(), args, envJSON, cmd.OutOrStdout())
		},
	}

	return cmd
}

// ExecuteWithArgs runs the handler with the given arguments.
// This is the main entry point for command execution.
//
// The method:
// 1. Extracts the --json flag from args
// 2. Determines output mode from flag and env var
// 3. Parses the remaining args into a Command
// 4. Executes the command via the runner
// 5. Parses output if a parser is registered
// 6. Writes output in the appropriate format
func (h *Handler) ExecuteWithArgs(ctx context.Context, args []string, envJSON string, out io.Writer) error {
	// Extract --json flag and determine output mode
	jsonFlag, remaining := ExtractJSONFlag(args)
	outputJSON := ShouldOutputJSON(jsonFlag, envJSON)

	// Parse args into Command
	if len(remaining) == 0 {
		return errors.New("no command specified")
	}

	cmd, err := domain.CommandFromArgs(remaining)
	if err != nil {
		return fmt.Errorf("parse command: %w", err)
	}

	// Execute the command
	cmdArgs := append(cmd.Subcommands, cmd.Args...)
	stdout, _, exitCode, err := h.runner.Run(ctx, cmd.Name, cmdArgs)
	if err != nil {
		return h.handleError(out, outputJSON, err, exitCode)
	}

	// Read stdout
	rawBytes, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("read stdout: %w", err)
	}
	raw := string(rawBytes)

	// Look up parser for this command
	parser, found := h.registry.Find(cmd.Name, cmd.Subcommands)

	var result domain.ParseResult
	var schema domain.Schema

	if found {
		// Parse the output
		schema = parser.Schema()
		result, err = parser.Parse(strings.NewReader(raw))
		if err != nil {
			result = domain.NewParseResultWithError(
				fmt.Errorf("parse error: %w", err),
				raw,
				exitCode,
			)
		} else {
			result.ExitCode = exitCode
			result.Raw = raw
		}
	} else {
		// No parser - passthrough mode
		result = domain.NewParseResult(nil, raw, exitCode)
	}

	// Write output based on mode
	if err := h.writeOutput(out, outputJSON, result, schema); err != nil {
		return err
	}

	// Propagate non-zero exit code as an ExitError
	if exitCode != 0 {
		return domain.NewExitError(exitCode, nil)
	}

	return nil
}

// handleError writes an error in the appropriate format.
func (h *Handler) handleError(out io.Writer, outputJSON bool, err error, exitCode int) error {
	if outputJSON {
		errResult := map[string]any{
			"error":    err.Error(),
			"exitCode": exitCode,
		}
		return json.NewEncoder(out).Encode(errResult)
	}
	return err
}

// writeOutput writes the result in the appropriate format.
func (h *Handler) writeOutput(out io.Writer, outputJSON bool, result domain.ParseResult, _ domain.Schema) error {
	if outputJSON {
		return h.writeJSON(out, result)
	}
	return h.writePassthrough(out, result)
}

// writeJSON writes the result as JSON.
func (h *Handler) writeJSON(out io.Writer, result domain.ParseResult) error {
	var output any

	if result.Error != nil {
		output = map[string]any{
			"error":    result.Error.Error(),
			"exitCode": result.ExitCode,
			"raw":      result.Raw,
		}
	} else if result.Data != nil {
		output = result.Data
	} else {
		// Unparsed output - passthrough with wrapper
		output = map[string]any{
			"raw":    result.Raw,
			"parsed": false,
		}
	}

	enc := json.NewEncoder(out)
	return enc.Encode(output)
}

// writePassthrough writes the raw output unchanged.
func (h *Handler) writePassthrough(out io.Writer, result domain.ParseResult) error {
	_, err := io.WriteString(out, result.Raw)
	return err
}

// Run executes the CLI with os.Args.
// This is typically called from main().
func (h *Handler) Run() error {
	return h.rootCmd.Execute()
}
