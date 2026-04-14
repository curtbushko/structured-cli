package lint

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrettierParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData PrettierResultCompact
	}{
		{
			name:  "empty output indicates all files formatted",
			input: "",
			wantData: PrettierResultCompact{
				Success:        true,
				TotalChecked:   0,
				NeedFormatting: 0,
				Files:          []string{},
			},
		},
		{
			name:  "checking message with no files indicates success",
			input: "Checking formatting...",
			wantData: PrettierResultCompact{
				Success:        true,
				TotalChecked:   0,
				NeedFormatting: 0,
				Files:          []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewPrettierParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			require.NoError(t, err)
			require.Nil(t, result.Error)

			got, ok := result.Data.(*PrettierResultCompact)
			require.True(t, ok, "ParseResult.Data type = %T, want *PrettierResultCompact", result.Data)

			assert.Equal(t, tt.wantData.Success, got.Success)
			assert.Equal(t, tt.wantData.TotalChecked, got.TotalChecked)
			assert.Equal(t, tt.wantData.NeedFormatting, got.NeedFormatting)
			assert.Equal(t, tt.wantData.Files, got.Files)
		})
	}
}

func TestPrettierParser_SingleUnformattedFile(t *testing.T) {
	// Prettier --check outputs files that differ from prettier formatting
	input := `Checking formatting...
[warn] src/index.js
[warn] Code style issues found in the above file. Run Prettier to fix.`

	parser := NewPrettierParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*PrettierResultCompact)
	require.True(t, ok, "ParseResult.Data type = %T, want *PrettierResultCompact", result.Data)

	assert.False(t, got.Success, "Success should be false when unformatted files present")
	assert.Equal(t, 1, got.NeedFormatting)
	require.Len(t, got.Files, 1)
	assert.Equal(t, "src/index.js", got.Files[0])
}

func TestPrettierParser_MultipleUnformattedFiles(t *testing.T) {
	input := `Checking formatting...
[warn] src/index.js
[warn] src/utils.ts
[warn] src/app.tsx
[warn] Code style issues found in 3 files. Run Prettier to fix.`

	parser := NewPrettierParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*PrettierResultCompact)
	require.True(t, ok, "ParseResult.Data type = %T, want *PrettierResultCompact", result.Data)

	assert.False(t, got.Success, "Success should be false when unformatted files present")
	assert.Equal(t, 3, got.NeedFormatting)
	require.Len(t, got.Files, 3)

	wantFiles := []string{"src/index.js", "src/utils.ts", "src/app.tsx"}
	assert.Equal(t, wantFiles, got.Files)
}

func TestPrettierParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches prettier with --check",
			cmd:         "prettier",
			subcommands: []string{"--check"},
			want:        true,
		},
		{
			name:        "matches prettier with --check and path",
			cmd:         "prettier",
			subcommands: []string{"--check", "src/"},
			want:        true,
		},
		{
			name:        "matches prettier without subcommands",
			cmd:         "prettier",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "does not match npx prettier",
			cmd:         "npx",
			subcommands: []string{"prettier"},
			want:        false,
		},
		{
			name:        "does not match node",
			cmd:         "node",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewPrettierParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got, "Matches(%q, %v)", tt.cmd, tt.subcommands)
		})
	}
}

func TestPrettierParser_Schema(t *testing.T) {
	parser := NewPrettierParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID, "Schema.ID should not be empty")
	assert.NotEmpty(t, schema.Title, "Schema.Title should not be empty")
	assert.Equal(t, "object", schema.Type)

	// Verify required properties exist for compact format
	requiredProps := []string{"success", "total_checked", "need_formatting", "files"}
	for _, prop := range requiredProps {
		_, ok := schema.Properties[prop]
		assert.True(t, ok, "Schema.Properties missing %q", prop)
	}
}

func TestPrettierParser_AllFilesFormatted(t *testing.T) {
	// When all files are formatted, prettier --check outputs success message
	input := `Checking formatting...
All matched files use Prettier code style!`

	parser := NewPrettierParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*PrettierResultCompact)
	require.True(t, ok, "ParseResult.Data type = %T, want *PrettierResultCompact", result.Data)

	assert.True(t, got.Success, "Success should be true when all files formatted")
	assert.Equal(t, 0, got.NeedFormatting)
	assert.Empty(t, got.Files)
}

func TestPrettierParser_Counts(t *testing.T) {
	// Prettier --check outputs files that need formatting
	// Note: Prettier doesn't report total files checked, only files needing formatting
	input := `Checking formatting...
[warn] src/index.js
[warn] src/utils.ts
[warn] src/app.tsx
[warn] Code style issues found in 3 files. Run Prettier to fix.`

	parser := NewPrettierParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*PrettierResultCompact)
	require.True(t, ok, "ParseResult.Data type = %T, want *PrettierResultCompact", result.Data)

	// Verify counts
	assert.Equal(t, 0, got.TotalChecked, "TotalChecked should be 0 (Prettier doesn't report total)")
	assert.Equal(t, 3, got.NeedFormatting, "NeedFormatting should match file count")

	// Verify Files list
	require.Len(t, got.Files, 3, "Files list should have 3 entries")
	assert.Contains(t, got.Files, "src/index.js")
	assert.Contains(t, got.Files, "src/utils.ts")
	assert.Contains(t, got.Files, "src/app.tsx")

	// Verify success is false when files need formatting
	assert.False(t, got.Success, "Success should be false when files need formatting")
}
