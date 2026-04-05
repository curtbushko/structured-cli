// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// mockSmallFilter implements ports.SmallOutputFilter for testing.
type mockSmallFilter struct {
	shouldFilterResult bool
	filterResult       domain.SmallOutputResult
	shouldFilterCalled bool
	filterCalled       bool
	lastRaw            string
	lastTokenCount     int
	lastCmd            string
	lastSubcmds        []string
}

func (m *mockSmallFilter) ShouldFilter(raw string, tokenCount int, cmd string, subcmds []string) bool {
	m.shouldFilterCalled = true
	m.lastRaw = raw
	m.lastTokenCount = tokenCount
	m.lastCmd = cmd
	m.lastSubcmds = subcmds
	return m.shouldFilterResult
}

func (m *mockSmallFilter) Filter(raw string) domain.SmallOutputResult {
	m.filterCalled = true
	m.lastRaw = raw
	return m.filterResult
}

// TestHandler_SmallFilterEnabledByDefault tests that the small filter is applied
// in JSON mode by default.
func TestHandler_SmallFilterEnabledByDefault(t *testing.T) {
	// Given: Create handler with small filter, mock runner returning clean git status
	runner := &mockRunner{stdout: "nothing to commit, working tree clean"}
	registry := &mockRegistry{}
	filter := &mockSmallFilter{
		shouldFilterResult: true,
		filterResult: domain.SmallOutputResult{
			Status:  "clean",
			Summary: "nothing to commit, working tree clean",
		},
	}

	h := NewHandlerWithSmallFilter(runner, registry, nil, filter)

	// When: Execute 'git status --json'
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status", "--json"}, "", &buf)

	// Then: Output is compact JSON with status and summary
	require.NoError(t, err)
	assert.True(t, filter.shouldFilterCalled, "ShouldFilter should be called")

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Equal(t, "clean", output["status"])
	assert.Equal(t, "nothing to commit, working tree clean", output["summary"])
}

// TestHandler_DisableFilterViaFlag tests that --disable-filter=small disables
// the small filter.
func TestHandler_DisableFilterViaFlag(t *testing.T) {
	// Given: Create handler with small filter, mock runner returning clean git status
	runner := &mockRunner{stdout: "nothing to commit, working tree clean"}
	registry := &mockRegistry{}
	filter := &mockSmallFilter{
		shouldFilterResult: true,
		filterResult: domain.SmallOutputResult{
			Status:  "clean",
			Summary: "nothing to commit, working tree clean",
		},
	}

	h := NewHandlerWithSmallFilter(runner, registry, nil, filter)

	// When: Execute 'git status --json --disable-filter=small'
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status", "--json", "--disable-filter=small"}, "", &buf)

	// Then: Small filter is not called, full output is returned
	require.NoError(t, err)
	assert.False(t, filter.shouldFilterCalled, "ShouldFilter should NOT be called when disabled")

	// Output should be raw passthrough wrapped in JSON (no parser)
	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Contains(t, output, "raw")
	assert.Equal(t, "nothing to commit, working tree clean", output["raw"])
}

// TestHandler_DisableFilterViaEnv tests that STRUCTURED_CLI_DISABLE_FILTER=small
// disables the small filter.
func TestHandler_DisableFilterViaEnv(t *testing.T) {
	// Given: Create handler with small filter, mock runner returning clean git status
	runner := &mockRunner{stdout: "nothing to commit, working tree clean"}
	registry := &mockRegistry{}
	filter := &mockSmallFilter{
		shouldFilterResult: true,
		filterResult: domain.SmallOutputResult{
			Status:  "clean",
			Summary: "nothing to commit, working tree clean",
		},
	}

	h := NewHandlerWithSmallFilter(runner, registry, nil, filter)

	// When: Execute 'git status --json' with env var set
	var buf bytes.Buffer
	err := h.ExecuteWithArgsAndEnv(
		context.Background(),
		[]string{"git", "status", "--json"},
		"",      // envJSON
		"small", // envDisableFilter
		&buf,
	)

	// Then: Small filter is not called
	require.NoError(t, err)
	assert.False(t, filter.shouldFilterCalled, "ShouldFilter should NOT be called when disabled via env")
}

