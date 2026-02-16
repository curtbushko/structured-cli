package golang_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/golang"
	"github.com/curtbushko/structured-cli/internal/application"
)

func TestRunParser_IntegrationMatch(t *testing.T) {
	t.Run("parser can be registered and found for go run", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewRunParser())

		parser, found := registry.Find("go", []string{"run"})

		if !found {
			t.Fatal("expected parser to be found for 'go run'")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser is not found for non-matching subcommands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewRunParser())

		_, found := registry.Find("go", []string{"build"})

		if found {
			t.Error("expected parser not to be found for go build")
		}
	})

	t.Run("parser is not found for non-go commands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewRunParser())

		_, found := registry.Find("make", []string{"run"})

		if found {
			t.Error("expected parser not to be found for make run")
		}
	})
}

func TestRunParser_SchemaValidation(t *testing.T) {
	parser := golang.NewRunParser()
	schema := parser.Schema()

	t.Run("schema has correct title", func(t *testing.T) {
		if schema.Title != "Go Run Output" {
			t.Errorf("Schema.Title = %q, want %q", schema.Title, "Go Run Output")
		}
	})

	t.Run("schema has ID", func(t *testing.T) {
		if schema.ID == "" {
			t.Error("Schema.ID is empty")
		}
	})

	t.Run("schema has properties", func(t *testing.T) {
		if schema.Properties == nil {
			t.Fatal("Schema.Properties is nil")
		}
		if len(schema.Properties) == 0 {
			t.Error("Schema.Properties is empty")
		}
	})

	t.Run("schema has expected properties", func(t *testing.T) {
		expectedProps := []string{"exitCode", "stdout", "stderr"}
		for _, prop := range expectedProps {
			if _, ok := schema.Properties[prop]; !ok {
				t.Errorf("Schema.Properties missing %q", prop)
			}
		}
	})
}

func TestRunParser_EndToEndFlow(t *testing.T) {
	// This test verifies the complete flow:
	// 1. Parser is registered
	// 2. Registry can find it
	// 3. Parser can parse input
	// 4. Result contains expected data

	registry := application.NewInMemoryParserRegistry()
	registry.Register(golang.NewRunParser())

	// Find the parser
	parser, found := registry.Find("go", []string{"run"})
	if !found {
		t.Fatal("parser not found in registry")
	}

	// Parse sample input - simple output
	result, err := parser.Parse(strings.NewReader("Hello, World!"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify the result
	runResult, ok := result.Data.(*golang.RunResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.RunResult, got %T", result.Data)
	}

	if runResult.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", runResult.ExitCode)
	}

	if runResult.Stdout != "Hello, World!" {
		t.Errorf("Stdout = %q, want %q", runResult.Stdout, "Hello, World!")
	}
}
