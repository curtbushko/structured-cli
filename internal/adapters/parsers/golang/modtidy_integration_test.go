package golang_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/golang"
	"github.com/curtbushko/structured-cli/internal/application"
)

func TestModTidyParser_IntegrationMatch(t *testing.T) {
	t.Run("parser can be registered and found for go mod tidy", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewModTidyParser())

		parser, found := registry.Find("go", []string{"mod", "tidy"})

		if !found {
			t.Fatal("expected parser to be found for 'go mod tidy'")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser is found with flags", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewModTidyParser())

		parser, found := registry.Find("go", []string{"mod", "tidy", "-v"})

		if !found {
			t.Fatal("expected parser to be found for 'go mod tidy -v'")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser is not found for non-matching subcommands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewModTidyParser())

		_, found := registry.Find("go", []string{"mod", "download"})

		if found {
			t.Error("expected parser not to be found for go mod download")
		}
	})

	t.Run("parser is not found for non-go commands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewModTidyParser())

		_, found := registry.Find("make", []string{"mod", "tidy"})

		if found {
			t.Error("expected parser not to be found for make mod tidy")
		}
	})
}

func TestModTidyParser_SchemaValidation(t *testing.T) {
	parser := golang.NewModTidyParser()
	schema := parser.Schema()

	t.Run("schema has correct title", func(t *testing.T) {
		if schema.Title != "Go Mod Tidy Output" {
			t.Errorf("Schema.Title = %q, want %q", schema.Title, "Go Mod Tidy Output")
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
		expectedProps := []string{"added", "removed"}
		for _, prop := range expectedProps {
			if _, ok := schema.Properties[prop]; !ok {
				t.Errorf("Schema.Properties missing %q", prop)
			}
		}
	})
}

func TestModTidyParser_EndToEndFlow(t *testing.T) {
	// This test verifies the complete flow:
	// 1. Parser is registered
	// 2. Registry can find it
	// 3. Parser can parse input
	// 4. Result contains expected data

	registry := application.NewInMemoryParserRegistry()
	registry.Register(golang.NewModTidyParser())

	// Find the parser
	parser, found := registry.Find("go", []string{"mod", "tidy"})
	if !found {
		t.Fatal("parser not found in registry")
	}

	// Parse sample input - successful tidy with no changes (empty output)
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify the result
	modtidy, ok := result.Data.(*golang.ModTidyResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.ModTidyResult, got %T", result.Data)
	}

	if len(modtidy.Added) != 0 {
		t.Errorf("Added = %v, want empty slice", modtidy.Added)
	}
	if len(modtidy.Removed) != 0 {
		t.Errorf("Removed = %v, want empty slice", modtidy.Removed)
	}

	// Parse sample input - with additions
	addedInput := `go: added github.com/example/pkg v1.0.0
go: added github.com/another/pkg v2.0.0
`
	result, err = parser.Parse(strings.NewReader(addedInput))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	modtidy, ok = result.Data.(*golang.ModTidyResult)
	if !ok {
		t.Fatalf("result.Data is not *golang.ModTidyResult, got %T", result.Data)
	}

	if len(modtidy.Added) != 2 {
		t.Errorf("len(Added) = %d, want 2", len(modtidy.Added))
	}
}
