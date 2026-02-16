// Package writer provides implementations of the OutputWriter port.
// This package is in the adapters layer and implements the ports.OutputWriter
// interface for formatting and writing parse results.
package writer

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/curtbushko/structured-cli/internal/domain"
)

func TestJSONWriter_Write_ValidJSON(t *testing.T) {
	// Arrange
	data := map[string]any{
		"branch":    "main",
		"clean":     true,
		"staged":    []string{},
		"modified":  []string{"README.md"},
		"untracked": []string{},
	}
	result := domain.NewParseResult(data, "raw output", 0)
	schema := domain.Schema{}
	writer := NewJSONWriter(false)
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify data matches
	if parsed["branch"] != "main" {
		t.Errorf("expected branch 'main', got %v", parsed["branch"])
	}
	if parsed["clean"] != true {
		t.Errorf("expected clean true, got %v", parsed["clean"])
	}
}

func TestJSONWriter_Write_Indented(t *testing.T) {
	// Arrange
	data := map[string]any{
		"branch": "main",
		"clean":  true,
	}
	result := domain.NewParseResult(data, "raw output", 0)
	schema := domain.Schema{}
	writer := NewJSONWriter(true) // Indent=true
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()

	// Verify output contains indentation (newlines and spaces)
	if !bytes.Contains(buf.Bytes(), []byte("\n")) {
		t.Error("expected indented output to contain newlines")
	}
	if !bytes.Contains(buf.Bytes(), []byte("  ")) {
		t.Error("expected indented output to contain spaces for indentation")
	}

	// Verify it's still valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("indented output is not valid JSON: %v", err)
	}
}

func TestJSONWriter_Write_NilData(t *testing.T) {
	// Arrange
	result := domain.NewParseResult(nil, "raw output", 0)
	schema := domain.Schema{}
	writer := NewJSONWriter(false)
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify output is "null" (with newline from encoder)
	output := bytes.TrimSpace(buf.Bytes())
	if string(output) != "null" {
		t.Errorf("expected 'null', got %q", string(output))
	}
}

func TestJSONWriter_Write_ComplexData(t *testing.T) {
	// Arrange - nested structure
	data := map[string]any{
		"commits": []map[string]any{
			{
				"hash":    "abc123",
				"author":  "John Doe",
				"message": "Initial commit",
			},
			{
				"hash":    "def456",
				"author":  "Jane Doe",
				"message": "Add feature",
			},
		},
		"total": 2,
	}
	result := domain.NewParseResult(data, "raw output", 0)
	schema := domain.Schema{}
	writer := NewJSONWriter(false)
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify output is valid JSON with expected structure
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	commits, ok := parsed["commits"].([]any)
	if !ok {
		t.Fatal("expected commits to be an array")
	}
	if len(commits) != 2 {
		t.Errorf("expected 2 commits, got %d", len(commits))
	}
}

func TestJSONWriter_Write_ArrayData(t *testing.T) {
	// Arrange - top-level array
	data := []string{"file1.txt", "file2.txt", "file3.txt"}
	result := domain.NewParseResult(data, "raw output", 0)
	schema := domain.Schema{}
	writer := NewJSONWriter(false)
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var parsed []string
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}
	if len(parsed) != 3 {
		t.Errorf("expected 3 items, got %d", len(parsed))
	}
}

func TestJSONWriter_Write_ErrorResult(t *testing.T) {
	// Arrange - error result with error and exitCode fields
	data := map[string]any{
		"error":    "fatal: not a git repository",
		"exitCode": 128,
	}
	result := domain.NewParseResult(data, "fatal: not a git repository\n", 128)
	schema := domain.Schema{}
	writer := NewJSONWriter(false)
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify error field
	errorMsg, ok := parsed["error"].(string)
	if !ok {
		t.Fatal("expected 'error' field to be a string")
	}
	if errorMsg != "fatal: not a git repository" {
		t.Errorf("expected error 'fatal: not a git repository', got %q", errorMsg)
	}

	// Verify exitCode field (JSON numbers are float64)
	exitCode, ok := parsed["exitCode"].(float64)
	if !ok {
		t.Fatal("expected 'exitCode' field to be a number")
	}
	if int(exitCode) != 128 {
		t.Errorf("expected exitCode 128, got %d", int(exitCode))
	}
}

