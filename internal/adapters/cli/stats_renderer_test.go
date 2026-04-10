// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// mockStatsRenderer implements ports.StatsRenderer for testing.
type mockStatsRenderer struct {
	renderSummaryCalled bool
	lastSummary         domain.StatsSummary
	renderErr           error
}

func (m *mockStatsRenderer) RenderSummary(w io.Writer, summary domain.StatsSummary) error {
	m.renderSummaryCalled = true
	m.lastSummary = summary
	if m.renderErr != nil {
		return m.renderErr
	}
	_, err := w.Write([]byte("Stats: rendered"))
	return err
}

func (m *mockStatsRenderer) RenderHistory(_ io.Writer, _ []domain.CommandRecord) error {
	return nil
}

func (m *mockStatsRenderer) RenderByParser(_ io.Writer, _ []domain.CommandStats) error {
	return nil
}

func (m *mockStatsRenderer) RenderByFilter(_ io.Writer, _ []domain.FilterStats) error {
	return nil
}

// mockThemeProvider implements ports.ThemeProvider for testing.
type mockThemeProvider struct {
	name string
}

func (m *mockThemeProvider) ColorFor(_ domain.SavingsCategory) string {
	return "\033[32m"
}

func (m *mockThemeProvider) Name() string {
	return m.name
}

func (m *mockThemeProvider) ListThemes() []string {
	return []string{m.name}
}

func (m *mockThemeProvider) SetTheme(_ string) error {
	return nil
}

func TestHandler_StatsFlag_ShowsEnhancedOutput(t *testing.T) {
	// Arrange: handler with mock tracker returning sample stats, --stats flag provided
	runner := &mockRunner{stdout: "On branch main\nnothing to commit"}
	registry := &mockRegistry{}
	tracker := &mockTracker{}
	renderer := &mockStatsRenderer{}
	themeProvider := &mockThemeProvider{name: "default"}

	h := NewHandlerWithStatsRenderer(runner, registry, tracker, nil, nil, nil, renderer, themeProvider)

	// Act: run command with --stats flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "--stats", "status"}, "", &buf)

	// Assert: stats renderer should be called
	require.NoError(t, err)
	assert.True(t, renderer.renderSummaryCalled, "StatsRenderer.RenderSummary should be called when --stats flag is set")
	output := buf.String()
	assert.Contains(t, output, "Stats: rendered", "output should contain rendered stats")
}

func TestHandler_NoStatsFlag_NoStatsOutput(t *testing.T) {
	// Arrange: handler with mock tracker, no --stats flag
	runner := &mockRunner{stdout: "On branch main\nnothing to commit"}
	registry := &mockRegistry{}
	tracker := &mockTracker{}
	renderer := &mockStatsRenderer{}
	themeProvider := &mockThemeProvider{name: "default"}

	h := NewHandlerWithStatsRenderer(runner, registry, tracker, nil, nil, nil, renderer, themeProvider)

	// Act: run command WITHOUT --stats flag
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "status"}, "", &buf)

	// Assert: stats renderer should NOT be called
	require.NoError(t, err)
	assert.False(t, renderer.renderSummaryCalled, "StatsRenderer.RenderSummary should NOT be called without --stats flag")
	output := buf.String()
	assert.NotContains(t, output, "Stats: rendered", "output should NOT contain stats")
}

func TestHandler_StatsFlag_WithTheme(t *testing.T) {
	// Arrange: handler with theme provider and theme resolver
	runner := &mockRunner{stdout: "On branch main\nnothing to commit"}
	registry := &mockRegistry{}
	tracker := &mockTracker{}
	renderer := &mockStatsRenderer{}
	defaultTheme := &mockThemeProvider{name: "default"}
	darkTheme := &mockThemeProvider{name: "dark"}

	var resolvedName string
	resolver := func(name string) ports.ThemeProvider {
		resolvedName = name
		return darkTheme
	}

	h := NewHandlerWithStatsRenderer(runner, registry, tracker, nil, nil, nil, renderer, defaultTheme)
	h.SetThemeResolver(resolver)

	// Act: run command with --stats and --theme flags
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "--stats", "--theme=dark", "status"}, "", &buf)

	// Assert: theme resolver should be called with "dark" and stats should render
	require.NoError(t, err)
	assert.True(t, renderer.renderSummaryCalled, "StatsRenderer.RenderSummary should be called")
	assert.Equal(t, "dark", resolvedName, "theme resolver should be called with the --theme flag value")
}

func TestHandler_ThemeFlag_NoResolver_UsesDefault(t *testing.T) {
	// Arrange: handler with theme provider but no resolver
	runner := &mockRunner{stdout: "On branch main\nnothing to commit"}
	registry := &mockRegistry{}
	tracker := &mockTracker{}
	renderer := &mockStatsRenderer{}
	defaultTheme := &mockThemeProvider{name: "default"}

	h := NewHandlerWithStatsRenderer(runner, registry, tracker, nil, nil, nil, renderer, defaultTheme)
	// No resolver set - should fall back to injected themeProvider

	// Act: run command with --stats and --theme flags
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "--stats", "--theme=dark", "status"}, "", &buf)

	// Assert: should succeed using default theme without panic
	require.NoError(t, err)
	assert.True(t, renderer.renderSummaryCalled, "StatsRenderer.RenderSummary should still be called")
}

func TestHandler_StatsFlag_NilRenderer_NoOutput(t *testing.T) {
	// Arrange: handler without stats renderer
	runner := &mockRunner{stdout: "On branch main\nnothing to commit"}
	registry := &mockRegistry{}
	tracker := &mockTracker{}

	h := NewHandlerWithStatsRenderer(runner, registry, tracker, nil, nil, nil, nil, nil)

	// Act: run command with --stats flag but no renderer
	var buf bytes.Buffer
	err := h.ExecuteWithArgs(context.Background(), []string{"git", "--stats", "status"}, "", &buf)

	// Assert: should succeed without panic
	require.NoError(t, err)
}
