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
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Handler manages CLI interactions using cobra.
// It coordinates between user input, command execution, and output formatting.
type Handler struct {
	runner      ports.CommandRunner
	registry    ports.ParserRegistry
	tracker     ports.Tracker
	smallFilter ports.SmallOutputFilter
	rootCmd     *cobra.Command
}

// NewHandler creates a new CLI Handler with the given dependencies.
// The runner executes commands; the registry finds parsers for command output.
// Tracking is disabled (nil tracker).
func NewHandler(runner ports.CommandRunner, registry ports.ParserRegistry) *Handler {
	h := &Handler{
		runner:   runner,
		registry: registry,
		tracker:  nil,
	}
	h.rootCmd = h.buildRootCommand()
	return h
}

// NewHandlerWithTracker creates a new CLI Handler with tracking support.
// If tracker is nil, tracking is disabled.
func NewHandlerWithTracker(
	runner ports.CommandRunner,
	registry ports.ParserRegistry,
	tracker ports.Tracker,
) *Handler {
	h := &Handler{
		runner:   runner,
		registry: registry,
		tracker:  tracker,
	}
	h.rootCmd = h.buildRootCommand()
	return h
}

// NewHandlerWithSmallFilter creates a new CLI Handler with small filter support.
// The small filter compacts terse output to minimal JSON when enabled.
// If filter is nil, small filtering is disabled.
func NewHandlerWithSmallFilter(
	runner ports.CommandRunner,
	registry ports.ParserRegistry,
	tracker ports.Tracker,
	smallFilter ports.SmallOutputFilter,
) *Handler {
	h := &Handler{
		runner:      runner,
		registry:    registry,
		tracker:     tracker,
		smallFilter: smallFilter,
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
		// Allow arbitrary args so that unknown commands (ls, git, etc.) are passed through
		// instead of being rejected as unknown subcommands
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Handle --help and -h specially since we disable flag parsing
			// Only check if -h or --help appears BEFORE the command name
			// to avoid triggering help when -h is meant for the wrapped command
			for _, arg := range args {
				if arg == "--help" || arg == "-h" {
					return cmd.Help()
				}
				// Stop checking once we hit a non-flag argument (the command)
				if !strings.HasPrefix(arg, "-") {
					break
				}
			}
			envJSON := os.Getenv(EnvJSONKey)
			return h.ExecuteWithArgs(cmd.Context(), args, envJSON, cmd.OutOrStdout())
		},
	}

	// Add stats subcommand if tracker is available
	if h.tracker != nil {
		cmd.AddCommand(buildStatsCommand(h.tracker))
	}

	return cmd
}

// gitLogFormat is the format string for git log to produce parseable output.
const gitLogFormat = "COMMIT_START%n%H%n%h%n%an%n%ae%n%aI%n%s%n%b%nCOMMIT_END"

// gitStatusArgs are the flags needed for parseable git status output.
var gitStatusArgs = []string{"--porcelain=v2", "--branch"}

// transformArgs modifies command arguments to produce parseable output.
// Some commands need special flags to output in a format the parser can handle.
func transformArgs(cmdName string, subcommands []string, args []string) []string {
	switch cmdName {
	case "ls":
		return transformLsArgs(subcommands, args)
	case "git":
		return transformGitArgs(subcommands, args)
	}
	return args
}

// transformLsArgs adds -l flag to ls if not present, so we can detect file types.
func transformLsArgs(subcommands []string, args []string) []string {
	// Check if -l is already present in subcommands or args
	allArgs := append(subcommands, args...)
	for _, arg := range allArgs {
		if arg == "-l" || arg == "-la" || arg == "-al" || arg == "-lh" ||
			strings.Contains(arg, "l") && strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") {
			return args
		}
	}
	// Add -l flag to get file type information
	return append([]string{"-l"}, args...)
}

