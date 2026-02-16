package golang_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/golang"
	"github.com/curtbushko/structured-cli/internal/application"
)

func TestBuildParser_IntegrationMatch(t *testing.T) {
	t.Run("parser can be registered and found for go build", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewBuildParser())

		parser, found := registry.Find("go", []string{"build"})

		if !found {
			t.Fatal("expected parser to be found for 'go build'")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser is not found for non-matching subcommands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewBuildParser())

		_, found := registry.Find("go", []string{"test"})

		if found {
			t.Error("expected parser not to be found for go test")
		}
	})

	t.Run("parser is not found for non-go commands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(golang.NewBuildParser())

		_, found := registry.Find("make", []string{"build"})

		if found {
			t.Error("expected parser not to be found for make build")
		}
	})
}

func TestBuildParser_SchemaValidation(t *testing.T) {
	parser := golang.NewBuildParser()
	schema := parser.Schema()

	t.Run("schema has correct title", func(t *testing.T) {
		if schema.Title != "Go Build Output" {
			t.Errorf("Schema.Title = %q, want %q", schema.Title, "Go Build Output")
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
		expectedProps := []string{"success", "packages", "errors"}
		for _, prop := range expectedProps {
			if _, ok := schema.Properties[prop]; !ok {
				t.Errorf("Schema.Properties missing %q", prop)
			}
		}
	})
}

func TestBuildParser_EndToEndFlow(t *testing.T) {
	// This test verifies the complete flow:
	// 1. Parser is registered
	// 2. Registry can find it
	// 3. Parser can parse input
	// 4. Result contains expected data

	registry := application.NewInMemoryParserRegistry()
	registry.Register(golang.NewBuildParser())

	// Find the parser
	parser, found := registry.Find("go", []string{"build"})
	if !found {
		t.Fatal("parser not found in registry")
	}

	// Parse sample input - successful build (no output)
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify the result
	build, ok := result.Data.(*golang.Build)
	if !ok {
		t.Fatalf("result.Data is not *golang.Build, got %T", result.Data)
	}

	if !build.Success {
		t.Error("Success = false, want true for empty output")
	}

	// Parse sample input - failed build
	errorInput := `main.go:10:5: undefined: foo
main.go:15:10: cannot use x (type int) as type string
`
	result, err = parser.Parse(strings.NewReader(errorInput))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	build, ok = result.Data.(*golang.Build)
	if !ok {
		t.Fatalf("result.Data is not *golang.Build, got %T", result.Data)
	}

	if build.Success {
		t.Error("Success = true, want false for error output")
	}
	if len(build.Errors) != 2 {
		t.Errorf("len(Errors) = %d, want 2", len(build.Errors))
	}
}
