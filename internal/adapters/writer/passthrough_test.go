package writer

import (
	"bytes"
	"testing"

	"github.com/curtbushko/structured-cli/internal/domain"
)

func TestPassthroughWriter_Write(t *testing.T) {
	// Arrange
	raw := "On branch main\nnothing to commit, working tree clean\n"
	result := domain.NewParseResult(nil, raw, 0)
	schema := domain.Schema{}
	writer := NewPassthroughWriter()
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Output should match raw exactly
	if buf.String() != raw {
		t.Errorf("expected %q, got %q", raw, buf.String())
	}
}

func TestPassthroughWriter_Write_PreservesNewlines(t *testing.T) {
	// Arrange - multiline output with various newline patterns
	raw := "line1\nline2\nline3\n\nline5 after blank\n"
	result := domain.NewParseResult(nil, raw, 0)
	schema := domain.Schema{}
	writer := NewPassthroughWriter()
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// All newlines should be preserved exactly
	if buf.String() != raw {
		t.Errorf("newlines not preserved:\nexpected: %q\ngot: %q", raw, buf.String())
	}
}

func TestPassthroughWriter_Write_EmptyString(t *testing.T) {
	// Arrange
	result := domain.NewParseResult(nil, "", 0)
	schema := domain.Schema{}
	writer := NewPassthroughWriter()
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if buf.String() != "" {
		t.Errorf("expected empty string, got %q", buf.String())
	}
}

func TestPassthroughWriter_Write_SpecialCharacters(t *testing.T) {
	// Arrange - output with special characters, tabs, unicode
	raw := "file\twith\ttabs\n\u2713 checkmark\n\x1b[32mcolored\x1b[0m\n"
	result := domain.NewParseResult(nil, raw, 0)
	schema := domain.Schema{}
	writer := NewPassthroughWriter()
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// All special characters should be preserved
	if buf.String() != raw {
		t.Errorf("special characters not preserved:\nexpected: %q\ngot: %q", raw, buf.String())
	}
}

func TestPassthroughWriter_Write_LargeOutput(t *testing.T) {
	// Arrange - simulate large command output
	var rawBuilder bytes.Buffer
	for i := 0; i < 1000; i++ {
		rawBuilder.WriteString("line of output that is fairly long to simulate real command output\n")
	}
	raw := rawBuilder.String()
	result := domain.NewParseResult(nil, raw, 0)
	schema := domain.Schema{}
	writer := NewPassthroughWriter()
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if buf.String() != raw {
		t.Errorf("large output not preserved (length: expected %d, got %d)", len(raw), buf.Len())
	}
}

func TestPassthroughWriter_Write_IgnoresData(t *testing.T) {
	// Arrange - data is set but should be ignored
	data := map[string]any{"branch": "main"}
	raw := "On branch main\n"
	result := domain.NewParseResult(data, raw, 0)
	schema := domain.Schema{}
	writer := NewPassthroughWriter()
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should output raw, not JSON
	if buf.String() != raw {
		t.Errorf("expected raw output %q, got %q", raw, buf.String())
	}
}