// transformGitArgs modifies git command arguments for parseable output.
func transformGitArgs(subcommands []string, args []string) []string {
	if len(subcommands) == 0 {
		return args
	}

	switch subcommands[0] {
	case "log":
		// Inject format flag for git log if not already present
		hasFormat := false
		for _, arg := range args {
			if strings.HasPrefix(arg, "--format") || strings.HasPrefix(arg, "--pretty") {
				hasFormat = true
				break
			}
		}
		if !hasFormat {
			args = append([]string{"--format=" + gitLogFormat, "--numstat"}, args...)
		}
	case "status":
		// Inject porcelain format for git status if not already present
		hasPorcelain := false
		for _, arg := range args {
			if strings.HasPrefix(arg, "--porcelain") {
				hasPorcelain = true
				break
			}
		}
		if !hasPorcelain {
			args = append(gitStatusArgs, args...)
		}
	}

	return args
}

// ExecuteWithArgs runs the handler with the given arguments.
// This is the main entry point for command execution.
//
// The method:
// 1. Extracts the --json flag from args
// 2. Determines output mode from flag and env var
// 3. Parses the remaining args into a Command
// 4. Executes the command via the runner (timed)
// 5. Parses output if a parser is registered
// 6. Tracks the execution (if tracker is configured)
// 7. Applies small filter if enabled
// 8. Writes output in the appropriate format
func (h *Handler) ExecuteWithArgs(ctx context.Context, args []string, envJSON string, out io.Writer) error {
	// Get env var for disable filter
	envDisableFilter := os.Getenv(EnvDisableFilterKey)
	return h.ExecuteWithArgsAndEnv(ctx, args, envJSON, envDisableFilter, out)
}

