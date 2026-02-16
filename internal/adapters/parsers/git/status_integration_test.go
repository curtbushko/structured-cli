package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
	"github.com/curtbushko/structured-cli/internal/application"
)

func TestStatusParser_RegistryIntegration(t *testing.T) {
	t.Run("parser can be registered and found", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(git.NewStatusParser())

		parser, found := registry.Find("git", []string{"status"})

		if !found {
			t.Fatal("expected parser to be found")
		}
		if parser == nil {
			t.Fatal("parser is nil")
		}
	})

	t.Run("parser is not found for non-matching commands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(git.NewStatusParser())

		_, found := registry.Find("git", []string{"log"})

		if found {
			t.Error("expected parser not to be found for git log")
		}
	})

	t.Run("parser is not found for non-git commands", func(t *testing.T) {
		registry := application.NewInMemoryParserRegistry()
		registry.Register(git.NewStatusParser())

		_, found := registry.Find("docker", []string{"status"})

		if found {
			t.Error("expected parser not to be found for docker status")
		}
	})
}

func TestStatusParser_SchemaValidation(t *testing.T) {
	parser := git.NewStatusParser()
	schema := parser.Schema()

	t.Run("schema has correct title", func(t *testing.T) {
		if schema.Title != "Git Status Output" {
			t.Errorf("Schema.Title = %q, want %q", schema.Title, "Git Status Output")
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

	t.Run("schema requires branch field", func(t *testing.T) {
		hasBranch := false
		for _, req := range schema.Required {
			if req == "branch" {
				hasBranch = true
				break
			}
		}
		if !hasBranch {
			t.Error("Schema.Required does not contain 'branch'")
		}
	})

	t.Run("schema has expected properties", func(t *testing.T) {
		expectedProps := []string{"branch", "upstream", "ahead", "behind", "staged", "modified", "deleted", "untracked", "conflicts", "clean"}
		for _, prop := range expectedProps {
			if _, ok := schema.Properties[prop]; !ok {
				t.Errorf("Schema.Properties missing %q", prop)
			}
		}
	})
}

func TestStatusParser_EndToEndFlow(t *testing.T) {
	// This test verifies the complete flow:
	// 1. Parser is registered
	// 2. Registry can find it
	// 3. Parser can parse input
	// 4. Result contains expected data

	registry := application.NewInMemoryParserRegistry()
	registry.Register(git.NewStatusParser())

	// Find the parser
	parser, found := registry.Find("git", []string{"status"})
	if !found {
		t.Fatal("parser not found in registry")
	}

	// Parse sample input
	input := `# branch.oid abc123
# branch.head feature-branch
# branch.upstream origin/feature-branch
# branch.ab +2 -1
1 M. N... 100644 100644 100644 abc1234 def5678 modified-file.go
? untracked-file.txt
`
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Verify the result
	status, ok := result.Data.(*git.Status)
	if !ok {
		t.Fatalf("result.Data is not *git.Status, got %T", result.Data)
	}

	if status.Branch != "feature-branch" {
		t.Errorf("Branch = %q, want %q", status.Branch, "feature-branch")
	}
	if status.Upstream == nil || *status.Upstream != "origin/feature-branch" {
		t.Errorf("Upstream = %v, want %q", status.Upstream, "origin/feature-branch")
	}
	if status.Ahead != 2 {
		t.Errorf("Ahead = %d, want 2", status.Ahead)
	}
	if status.Behind != 1 {
		t.Errorf("Behind = %d, want 1", status.Behind)
	}
	if len(status.Staged) != 1 {
		t.Errorf("len(Staged) = %d, want 1", len(status.Staged))
	}
	if len(status.Untracked) != 1 {
		t.Errorf("len(Untracked) = %d, want 1", len(status.Untracked))
	}
	if status.Clean {
		t.Error("Clean = true, want false")
	}
}
