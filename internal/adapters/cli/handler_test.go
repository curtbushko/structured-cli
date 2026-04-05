package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// mockTracker implements ports.Tracker for testing.
type mockTracker struct {
	recordCalled        bool
	recordFailureCalled bool
	lastRecord          domain.CommandRecord
	lastFailure         domain.ParseFailure
}

// mockRunner implements ports.CommandRunner for testing.
type mockRunner struct {
	stdout   string
	stderr   string
	exitCode int
	runErr   error
	called   bool
	lastCmd  string
	lastArgs []string
}

func (m *mockRunner) Run(_ context.Context, cmd string, args []string) (io.Reader, io.Reader, int, error) {
	m.called = true
	m.lastCmd = cmd
	m.lastArgs = args
	return strings.NewReader(m.stdout), strings.NewReader(m.stderr), m.exitCode, m.runErr
}

// mockParser implements ports.Parser for testing.
type mockParser struct {
	schema   domain.Schema
	result   domain.ParseResult
	parseErr error
	matchCmd string
	matchSub []string
}

func (m *mockParser) Parse(_ io.Reader) (domain.ParseResult, error) {
	return m.result, m.parseErr
}

func (m *mockParser) Schema() domain.Schema {
	return m.schema
}

func (m *mockParser) Matches(cmd string, subcommands []string) bool {
	if m.matchCmd != cmd {
		return false
	}
	if len(m.matchSub) != len(subcommands) {
		return false
	}
	for i, s := range m.matchSub {
		if s != subcommands[i] {
			return false
		}
	}
	return true
}

// mockRegistry implements ports.ParserRegistry for testing.
type mockRegistry struct {
	parsers []mockParser
}

func (m *mockRegistry) Find(cmd string, subcommands []string) (ports.Parser, bool) {
	for i := range m.parsers {
		if m.parsers[i].Matches(cmd, subcommands) {
			return &m.parsers[i], true
		}
	}
	return nil, false
}

func (m *mockRegistry) Register(_ ports.Parser) {}

func (m *mockRegistry) All() []ports.Parser {
	result := make([]ports.Parser, len(m.parsers))
	for i := range m.parsers {
		result[i] = &m.parsers[i]
	}
	return result
}

func TestHandler_DetermineOutputMode(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		envJSON  string
		wantJSON bool
	}{
		{
			name:     "default passthrough mode",
			args:     []string{"git", "status"},
			envJSON:  "",
			wantJSON: false,
		},
		{
			name:     "JSON via --json flag",
			args:     []string{"git", "--json", "status"},
			envJSON:  "",
			wantJSON: true,
		},
		{
			name:     "JSON via env var",
			args:     []string{"git", "status"},
			envJSON:  "true",
			wantJSON: true,
		},
		{
			name:     "flag overrides env false",
			args:     []string{"git", "--json", "status"},
			envJSON:  "false",
			wantJSON: true,
		},
		{
			name:     "passthrough with env false",
			args:     []string{"git", "status"},
			envJSON:  "false",
			wantJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(nil, nil)
			jsonFlag, remaining := ExtractJSONFlag(tt.args)
			gotJSON := ShouldOutputJSON(jsonFlag, tt.envJSON)

			if gotJSON != tt.wantJSON {
				t.Errorf("output mode = %v, want JSON = %v", gotJSON, tt.wantJSON)
			}

			// Verify remaining args don't contain --json
			for _, arg := range remaining {
				if arg == "--json" {
					t.Errorf("remaining args still contain --json: %v", remaining)
				}
			}

			// Suppress unused variable warning
			_ = h
		})
	}
}

func TestHandler_NewHandler(t *testing.T) {
	runner := &mockRunner{}
	registry := &mockRegistry{}

	h := NewHandler(runner, registry)

	if h == nil {
		t.Fatal("NewHandler returned nil")
	}

	if h.Runner() != runner {
		t.Error("Runner not set correctly")
	}

	if h.Registry() != registry {
		t.Error("Registry not set correctly")
	}
}

func TestHandler_RootCommand(t *testing.T) {
	h := NewHandler(nil, nil)

	cmd := h.RootCommand()
	if cmd == nil {
		t.Fatal("RootCommand returned nil")
	}

	// Verify basic command properties
	if cmd.Use == "" {
		t.Error("Root command Use should not be empty")
	}
}

func TestHandler_ExecuteWithArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		envJSON    string
		stdout     string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "passthrough mode outputs raw",
			args:       []string{"git", "status"},
			envJSON:    "",
			stdout:     "On branch main\n",
			wantOutput: "On branch main\n",
			wantErr:    false,
		},
		{
			name:       "json mode with --json flag",
			args:       []string{"--json", "git", "status"},
			envJSON:    "",
			stdout:     "On branch main\n",
			wantOutput: `{"raw":"On branch main\n","parsed":false}`,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRunner{stdout: tt.stdout}
			registry := &mockRegistry{}

			h := NewHandler(runner, registry)

			var buf bytes.Buffer
			err := h.ExecuteWithArgs(context.Background(), tt.args, tt.envJSON, &buf)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWithArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify runner was called with correct args (without --json)
			if runner.called {
				for _, arg := range runner.lastArgs {
					if arg == "--json" {
						t.Error("--json flag passed to underlying command")
					}
				}
			}
		})
	}
}