// ExecuteWithArgsAndEnv runs the handler with the given arguments and environment values.
// This variant allows explicit passing of environment values for testing.
func (h *Handler) ExecuteWithArgsAndEnv(
	ctx context.Context,
	args []string,
	envJSON string,
	envDisableFilter string,
	out io.Writer,
) error {
	// Extract --json flag and determine output mode
	jsonFlag, remaining := ExtractJSONFlag(args)
	outputJSON := ShouldOutputJSON(jsonFlag, envJSON)

	// Extract --disable-filter flag
	disableFilters, remaining := ExtractDisableFilter(remaining)

	// Determine if small filter should be disabled
	smallFilterDisabled := ShouldDisableFilter(FilterNameSmall, disableFilters, envDisableFilter)

	// Parse args into Command
	if len(remaining) == 0 {
		return errors.New("no command specified")
	}

	cmd, err := domain.CommandFromArgs(remaining)
	if err != nil {
		return fmt.Errorf("parse command: %w", err)
	}

	// Transform args for parseable output if in JSON mode
	transformedArgs := cmd.Args
	if outputJSON {
		transformedArgs = transformArgs(cmd.Name, cmd.Subcommands, cmd.Args)
	}

	// Start timing the execution
	startTime := time.Now()

	// Execute the command
	cmdArgs := append(cmd.Subcommands, transformedArgs...)
	stdout, stderr, exitCode, err := h.runner.Run(ctx, cmd.Name, cmdArgs)
	if err != nil {
		return h.handleError(out, outputJSON, err, exitCode)
	}

	// Read stdout
	rawBytes, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("read stdout: %w", err)
	}
	raw := string(rawBytes)

	// Read stderr for error information
	stderrBytes, _ := io.ReadAll(stderr)
	stderrStr := string(stderrBytes)

	// Look up parser for this command
	parser, found := h.registry.Find(cmd.Name, cmd.Subcommands)

	result, schema, parseErr := h.parseOutputWithError(parser, found, raw, stderrStr, exitCode)

	// Calculate execution time
	execTime := time.Since(startTime)

	// Track the execution (errors are logged but do not fail command execution)
	h.trackExecution(ctx, cmd, raw, result, execTime, parseErr)

	// Apply small filter if enabled and in JSON mode
	if outputJSON && !smallFilterDisabled && h.smallFilter != nil && parseErr == nil && exitCode == 0 {
		tokenCount := domain.EstimateTokens(raw)
		if h.smallFilter.ShouldFilter(raw, tokenCount, cmd.Name, cmd.Subcommands) {
			smallResult := h.smallFilter.Filter(raw)
			return h.writeSmallFilterResult(out, smallResult)
		}
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

// writeSmallFilterResult writes the small filter result as JSON.
func (h *Handler) writeSmallFilterResult(out io.Writer, result domain.SmallOutputResult) error {
	enc := json.NewEncoder(out)
	return enc.Encode(result)
}

// parseOutputWithError processes command output through a parser or returns passthrough result.
// It also returns the parse error (if any) for tracking purposes.
func (h *Handler) parseOutputWithError(parser ports.Parser, found bool, raw, stderrStr string, exitCode int) (domain.ParseResult, domain.Schema, error) {
	var result domain.ParseResult
	var schema domain.Schema

	if !found {
		// No parser - passthrough mode
		if stderrStr != "" {
			raw = raw + stderrStr
		}
		return domain.NewParseResult(nil, raw, exitCode), schema, nil
	}

	// Check for fatal errors like "not a git repository"
	if isFatalError(exitCode, raw, stderrStr) {
		fatalErr := errors.New(strings.TrimSpace(stderrStr))
		return domain.NewParseResultWithError(
			fatalErr,
			stderrStr,
			exitCode,
		), schema, fatalErr
	}

	// Parse the output (include stderr if stdout is empty)
	parseInput := raw
	if parseInput == "" && stderrStr != "" {
		parseInput = stderrStr
	}

	schema = parser.Schema()
	result, err := parser.Parse(strings.NewReader(parseInput))
	if err != nil {
		wrappedErr := fmt.Errorf("parse error: %w", err)
		return domain.NewParseResultWithError(
			wrappedErr,
			parseInput,
			exitCode,
		), schema, wrappedErr
	}

	result.ExitCode = exitCode
	result.Raw = parseInput
	syncExitCodeAndSuccess(result.Data, exitCode)
	return result, schema, nil
}

// isFatalError checks if the command output indicates a fatal error.
func isFatalError(exitCode int, raw, stderrStr string) bool {
	if exitCode == 0 || raw != "" || stderrStr == "" {
		return false
	}
	return strings.Contains(stderrStr, "not a git repository") ||
		strings.Contains(stderrStr, "command not found")
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
			"raw":      result.Raw,
			"parsed":   false,
			"exitCode": result.ExitCode,
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

// trackExecution records the command execution metrics if tracking is enabled.
// Any tracking errors are ignored to avoid failing command execution.
func (h *Handler) trackExecution(
	ctx context.Context,
	cmd domain.Command,
	raw string,
	result domain.ParseResult,
	execTime time.Duration,
	parseErr error,
) {
	if h.tracker == nil {
		return
	}

	// If there was a parse error, record it as a failure
	if parseErr != nil {
		h.recordParseFailure(ctx, cmd, parseErr)
		return
	}

	// Calculate token metrics
	rawTokens := domain.EstimateTokens(raw)
	parsedJSON := h.serializeResult(result.Data)
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
	_ = h.tracker.Record(ctx, record)
}

// recordParseFailure records a parse failure for tracking.
func (h *Handler) recordParseFailure(ctx context.Context, cmd domain.Command, parseErr error) {
	// Build full command string
	cmdParts := append([]string{cmd.Name}, cmd.Subcommands...)
	fullCmd := strings.Join(cmdParts, " ")

	failure := domain.NewParseFailure(fullCmd, parseErr.Error(), true)
	_ = h.tracker.RecordFailure(ctx, failure)
}

// serializeResult converts result data to JSON for token counting.
func (h *Handler) serializeResult(data any) string {
	if data == nil {
		return ""
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}
