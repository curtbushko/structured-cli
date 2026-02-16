package golang_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/golang"
	"github.com/curtbushko/structured-cli/internal/application"
)

func TestVetParser_IntegrationMatch(t *testing.T) {
	t.Run("parser can be registered and found for go vet", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewVetParser())

		parser, found := registry.Find("go", []string{"vet"})

		if !found {
			t.Fatal("expected parser to be found for 'go vet'")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser is not found for non-matching subcommands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewVetParser())

		_, found := registry.Find("go", []string{"build"})

		if found {
			t.Error("expected parser not to be found for go build")
		}
	})

	t.Run("parser is not found for non-go commands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewVetParser())

		_, found := registry.Find("make", []string{"vet"})

		if found {
			t.Error("expected parser not to be found for make vet")
		}
	})
}

func TestVetParser_SchemaValidation(t *testing.T) {
	parser := golang.NewVetParser()
	schema := parser.Schema()

	t.Run("schema has correct title", func(t *testing.T) {
		if schema.Title != "Go Vet Output" {
			t.Errorf("Schema.Title = %q, want %q", schema.Title, "Go Vet Output")
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
		expectedProps := []string{"issues"}
		for _, prop := range expectedProps {
			if _, ok := schema.Properties[prop]; !ok {
				t.Errorf("Schema.Properties missing %q", prop)
			}
		}
	})
}

func TestVetParser_EndToEndFlow(t *testing.T) {
	// This test verifies the complete flow:
	// 1. Parser is registered
	// 2. Registry can find it
	// 3. Parser can parse input
	// 4. Result contains expected data

	registry := application.NewInMemoryParserRegistry()
	registry.Register(golang.NewVetParser())

	// Find the parser
	parser, found := registry.Find("go", []string{"vet"})
	if !found {
		t.Fatal("parser not found in registry")
	}

	// Parse sample input - clean vet output (no issues)
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify the result
	vetResult, ok := result.Data.(*golang.VetResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.VetResult, got %T", result.Data)
	}

	if len(vetResult.Issues) != 0 {
		t.Errorf("len(Issues) = %d, want 0 for clean vet output", len(vetResult.Issues))
	}

	// Parse sample input - vet with issues
	issueInput := `main.go:10:5: printf call has arguments but no formatting directives
utils.go:25:10: unreachable code
`
	result, err = parser.Parse(strings.NewReader(issueInput))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	vetResult, ok = result.Data.(*golang.VetResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.VetResult, got %T", result.Data)
	}

	if len(vetResult.Issues) != 2 {
		t.Errorf("len(Issues) = %d, want 2", len(vetResult.Issues))
	}
}
