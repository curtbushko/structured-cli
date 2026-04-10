package cli

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// mockThemeLister implements ports.ThemeProvider for theme command testing.
type mockThemeLister struct {
	name      string
	themes    []string
	setErr    error
	setCalled string
}

func (m *mockThemeLister) ColorFor(_ domain.SavingsCategory) string {
	return ""
}

func (m *mockThemeLister) Name() string {
	return m.name
}

func (m *mockThemeLister) ListThemes() []string {
	return m.themes
}

func (m *mockThemeLister) SetTheme(name string) error {
	m.setCalled = name
	return m.setErr
}

func TestThemeListCommand_OutputsThemeNames(t *testing.T) {
	// Given: ThemeProvider with multiple themes
	provider := &mockThemeLister{
		name:   "default",
		themes: []string{"default", "dark", "light"},
	}

	// When: run theme list
	var buf bytes.Buffer
	err := executeThemeListCommand(provider, false, &buf)

	// Then: output contains all theme names
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "default")
	assert.Contains(t, output, "dark")
	assert.Contains(t, output, "light")
}

func TestThemeListCommand_JSON(t *testing.T) {
	// Given: ThemeProvider with multiple themes and --json flag
	provider := &mockThemeLister{
		name:   "default",
		themes: []string{"default", "dark", "light"},
	}

	// When: run theme list --json
	var buf bytes.Buffer
	err := executeThemeListCommand(provider, true, &buf)

	// Then: output is valid JSON array of theme name strings
	require.NoError(t, err)
	var names []string
	err = json.Unmarshal(buf.Bytes(), &names)
	require.NoError(t, err)
	assert.Equal(t, []string{"default", "dark", "light"}, names)
}

func TestThemeSelectCommand_ValidTheme(t *testing.T) {
	// Given: ThemeProvider that includes 'dark'
	provider := &mockThemeLister{
		name:   "default",
		themes: []string{"default", "dark", "light"},
	}

	// When: run theme select dark
	var buf bytes.Buffer
	err := executeThemeSelectCommand(provider, "dark", &buf)

	// Then: exits successfully, success message in output
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "dark")
	assert.Equal(t, "dark", provider.setCalled)
}

func TestThemeSelectCommand_InvalidTheme(t *testing.T) {
	// Given: ThemeProvider that does NOT include 'unknown'
	provider := &mockThemeLister{
		name:   "default",
		themes: []string{"default", "dark", "light"},
	}

	// When: run theme select unknown
	var buf bytes.Buffer
	err := executeThemeSelectCommand(provider, "unknown", &buf)

	// Then: returns error mentioning unknown theme
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown")
}

func TestThemeSelectCommand_PersistsSelection(t *testing.T) {
	// Given: writable theme provider
	provider := &mockThemeLister{
		name:   "default",
		themes: []string{"default", "dark"},
	}

	// When: run theme select dark
	var buf bytes.Buffer
	err := executeThemeSelectCommand(provider, "dark", &buf)

	// Then: SetTheme was called with the correct name
	require.NoError(t, err)
	assert.Equal(t, "dark", provider.setCalled)
}

func TestThemeSelectCommand_SetError(t *testing.T) {
	// Given: ThemeProvider that returns an error on SetTheme
	provider := &mockThemeLister{
		name:   "default",
		themes: []string{"default", "dark"},
		setErr: assert.AnError,
	}

	// When: run theme select dark
	var buf bytes.Buffer
	err := executeThemeSelectCommand(provider, "dark", &buf)

	// Then: returns the error
	require.Error(t, err)
}
