package golang_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/golang"
	"github.com/curtbushko/structured-cli/internal/application"
)

func TestTestParser_IntegrationMatch(t *testing.T) {
	t.Run("parser can be registered and found for go test", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewTestParser())

		parser, found := registry.Find("go", []string{"test"})

		if !found {
			t.Fatal("expected parser to be found for 'go test'")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser is not found for non-matching subcommands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewTestParser())

		_, found := registry.Find("go", []string{"build"})

		if found {
			t.Error("expected parser not to be found for go build")
		}
	})

	t.Run("parser is not found for non-go commands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewTestParser())

		_, found := registry.Find("make", []string{"test"})

		if found {
			t.Error("expected parser not to be found for make test")
		}
	})
}

func TestTestParser_SchemaValidation(t *testing.T) {
	parser := golang.NewTestParser()
	schema := parser.Schema()

	t.Run("schema has correct title", func(t *testing.T) {
		if schema.Title != "Go Test Output" {
			t.Errorf("Schema.Title = %q, want %q", schema.Title, "Go Test Output")
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
		expectedProps := []string{"passed", "failed", "skipped", "packages"}
		for _, prop := range expectedProps {
			if _, ok := schema.Properties[prop]; !ok {
				t.Errorf("Schema.Properties missing %q", prop)
			}
		}
	})
}

func TestTestParser_EndToEndFlow(t *testing.T) {
	// This test verifies the complete flow:
	// 1. Parser is registered
	// 2. Registry can find it
	// 3. Parser can parse input
	// 4. Result contains expected data

	registry := application.NewInMemoryParserRegistry()
	registry.Register(golang.NewTestParser())

	// Find the parser
	parser, found := registry.Find("go", []string{"test"})
	if !found {
		t.Fatal("parser not found in registry")
	}

	// Parse sample input - go test -json output
	jsonInput := `{"Time":"2024-01-01T10:00:00Z","Action":"run","Package":"example.com/pkg","Test":"TestFoo"}
{"Time":"2024-01-01T10:00:01Z","Action":"pass","Package":"example.com/pkg","Test":"TestFoo","Elapsed":0.5}
{"Time":"2024-01-01T10:00:01Z","Action":"pass","Package":"example.com/pkg","Elapsed":1.0}
`
	result, err := parser.Parse(strings.NewReader(jsonInput))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify the result
	testResult, ok := result.Data.(*golang.TestResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.TestResult, got %T", result.Data)
	}

	if testResult.Passed != 1 {
		t.Errorf("Passed = %d, want 1", testResult.Passed)
	}

	if testResult.Failed != 0 {
		t.Errorf("Failed = %d, want 0", testResult.Failed)
	}
}