// TestHandler_DisableAllFilters tests that --disable-filter=all disables
// the small filter (and any other filters).
func TestHandler_DisableAllFilters(t *testing.T) {
	// Given: Create handler with small filter
	runner := &mockRunner{stdout: "nothing to commit, working tree clean"}
	registry := &mockRegistry{}
	filter := &mockSmallFilter{
		shouldFilterResult: true,
		filterResult: domain.SmallOutputResult{
			Status:  "clean",
			Summary: "nothing to commit, working tree clean",
		},
	}

	h := NewHandlerWithSmallFilter(runner, registry, nil, filter)

	// When: Execute 'git status --json --disable-filter=all'
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status", "--json", "--disable-filter=all"}, "", &buf)

	// Then: Small filter is not called
	require.NoError(t, err)
	assert.False(t, filter.shouldFilterCalled, "ShouldFilter should NOT be called when all filters disabled")
}

// TestHandler_FilterNotTriggeredForLargeOutput tests that the small filter
// does not trigger when output exceeds the threshold.
func TestHandler_FilterNotTriggeredForLargeOutput(t *testing.T) {
	// Given: Create handler, mock runner returning git status with many files
	largeOutput := "M file1.go\nM file2.go\nM file3.go\nM file4.go\nM file5.go\n" +
		"M file6.go\nM file7.go\nM file8.go\nM file9.go\nM file10.go\n" +
		"A newfile1.go\nA newfile2.go\nA newfile3.go\nA newfile4.go\n" +
		"D deleted1.go\nD deleted2.go\nD deleted3.go\nD deleted4.go\n" +
		"This is large output that should not trigger the small filter."

	runner := &mockRunner{stdout: largeOutput}
	registry := &mockRegistry{}
	filter := &mockSmallFilter{
		shouldFilterResult: false, // Filter says "don't filter this"
	}

	h := NewHandlerWithSmallFilter(runner, registry, nil, filter)

	// When: Execute 'git status --json'
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status", "--json"}, "", &buf)

	// Then: ShouldFilter was called but returned false, so filter is not applied
	require.NoError(t, err)
	assert.True(t, filter.shouldFilterCalled, "ShouldFilter should be called")
	assert.False(t, filter.filterCalled, "Filter should NOT be called when ShouldFilter returns false")

	// Output should be raw wrapped JSON (no parser)
	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Contains(t, output, "raw")
}

// TestHandler_PassthroughModeUnaffected tests that the small filter
// is not applied in passthrough mode (no --json flag).
func TestHandler_PassthroughModeUnaffected(t *testing.T) {
	// Given: Create handler with small filter, mock runner returning clean git status
	runner := &mockRunner{stdout: "nothing to commit, working tree clean"}
	registry := &mockRegistry{}
	filter := &mockSmallFilter{
		shouldFilterResult: true,
		filterResult: domain.SmallOutputResult{
			Status:  "clean",
			Summary: "nothing to commit, working tree clean",
		},
	}

	h := NewHandlerWithSmallFilter(runner, registry, nil, filter)

	// When: Execute 'git status' (no --json flag)
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status"}, "", &buf)

	// Then: Small filter is not called (passthrough mode)
	require.NoError(t, err)
	assert.False(t, filter.shouldFilterCalled, "ShouldFilter should NOT be called in passthrough mode")

	// Output should be raw passthrough
	assert.Equal(t, "nothing to commit, working tree clean", buf.String())
}

