// Package application contains the business logic and use cases for structured-cli.
package application

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/curtbushko/structured-cli/internal/domain"
)

func TestErrorHandler_CommandFailure_JSON(t *testing.T) {
	// Arrange - Command failed with exit code 128 (git not a repo error)
	handler := NewErrorHandler()
	parseResult := domain.ParseResult{
		Data:     nil,
		Raw:      "fatal: not a git repository (or any of the parent directories): .git\n",
		ExitCode: 128,
		Error:    errors.New("command failed"),
	}

	// Act
	errResult := handler.FormatCommandError(parseResult)

	// Assert - JSON contains 'error' and 'exitCode' fields
	data, ok := errResult.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected Data to be map[string]any, got %T", errResult.Data)
	}

	errorMsg, ok := data["error"].(string)
	if !ok {
		t.Fatal("expected 'error' field to be a string")
	}
	if errorMsg == "" {
		t.Error("expected 'error' field to be non-empty")
	}

	exitCode, ok := data["exitCode"].(int)
	if !ok {
		t.Fatal("expected 'exitCode' field to be an int")
	}
	if exitCode != 128 {
		t.Errorf("expected exitCode 128, got %d", exitCode)
	}

	// Verify the result serializes to valid JSON
	jsonBytes, err := json.Marshal(errResult.Data)
	if err != nil {
		t.Fatalf("failed to marshal error result to JSON: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("error result is not valid JSON: %v", err)
	}
}

func TestErrorHandler_CommandFailure_Passthrough(t *testing.T) {
	// Arrange - Command failed with stderr output
	handler := NewErrorHandler()
	stderrOutput := "error: pathspec 'nonexistent' did not match any file(s) known to git\n"
	parseResult := domain.ParseResult{
		Data:     nil,
		Raw:      stderrOutput,
		ExitCode: 1,
		Error:    errors.New("command failed"),
	}

	// Act
	errResult := handler.FormatPassthroughError(parseResult)

	// Assert - Raw stderr output returned unchanged
	if errResult.Raw != stderrOutput {
		t.Errorf("expected raw output %q, got %q", stderrOutput, errResult.Raw)
	}

	// ExitCode should be preserved
	if errResult.ExitCode != 1 {
		t.Errorf("expected exitCode 1, got %d", errResult.ExitCode)
	}
}

func TestErrorHandler_ParserFailure_JSON(t *testing.T) {
	// Arrange - Parser returned error, raw output available
	handler := NewErrorHandler()
	rawOutput := "On branch main\nYour branch is up to date with 'origin/main'.\n\nunexpected format here"
	parseResult := domain.ParseResult{
		Data:     nil,
		Raw:      rawOutput,
		ExitCode: 0,
		Error:    errors.New("unexpected format: could not parse branch info"),
	}

	// Act
	errResult := handler.FormatParserError(parseResult)

	// Assert - JSON contains 'error', 'raw', and 'exitCode' fields
	data, ok := errResult.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected Data to be map[string]any, got %T", errResult.Data)
	}

	// Check error field
	errorMsg, ok := data["error"].(string)
	if !ok {
		t.Fatal("expected 'error' field to be a string")
	}
	if errorMsg == "" {
		t.Error("expected 'error' field to be non-empty")
	}

	// Check raw field
	raw, ok := data["raw"].(string)
	if !ok {
		t.Fatal("expected 'raw' field to be a string")
	}
	if raw != rawOutput {
		t.Errorf("expected raw %q, got %q", rawOutput, raw)
	}

	// Check exitCode field
	exitCode, ok := data["exitCode"].(int)
	if !ok {
		t.Fatal("expected 'exitCode' field to be an int")
	}
	if exitCode != 0 {
		t.Errorf("expected exitCode 0, got %d", exitCode)
	}

	// Verify the result serializes to valid JSON
	jsonBytes, err := json.Marshal(errResult.Data)
	if err != nil {
		t.Fatalf("failed to marshal error result to JSON: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("error result is not valid JSON: %v", err)
	}
}

func TestErrorHandler_FormatError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedSubstr string
	}{
		{
			name:           "simple error",
			err:            errors.New("command not found"),
			expectedSubstr: "command not found",
		},
		{
			name:           "wrapped error",
			err:            errors.New("parse error: unexpected token at line 5"),
			expectedSubstr: "parse error",
		},
		{
			name:           "nil error",
			err:            nil,
			expectedSubstr: "",
		},
	}

	handler := NewErrorHandler()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			msg := handler.FormatErrorMessage(tt.err)

			// Assert
			if tt.err == nil {
				if msg != "" {
					t.Errorf("expected empty message for nil error, got %q", msg)
				}
				return
			}

			if msg == "" {
				t.Error("expected non-empty error message")
			}

			// Error message should contain the expected substring
			if tt.expectedSubstr != "" && !containsSubstring(msg, tt.expectedSubstr) {
				t.Errorf("expected message to contain %q, got %q", tt.expectedSubstr, msg)
			}
		})
	}
}

func TestErrorHandler_PreservesExitCode(t *testing.T) {
	handler := NewErrorHandler()

	testCodes := []int{0, 1, 2, 128, 255}

	for _, code := range testCodes {
		t.Run("exit code preserved", func(t *testing.T) {
			parseResult := domain.ParseResult{
				Data:     nil,
				Raw:      "some output",
				ExitCode: code,
				Error:    errors.New("test error"),
			}

			// Test command error
			cmdResult := handler.FormatCommandError(parseResult)
			data := cmdResult.Data.(map[string]any)
			if data["exitCode"].(int) != code {
				t.Errorf("FormatCommandError: expected exitCode %d, got %d", code, data["exitCode"])
			}

			// Test parser error
			parserResult := handler.FormatParserError(parseResult)
			data = parserResult.Data.(map[string]any)
			if data["exitCode"].(int) != code {
				t.Errorf("FormatParserError: expected exitCode %d, got %d", code, data["exitCode"])
			}
		})
	}
}

// containsSubstring checks if s contains substr (case-insensitive would be better but simple for now).
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
