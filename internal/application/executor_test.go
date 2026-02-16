package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// mockRunner implements ports.CommandRunner for testing.
type mockRunner struct {
	stdout   string
	stderr   string
	exitCode int
	err      error
}

func (m *mockRunner) Run(_ context.Context, _ string, _ []string) (io.Reader, io.Reader, int, error) {
	if m.err != nil {
		return nil, nil, 0, m.err
	}
	return strings.NewReader(m.stdout), strings.NewReader(m.stderr), m.exitCode, nil
}

// mockParser implements ports.Parser for testing.
type mockParser struct {
	result      domain.ParseResult
	parseErr    error
	schema      domain.Schema
	matches     bool
	parseCalled bool
}

func (m *mockParser) Parse(_ io.Reader) (domain.ParseResult, error) {
	m.parseCalled = true
	if m.parseErr != nil {
		return domain.ParseResult{}, m.parseErr
	}
	return m.result, nil
}

func (m *mockParser) Schema() domain.Schema {
	return m.schema
}

func (m *mockParser) Matches(_ string, _ []string) bool {
	return m.matches
}

// mockRegistry implements ports.ParserRegistry for testing.
type mockRegistry struct {
	parser ports.Parser
	found  bool
}

func (m *mockRegistry) Find(_ string, _ []string) (ports.Parser, bool) {
	return m.parser, m.found
}

func (m *mockRegistry) Register(_ ports.Parser) {}

func (m *mockRegistry) All() []ports.Parser {
	if m.parser != nil {
		return []ports.Parser{m.parser}
	}
	return nil
}

// mockWriter implements ports.OutputWriter for testing.
type mockWriter struct {
	writeCalled bool
	result      domain.ParseResult
	schema      domain.Schema
	writeErr    error
}

func (m *mockWriter) Write(_ io.Writer, result domain.ParseResult, schema domain.Schema) error {
	m.writeCalled = true
	m.result = result
	m.schema = schema
	return m.writeErr
}

func TestExecutor_Execute_WithParser(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cmd := domain.NewCommand("git", []string{"status"}, nil)

	expectedData := map[string]any{"branch": "main", "clean": true}
	rawOutput := "On branch main\nnothing to commit"
	expectedSchema := domain.NewSchema("git-status", "Git Status", "object", nil, nil)

	parser := &mockParser{
		result:  domain.NewParseResult(expectedData, rawOutput, 0),
		schema:  expectedSchema,
		matches: true,
	}

	runner := &mockRunner{
		stdout:   rawOutput,
		exitCode: 0,
	}

	registry := &mockRegistry{
		parser: parser,
		found:  true,
	}

	writer := &mockWriter{}

	executor := NewExecutor(runner, registry, writer)

	// Act
	var buf bytes.Buffer
	err := executor.Execute(ctx, cmd, &buf)

	// Assert
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if !parser.parseCalled {
		t.Error("Execute() did not call parser.Parse()")
	}

	if !writer.writeCalled {
		t.Error("Execute() did not call writer.Write()")
	}

	if writer.result.Data == nil {
		t.Error("Execute() writer.result.Data is nil, want data")
	}
}

func TestExecutor_Execute_NoParser(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cmd := domain.NewCommand("unknown", []string{"cmd"}, nil)

	rawOutput := "some raw output"

	runner := &mockRunner{
		stdout:   rawOutput,
		exitCode: 0,
	}

	registry := &mockRegistry{
		parser: nil,
		found:  false,
	}

	writer := &mockWriter{}

	executor := NewExecutor(runner, registry, writer)

	// Act
	var buf bytes.Buffer
	err := executor.Execute(ctx, cmd, &buf)

	// Assert
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if !writer.writeCalled {
		t.Error("Execute() did not call writer.Write()")
	}

	// Should pass through raw output when no parser found
	if writer.result.Raw != rawOutput {
		t.Errorf("Execute() writer.result.Raw = %q, want %q", writer.result.Raw, rawOutput)
	}
}