func TestJSONWriter_Write_ParserErrorResult(t *testing.T) {
	// Arrange - parser error result with error, raw, and exitCode fields
	rawOutput := "On branch main\nunexpected format"
	data := map[string]any{
		"error":    "parser error: unexpected format",
		"raw":      rawOutput,
		"exitCode": 0,
	}
	result := domain.NewParseResult(data, rawOutput, 0)
	schema := domain.Schema{}
	writer := NewJSONWriter(false)
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify all three fields
	if _, ok := parsed["error"].(string); !ok {
		t.Fatal("expected 'error' field to be a string")
	}
	if raw, ok := parsed["raw"].(string); !ok || raw != rawOutput {
		t.Errorf("expected 'raw' field to be %q, got %v", rawOutput, parsed["raw"])
	}
	if exitCode, ok := parsed["exitCode"].(float64); !ok || int(exitCode) != 0 {
		t.Errorf("expected 'exitCode' field to be 0, got %v", parsed["exitCode"])
	}
}

func TestJSONWriter_Write_FallbackResult(t *testing.T) {
	// Arrange - FallbackResult for unsupported command
	rawOutput := "stash@{0}: WIP on main: abc123 Some commit message"
	fallbackResult := domain.NewFallbackResult(rawOutput, 0)
	result := domain.NewParseResult(fallbackResult, rawOutput, 0)
	schema := domain.Schema{}
	writer := NewJSONWriter(false)
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify 'raw' field contains the raw output
	raw, ok := parsed["raw"].(string)
	if !ok {
		t.Fatal("expected 'raw' field to be a string")
	}
	if raw != rawOutput {
		t.Errorf("expected raw %q, got %q", rawOutput, raw)
	}

	// Verify 'parsed' field is false
	parsedFlag, ok := parsed["parsed"].(bool)
	if !ok {
		t.Fatal("expected 'parsed' field to be a boolean")
	}
	if parsedFlag != false {
		t.Errorf("expected parsed false, got %v", parsedFlag)
	}

	// Verify 'exitCode' field
	exitCode, ok := parsed["exitCode"].(float64)
	if !ok {
		t.Fatal("expected 'exitCode' field to be a number")
	}
	if int(exitCode) != 0 {
		t.Errorf("expected exitCode 0, got %d", int(exitCode))
	}
}

func TestJSONWriter_Write_FallbackResult_NonZeroExit(t *testing.T) {
	// Arrange - FallbackResult with non-zero exit code
	rawOutput := "No stash entries found."
	fallbackResult := domain.NewFallbackResult(rawOutput, 1)
	result := domain.NewParseResult(fallbackResult, rawOutput, 1)
	schema := domain.Schema{}
	writer := NewJSONWriter(false)
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify output is valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	// Verify 'exitCode' field is 1
	exitCode, ok := parsed["exitCode"].(float64)
	if !ok {
		t.Fatal("expected 'exitCode' field to be a number")
	}
	if int(exitCode) != 1 {
		t.Errorf("expected exitCode 1, got %d", int(exitCode))
	}
}

func TestJSONWriter_Write_FallbackResult_Indented(t *testing.T) {
	// Arrange - FallbackResult with indentation
	rawOutput := "line1\nline2\nline3"
	fallbackResult := domain.NewFallbackResult(rawOutput, 0)
	result := domain.NewParseResult(fallbackResult, rawOutput, 0)
	schema := domain.Schema{}
	writer := NewJSONWriter(true) // Indented
	var buf bytes.Buffer

	// Act
	err := writer.Write(&buf, result, schema)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()

	// Verify output contains indentation (newlines and spaces)
	if !bytes.Contains(buf.Bytes(), []byte("\n")) {
		t.Error("expected indented output to contain newlines")
	}

	// Verify it's still valid JSON
	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("indented output is not valid JSON: %v", err)
	}

	// Verify structure
	if _, ok := parsed["raw"].(string); !ok {
		t.Error("expected 'raw' field")
	}
	if _, ok := parsed["parsed"].(bool); !ok {
		t.Error("expected 'parsed' field")
	}
}
