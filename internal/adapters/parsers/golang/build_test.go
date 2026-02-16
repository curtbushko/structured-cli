package golang

import (
	"strings"
	"testing"
)

func TestBuildParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData Build
	}{
		{
			name:  "empty output indicates successful build",
			input: "",
			wantData: Build{
				Success:  true,
				Packages: []string{},
				Errors:   []BuildError{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewBuildParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*Build)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *Build", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("Build.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Packages) != len(tt.wantData.Packages) {
				t.Errorf("Build.Packages length = %d, want %d", len(got.Packages), len(tt.wantData.Packages))
			}

			if len(got.Errors) != len(tt.wantData.Errors) {
				t.Errorf("Build.Errors length = %d, want %d", len(got.Errors), len(tt.wantData.Errors))
			}
		})
	}
}

func TestBuildParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches go build",
			cmd:         "go",
			subcommands: []string{"build"},
			want:        true,
		},
		{
			name:        "matches go build with path",
			cmd:         "go",
			subcommands: []string{"build", "./..."},
			want:        true,
		},
		{
			name:        "does not match go test",
			cmd:         "go",
			subcommands: []string{"test"},
			want:        false,
		},
		{
			name:        "does not match git",
			cmd:         "git",
			subcommands: []string{"build"},
			want:        false,
		},
		{
			name:        "does not match go without subcommand",
			cmd:         "go",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{"build"},
			want:        false,
		},
	}

	parser := NewBuildParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestBuildParser_Schema(t *testing.T) {
	parser := NewBuildParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema.ID should not be empty")
	}

	if schema.Title == "" {
		t.Error("Schema.Title should not be empty")
	}

	if schema.Type != "object" {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, "object")
	}

	// Verify required properties exist
	requiredProps := []string{"success", "packages", "errors"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestBuildParser_SingleError(t *testing.T) {
	input := "main.go:10:5: undefined: foo"

	parser := NewBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*Build)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Build", result.Data)
	}

	if got.Success {
		t.Error("Build.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("Build.Errors length = %d, want 1", len(got.Errors))
	}

	wantErr := BuildError{
		File:    "main.go",
		Line:    10,
		Column:  5,
		Message: "undefined: foo",
	}

	if got.Errors[0] != wantErr {
		t.Errorf("Build.Errors[0] = %+v, want %+v", got.Errors[0], wantErr)
	}
}

func TestBuildParser_MultipleErrors(t *testing.T) {
	input := `main.go:10:5: undefined: foo
main.go:15:10: cannot use x (type int) as type string
utils.go:25:3: missing return`

	parser := NewBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*Build)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Build", result.Data)
	}

	if got.Success {
		t.Error("Build.Success = true, want false when errors present")
	}

	if len(got.Errors) != 3 {
		t.Fatalf("Build.Errors length = %d, want 3", len(got.Errors))
	}

	wantErrors := []BuildError{
		{File: "main.go", Line: 10, Column: 5, Message: "undefined: foo"},
		{File: "main.go", Line: 15, Column: 10, Message: "cannot use x (type int) as type string"},
		{File: "utils.go", Line: 25, Column: 3, Message: "missing return"},
	}

	for i, wantErr := range wantErrors {
		if got.Errors[i] != wantErr {
			t.Errorf("Build.Errors[%d] = %+v, want %+v", i, got.Errors[i], wantErr)
		}
	}
}

func TestBuildParser_PackageError(t *testing.T) {
	input := "package main: error importing package"

	parser := NewBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*Build)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Build", result.Data)
	}

	if got.Success {
		t.Error("Build.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("Build.Errors length = %d, want 1", len(got.Errors))
	}

	// Package errors don't have line/column info
	if got.Errors[0].Message != "package main: error importing package" {
		t.Errorf("Build.Errors[0].Message = %q, want %q", got.Errors[0].Message, "package main: error importing package")
	}
}

func TestBuildParser_ErrorNoColumn(t *testing.T) {
	input := "main.go:10: syntax error: unexpected EOF"

	parser := NewBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*Build)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Build", result.Data)
	}

	if got.Success {
		t.Error("Build.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("Build.Errors length = %d, want 1", len(got.Errors))
	}

	wantErr := BuildError{
		File:    "main.go",
		Line:    10,
		Column:  0, // No column in this format
		Message: "syntax error: unexpected EOF",
	}

	if got.Errors[0] != wantErr {
		t.Errorf("Build.Errors[0] = %+v, want %+v", got.Errors[0], wantErr)
	}
}