func TestExecutor_Execute_RunnerError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cmd := domain.NewCommand("git", []string{"status"}, nil)

	runnerErr := errors.New("command not found")

	runner := &mockRunner{
		err: runnerErr,
	}

	registry := &mockRegistry{
		found: false,
	}

	writer := &mockWriter{}

	executor := NewExecutor(runner, registry, writer)

	// Act
	var buf bytes.Buffer
	err := executor.Execute(ctx, cmd, &buf)

	// Assert
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	if !errors.Is(err, runnerErr) {
		t.Errorf("Execute() error = %v, want %v", err, runnerErr)
	}
}

func TestExecutor_Execute_ParserError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cmd := domain.NewCommand("git", []string{"status"}, nil)

	rawOutput := "unexpected format output"
	parseErr := errors.New("parse failed")

	parser := &mockParser{
		parseErr: parseErr,
		matches:  true,
	}

	runner := &mockRunner{
		stdout:   rawOutput,
		exitCode: 0,
	}

	registry := &mockRegistry{
		parser: parser,
		found:  true,
	}

	writer := &mockWriter{}

	executor := NewExecutor(runner, registry, writer)

	// Act
	var buf bytes.Buffer
	err := executor.Execute(ctx, cmd, &buf)

	// Assert
	// When parser fails, we should still write a result with the error and raw output
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil (error should be in result)", err)
	}

	if !writer.writeCalled {
		t.Error("Execute() did not call writer.Write()")
	}

	// The result should contain the raw output even when parser fails
	if writer.result.Raw != rawOutput {
		t.Errorf("Execute() writer.result.Raw = %q, want %q", writer.result.Raw, rawOutput)
	}

	// The result should contain the error
	if writer.result.Error == nil {
		t.Error("Execute() writer.result.Error = nil, want parse error")
	}
}

func TestExecutor_UnsupportedCommand_JSON(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cmd := domain.NewCommand("git", []string{"stash", "show"}, nil)

	rawOutput := "stash@{0}: WIP on main: abc123 Some commit message"

	runner := &mockRunner{
		stdout:   rawOutput,
		exitCode: 0,
	}

	registry := &mockRegistry{
		parser: nil,
		found:  false, // No parser for "git stash show"
	}

	writer := &mockWriter{}

	executor := NewExecutor(runner, registry, writer)

	// Act
	var buf bytes.Buffer
	err := executor.Execute(ctx, cmd, &buf)

	// Assert
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if !writer.writeCalled {
		t.Error("Execute() did not call writer.Write()")
	}

	// When no parser is found, result.Data should be a FallbackResult
	fallbackResult, ok := writer.result.Data.(domain.FallbackResult)
	if !ok {
		t.Fatalf("Execute() writer.result.Data is %T, want domain.FallbackResult", writer.result.Data)
	}

	// FallbackResult should contain raw output
	if fallbackResult.Raw != rawOutput {
		t.Errorf("FallbackResult.Raw = %q, want %q", fallbackResult.Raw, rawOutput)
	}

	// FallbackResult.Parsed should be false
	if fallbackResult.Parsed != false {
		t.Errorf("FallbackResult.Parsed = %v, want false", fallbackResult.Parsed)
	}

	// FallbackResult should preserve exit code
	if fallbackResult.ExitCode != 0 {
		t.Errorf("FallbackResult.ExitCode = %d, want 0", fallbackResult.ExitCode)
	}
}

func TestExecutor_UnsupportedCommand_NonZeroExit(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cmd := domain.NewCommand("git", []string{"stash", "show"}, nil)

	rawOutput := "No stash entries found."
	exitCode := 1

	runner := &mockRunner{
		stdout:   rawOutput,
		exitCode: exitCode,
	}

	registry := &mockRegistry{
		parser: nil,
		found:  false,
	}

	writer := &mockWriter{}

	executor := NewExecutor(runner, registry, writer)

	// Act
	var buf bytes.Buffer
	err := executor.Execute(ctx, cmd, &buf)

	// Assert
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	fallbackResult, ok := writer.result.Data.(domain.FallbackResult)
	if !ok {
		t.Fatalf("Execute() writer.result.Data is %T, want domain.FallbackResult", writer.result.Data)
	}

	// Verify exit code is preserved in fallback result
	if fallbackResult.ExitCode != exitCode {
		t.Errorf("FallbackResult.ExitCode = %d, want %d", fallbackResult.ExitCode, exitCode)
	}
}