func TestTransformLsArgs(t *testing.T) {
	tests := []struct {
		name        string
		subcommands []string
		args        []string
		wantAdded   bool // whether -l should be added to result
	}{
		{
			name:        "plain ls adds -l",
			subcommands: []string{},
			args:        []string{},
			wantAdded:   true,
		},
		{
			name:        "ls with path adds -l",
			subcommands: []string{},
			args:        []string{"/tmp"},
			wantAdded:   true,
		},
		{
			name:        "ls -a adds -l",
			subcommands: []string{"-a"},
			args:        []string{},
			wantAdded:   true,
		},
		{
			name:        "ls -l already present - no add",
			subcommands: []string{"-l"},
			args:        []string{},
			wantAdded:   false,
		},
		{
			name:        "ls -la already has l - no add",
			subcommands: []string{"-la"},
			args:        []string{},
			wantAdded:   false,
		},
		{
			name:        "ls -al already has l - no add",
			subcommands: []string{"-al"},
			args:        []string{},
			wantAdded:   false,
		},
		{
			name:        "ls -lh already has l - no add",
			subcommands: []string{"-lh"},
			args:        []string{},
			wantAdded:   false,
		},
		{
			name:        "ls with -l in args - no add",
			subcommands: []string{},
			args:        []string{"-l", "/tmp"},
			wantAdded:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformLsArgs(tt.subcommands, tt.args)

			// Check if -l was added (result should have -l as first element if added)
			added := len(result) > 0 && result[0] == "-l" && (len(tt.args) == 0 || tt.args[0] != "-l")

			if added != tt.wantAdded {
				t.Errorf("transformLsArgs() added -l = %v, want %v, result = %v", added, tt.wantAdded, result)
			}
		})
	}
}

func TestTransformArgs_Ls(t *testing.T) {
	tests := []struct {
		name        string
		cmdName     string
		subcommands []string
		args        []string
		wantFirst   string
	}{
		{
			name:        "ls without -l gets -l added",
			cmdName:     "ls",
			subcommands: []string{},
			args:        []string{},
			wantFirst:   "-l",
		},
		{
			name:        "ls with path gets -l prepended",
			cmdName:     "ls",
			subcommands: []string{},
			args:        []string{"/tmp"},
			wantFirst:   "-l",
		},
		{
			name:        "non-ls command unchanged",
			cmdName:     "cat",
			subcommands: []string{},
			args:        []string{"file.txt"},
			wantFirst:   "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformArgs(tt.cmdName, tt.subcommands, tt.args)

			if len(result) == 0 {
				if tt.wantFirst != "" {
					t.Errorf("transformArgs() returned empty, want first = %q", tt.wantFirst)
				}
				return
			}

			if result[0] != tt.wantFirst {
				t.Errorf("transformArgs() first = %q, want %q", result[0], tt.wantFirst)
			}
		})
	}
}

func (m *mockTracker) Record(_ context.Context, record domain.CommandRecord) error {
	m.recordCalled = true
	m.lastRecord = record
	return nil
}

func (m *mockTracker) RecordFailure(_ context.Context, failure domain.ParseFailure) error {
	m.recordFailureCalled = true
	m.lastFailure = failure
	return nil
}

func (m *mockTracker) Stats(_ context.Context, _ ports.StatsOptions) (domain.StatsSummary, error) {
	return domain.StatsSummary{}, nil
}

func (m *mockTracker) History(_ context.Context, _ int) ([]domain.CommandRecord, error) {
	return nil, nil
}

func (m *mockTracker) StatsByParser(_ context.Context) ([]domain.CommandStats, error) {
	return nil, nil
}

func (m *mockTracker) StatsByFilter(_ context.Context) ([]domain.FilterStats, error) {
	return nil, nil
}

func (m *mockTracker) Cleanup(_ context.Context, _ time.Duration) error {
	return nil
}

func (m *mockTracker) Close() error {
	return nil
}

func TestHandler_TracksExecution(t *testing.T) {
	// Arrange
	runner := &mockRunner{stdout: "On branch main\nnothing to commit"}
	registry := &mockRegistry{}
	tracker := &mockTracker{}

	h := NewHandlerWithTracker(runner, registry, tracker)

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status"}, "", &buf)

	// Assert
	require.NoError(t, err)
	assert.True(t, tracker.recordCalled, "Tracker.Record should be called")
	assert.Equal(t, "git", tracker.lastRecord.Command)
	assert.Equal(t, []string{"status"}, tracker.lastRecord.Subcommands)
}

func TestHandler_SkipsTrackingWhenNilTracker(t *testing.T) {
	// Arrange - handler with nil tracker (tracking disabled)
	runner := &mockRunner{stdout: "On branch main"}
	registry := &mockRegistry{}

	h := NewHandler(runner, registry) // nil tracker

	// Act
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status"}, "", &buf)

	// Assert - should succeed without panic
	require.NoError(t, err)
}
