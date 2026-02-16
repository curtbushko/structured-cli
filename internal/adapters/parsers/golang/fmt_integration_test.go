package golang_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/golang"
	"github.com/curtbushko/structured-cli/internal/application"
)

func TestFmtParser_IntegrationMatch(t *testing.T) {
	t.Run("parser can be registered and found for gofmt", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewFmtParser())

		parser, found := registry.Find("gofmt", nil)

		if !found {
			t.Fatal("expected parser to be found for 'gofmt'")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser can be registered and found for go fmt", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewFmtParser())

		parser, found := registry.Find("go", []string{"fmt"})

		if !found {
			t.Fatal("expected parser to be found for 'go fmt'")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser is not found for non-matching subcommands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewFmtParser())

		_, found := registry.Find("go", []string{"build"})

		if found {
			t.Error("expected parser not to be found for go build")
		}
	})

	t.Run("parser is not found for non-go commands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewFmtParser())

		_, found := registry.Find("make", []string{"fmt"})

		if found {
			t.Error("expected parser not to be found for make fmt")
		}
	})
}

func TestFmtParser_SchemaValidation(t *testing.T) {
	parser := golang.NewFmtParser()
	schema := parser.Schema()

	t.Run("schema has correct title", func(t *testing.T) {
		if schema.Title != "Go Fmt Output" {
			t.Errorf("Schema.Title = %q, want %q", schema.Title, "Go Fmt Output")
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
		expectedProps := []string{"unformatted"}
		for _, prop := range expectedProps {
			if _, ok := schema.Properties[prop]; !ok {
				t.Errorf("Schema.Properties missing %q", prop)
			}
		}
	})
}

func TestFmtParser_EndToEndFlow(t *testing.T) {
	// This test verifies the complete flow:
	// 1. Parser is registered
	// 2. Registry can find it
	// 3. Parser can parse input
	// 4. Result contains expected data

	registry := application.NewInMemoryParserRegistry()
	registry.Register(golang.NewFmtParser())

	// Find the parser
	parser, found := registry.Find("gofmt", nil)
	if !found {
		t.Fatal("parser not found in registry")
	}

	// Parse sample input - no files need formatting (empty output)
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify the result
	fmtResult, ok := result.Data.(*golang.FmtResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.FmtResult, got %T", result.Data)
	}

	if len(fmtResult.Unformatted) != 0 {
		t.Errorf("len(Unformatted) = %d, want 0 for empty output", len(fmtResult.Unformatted))
	}

	// Parse sample input - files need formatting
	filesInput := `main.go
internal/foo/bar.go
`
	result, err = parser.Parse(strings.NewReader(filesInput))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	fmtResult, ok = result.Data.(*golang.FmtResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.FmtResult, got %T", result.Data)
	}

	if len(fmtResult.Unformatted) != 2 {
		t.Errorf("len(Unformatted) = %d, want 2", len(fmtResult.Unformatted))
	}
}
