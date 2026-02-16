package golang_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/golang"
	"github.com/curtbushko/structured-cli/internal/application"
)

func TestGenerateParser_IntegrationMatch(t *testing.T) {
	t.Run("parser can be registered and found for go generate", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewGenerateParser())

		parser, found := registry.Find("go", []string{"generate"})

		if !found {
			t.Fatal("expected parser to be found for 'go generate'")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser is not found for non-matching subcommands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewGenerateParser())

		_, found := registry.Find("go", []string{"build"})

		if found {
			t.Error("expected parser not to be found for go build")
		}
	})

	t.Run("parser is not found for non-go commands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewGenerateParser())

		_, found := registry.Find("make", []string{"generate"})

		if found {
			t.Error("expected parser not to be found for make generate")
		}
	})
}

func TestGenerateParser_SchemaValidation(t *testing.T) {
	parser := golang.NewGenerateParser()
	schema := parser.Schema()

	t.Run("schema has correct title", func(t *testing.T) {
		if schema.Title != "Go Generate Output" {
			t.Errorf("Schema.Title = %q, want %q", schema.Title, "Go Generate Output")
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
		expectedProps := []string{"success", "generated"}
		for _, prop := range expectedProps {
			if _, ok := schema.Properties[prop]; !ok {
				t.Errorf("Schema.Properties missing %q", prop)
			}
		}
	})
}

func TestGenerateParser_EndToEndFlow(t *testing.T) {
	// This test verifies the complete flow:
	// 1. Parser is registered
	// 2. Registry can find it
	// 3. Parser can parse input
	// 4. Result contains expected data

	registry := application.NewInMemoryParserRegistry()
	registry.Register(golang.NewGenerateParser())

	// Find the parser
	parser, found := registry.Find("go", []string{"generate"})
	if !found {
		t.Fatal("parser not found in registry")
	}

	// Parse sample input - no output (successful, no directives)
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify the result
	genResult, ok := result.Data.(*golang.GenerateResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.GenerateResult, got %T", result.Data)
	}

	if !genResult.Success {
		t.Error("Success = false, want true for empty output")
	}
	if len(genResult.Generated) != 0 {
		t.Errorf("len(Generated) = %d, want 0 for empty output", len(genResult.Generated))
	}

	// Parse sample input - go generate -v output
	generateInput := `generate.go:3: running stringer -type=Pill
models.go:10: running mockgen -source=models.go -destination=mock_models.go
`
	result, err = parser.Parse(strings.NewReader(generateInput))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	genResult, ok = result.Data.(*golang.GenerateResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.GenerateResult, got %T", result.Data)
	}

	if !genResult.Success {
		t.Error("Success = false, want true")
	}
	if len(genResult.Generated) != 2 {
		t.Errorf("len(Generated) = %d, want 2", len(genResult.Generated))
	}
}