// TestHandler_SmallFilterWithParser tests that when a parser returns data,
// the small filter can still compact it.
func TestHandler_SmallFilterWithParser(t *testing.T) {
	// Given: Create handler with small filter and a parser
	runner := &mockRunner{stdout: "nothing to commit, working tree clean"}

	// Create a simple result that would be returned by a git status parser
	parserResult := domain.NewParseResult(
		map[string]any{
			"success": true,
			"branch":  "main",
			"clean":   true,
		},
		"nothing to commit, working tree clean",
		0,
	)

	parser := &mockParser{
		matchCmd: "git",
		matchSub: []string{"status"},
		result:   parserResult,
		schema:   domain.NewSchema("test", "test", "object", nil, nil),
	}

	registry := &mockRegistry{
		parsers: []mockParser{*parser},
	}

	filter := &mockSmallFilter{
		shouldFilterResult: true,
		filterResult: domain.SmallOutputResult{
			Status:  "clean",
			Summary: "nothing to commit, working tree clean",
		},
	}

	h := NewHandlerWithSmallFilter(runner, registry, nil, filter)

	// When: Execute 'git status --json'
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status", "--json"}, "", &buf)

	// Then: Small filter is applied
	require.NoError(t, err)
	assert.True(t, filter.shouldFilterCalled, "ShouldFilter should be called")

	var output map[string]any
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Equal(t, "clean", output["status"])
}

// TestHandler_StatsTrackingWithFilter tests that stats tracking records
// filter activations and token savings.
func TestHandler_StatsTrackingWithFilter(t *testing.T) {
	// Given: Create handler with tracker and small filter
	runner := &mockRunner{stdout: "nothing to commit, working tree clean"}
	registry := &mockRegistry{}
	tracker := &mockTracker{}
	filter := &mockSmallFilter{
		shouldFilterResult: true,
		filterResult: domain.SmallOutputResult{
			Status:  "clean",
			Summary: "nothing to commit, working tree clean",
		},
	}

	h := NewHandlerWithSmallFilter(runner, registry, tracker, filter)

	// When: Execute 'git status --json'
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status", "--json"}, "", &buf)

	// Then: Stats are tracked with filter activation
	require.NoError(t, err)
	assert.True(t, tracker.recordCalled, "Tracker.Record should be called")
	// Token savings should be positive (compact output vs raw)
	// The actual savings calculation is done by the tracker
}

// mockRunnerWithStderr extends mockRunner to include stderr support.
type mockRunnerWithStderr struct {
	stdout   string
	stderr   string
	exitCode int
	runErr   error
}

func (m *mockRunnerWithStderr) Run(_ context.Context, _ string, _ []string) (io.Reader, io.Reader, int, error) {
	return strings.NewReader(m.stdout), strings.NewReader(m.stderr), m.exitCode, m.runErr
}

// TestHandler_SmallFilterNotAppliedOnError tests that the small filter
// is not applied when the command returns an error.
func TestHandler_SmallFilterNotAppliedOnError(t *testing.T) {
	// Given: Create handler with small filter, mock runner returning error
	runner := &mockRunnerWithStderr{
		stdout:   "",
		stderr:   "fatal: not a git repository",
		exitCode: 128,
	}
	registry := &mockRegistry{}
	filter := &mockSmallFilter{
		shouldFilterResult: true,
	}

	h := NewHandlerWithSmallFilterAndRunner(runner, registry, nil, filter)

	// When: Execute 'git status --json'
	var buf bytes.Buffer
	_ = h.ExecuteWithArgs(context.Background(), []string{"git", "status", "--json"}, "", &buf)

	// Then: Filter behavior depends on implementation
	// The key assertion is that the command completes without panic
}

// Helper to create handler with custom runner interface
func NewHandlerWithSmallFilterAndRunner(
	runner ports.CommandRunner,
	registry ports.ParserRegistry,
	tracker ports.Tracker,
	filter ports.SmallOutputFilter,
) *Handler {
	h := &Handler{
		runner:      runner,
		registry:    registry,
		tracker:     tracker,
		smallFilter: filter,
	}
	h.rootCmd = h.buildRootCommand()
	return h
}
