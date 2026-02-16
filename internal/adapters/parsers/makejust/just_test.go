package makejust

import (
	"strings"
	"testing"
)

func TestJustParser_EmptyOutput(t *testing.T) {
	parser := NewJustParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*JustResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JustResult", result.Data)
	}

	if !got.Success {
		t.Error("JustResult.Success = false, want true for empty output")
	}
}

func TestJustParser_SuccessfulRecipe(t *testing.T) {
	input := `echo "Building project..."
Building project...
gcc -o app main.c`

	parser := NewJustParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*JustResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JustResult", result.Data)
	}

	if !got.Success {
		t.Error("JustResult.Success = false, want true")
	}

	if got.ExitCode != 0 {
		t.Errorf("JustResult.ExitCode = %d, want 0", got.ExitCode)
	}
}

func TestJustParser_RecipeFailure(t *testing.T) {
	input := `echo "Running tests..."
Running tests...
error: Recipe 'test' failed on line 5 with exit code 1`

	parser := NewJustParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*JustResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JustResult", result.Data)
	}

	if got.Success {
		t.Error("JustResult.Success = true, want false for recipe failure")
	}

	if got.Error == "" {
		t.Error("JustResult.Error should not be empty for recipe failure")
	}

	if got.ExitCode != 1 {
		t.Errorf("JustResult.ExitCode = %d, want 1", got.ExitCode)
	}
}

func TestJustParser_RecipeListing(t *testing.T) {
	input := `Available recipes:
    build       # Build the application
    test        # Run tests
    clean       # Clean build artifacts
    deploy arg  # Deploy to environment`

	parser := NewJustParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*JustResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JustResult", result.Data)
	}

	if len(got.Recipes) < 4 {
		t.Fatalf("JustResult.Recipes length = %d, want >= 4", len(got.Recipes))
	}

	// Check build recipe
	found := false
	for _, recipe := range got.Recipes {
		if recipe.Name == "build" {
			found = true
			if recipe.Description != "Build the application" {
				t.Errorf("Recipe.Description = %q, want %q", recipe.Description, "Build the application")
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find recipe 'build' in Recipes list")
	}

	// Check deploy recipe with parameter
	for _, recipe := range got.Recipes {
		if recipe.Name == "deploy" {
			if len(recipe.Parameters) != 1 {
				t.Errorf("Recipe 'deploy' Parameters length = %d, want 1", len(recipe.Parameters))
			}
			if len(recipe.Parameters) > 0 && recipe.Parameters[0].Name != "arg" {
				t.Errorf("Recipe 'deploy' Parameter name = %q, want %q", recipe.Parameters[0].Name, "arg")
			}
			break
		}
	}
}

func TestJustParser_DryRun(t *testing.T) {
	input := `#!/usr/bin/env bash
set -euo pipefail
echo "Building..."
gcc -c main.c -o main.o
gcc main.o -o myapp`

	parser := NewJustParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*JustResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JustResult", result.Data)
	}

	if len(got.Commands) < 3 {
		t.Errorf("JustResult.Commands length = %d, want >= 3", len(got.Commands))
	}
}

func TestJustParser_UnknownRecipe(t *testing.T) {
	input := `error: Justfile does not contain recipe 'nonexistent'.`

	parser := NewJustParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*JustResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JustResult", result.Data)
	}

	if got.Success {
		t.Error("JustResult.Success = true, want false for unknown recipe")
	}

	if !strings.Contains(got.Error, "does not contain recipe") {
		t.Errorf("JustResult.Error = %q, should contain 'does not contain recipe'", got.Error)
	}
}

func TestJustParser_NoJustfile(t *testing.T) {
	input := `error: No justfile found`

	parser := NewJustParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*JustResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JustResult", result.Data)
	}

	if got.Success {
		t.Error("JustResult.Success = true, want false for missing justfile")
	}

	if !strings.Contains(got.Error, "No justfile") {
		t.Errorf("JustResult.Error = %q, should contain 'No justfile'", got.Error)
	}
}

func TestJustParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches just",
			cmd:         "just",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches just with recipe",
			cmd:         "just",
			subcommands: []string{"build"},
			want:        true,
		},
		{
			name:        "matches just with flags",
			cmd:         "just",
			subcommands: []string{"--list"},
			want:        true,
		},
		{
			name:        "matches just with recipe and args",
			cmd:         "just",
			subcommands: []string{"deploy", "production"},
			want:        true,
		},
		{
			name:        "does not match make",
			cmd:         "make",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match other commands",
			cmd:         "ninja",
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

	parser := NewJustParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestJustParser_Schema(t *testing.T) {
	parser := NewJustParser()
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

	requiredProps := []string{"success", "exit_code"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestJustParser_RecipeWithDefaultParam(t *testing.T) {
	input := `Available recipes:
    greet name='World' # Greet someone`

	parser := NewJustParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*JustResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JustResult", result.Data)
	}

	if len(got.Recipes) != 1 {
		t.Fatalf("JustResult.Recipes length = %d, want 1", len(got.Recipes))
	}

	recipe := got.Recipes[0]
	if recipe.Name != "greet" {
		t.Errorf("Recipe.Name = %q, want %q", recipe.Name, "greet")
	}

	if len(recipe.Parameters) != 1 {
		t.Fatalf("Recipe.Parameters length = %d, want 1", len(recipe.Parameters))
	}

	param := recipe.Parameters[0]
	if param.Name != "name" {
		t.Errorf("Parameter.Name = %q, want %q", param.Name, "name")
	}

	if param.Default != "World" {
		t.Errorf("Parameter.Default = %q, want %q", param.Default, "World")
	}
}
