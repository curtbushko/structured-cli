// Package schema provides implementations of the SchemaRepository port.
package schema

import (
	"embed"
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/ports"
)

//go:embed testdata/*.json
var testSchemaFS embed.FS

func TestEmbeddedRepo_Get_Exists(t *testing.T) {
	// Arrange
	repo := NewEmbeddedSchemaRepository(testSchemaFS, "testdata")

	// Act
	schema, err := repo.Get("test-schema")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if schema.ID != "https://structured-cli.dev/schemas/test-schema.json" {
		t.Errorf("schema.ID = %q, want %q", schema.ID, "https://structured-cli.dev/schemas/test-schema.json")
	}

	if schema.Title != "Test Schema" {
		t.Errorf("schema.Title = %q, want %q", schema.Title, "Test Schema")
	}

	if schema.Type != "object" {
		t.Errorf("schema.Type = %q, want %q", schema.Type, "object")
	}

	if len(schema.Properties) != 2 {
		t.Errorf("len(schema.Properties) = %d, want 2", len(schema.Properties))
	}

	if _, ok := schema.Properties["name"]; !ok {
		t.Error("schema.Properties should contain 'name'")
	}

	if _, ok := schema.Properties["value"]; !ok {
		t.Error("schema.Properties should contain 'value'")
	}

	if len(schema.Required) != 1 || schema.Required[0] != "name" {
		t.Errorf("schema.Required = %v, want [\"name\"]", schema.Required)
	}

	if schema.Raw() == nil {
		t.Error("schema.Raw() should not be nil")
	}
}

func TestEmbeddedRepo_Get_NotExists(t *testing.T) {
	// Arrange
	repo := NewEmbeddedSchemaRepository(testSchemaFS, "testdata")

	// Act
	_, err := repo.Get("nonexistent")

	// Assert
	if err == nil {
		t.Fatal("expected error for nonexistent schema, got nil")
	}
}

func TestEmbeddedRepo_List(t *testing.T) {
	// Arrange
	repo := NewEmbeddedSchemaRepository(testSchemaFS, "testdata")

	// Act
	names := repo.List()

	// Assert
	if len(names) == 0 {
		t.Fatal("expected at least one schema name, got empty list")
	}

	found := false
	for _, name := range names {
		if name == "test-schema" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("List() = %v, want to contain 'test-schema'", names)
	}
}

func TestEmbeddedRepo_SchemaValid(t *testing.T) {
	// Arrange
	repo := NewEmbeddedSchemaRepository(testSchemaFS, "testdata")

	// Act
	schema, err := repo.Get("test-schema")
	if err != nil {
		t.Fatalf("failed to get schema: %v", err)
	}

	// Assert - Check schema has valid JSON Schema structure
	if schema.Type == "" {
		t.Error("schema.Type should not be empty")
	}

	if schema.Properties == nil {
		t.Error("schema.Properties should not be nil")
	}

	// Verify raw JSON contains $schema field
	raw := schema.Raw()
	if raw == nil {
		t.Fatal("schema.Raw() should not be nil")
	}

	rawStr := string(raw)
	if !strings.Contains(rawStr, `"$schema"`) {
		t.Error("raw JSON should contain $schema field")
	}
	if !strings.Contains(rawStr, `"type"`) {
		t.Error("raw JSON should contain type field")
	}
	if !strings.Contains(rawStr, `"properties"`) {
		t.Error("raw JSON should contain properties field")
	}
}

func TestEmbeddedRepo_ImplementsSchemaRepository(t *testing.T) {
	// This test verifies at compile time that EmbeddedSchemaRepository
	// implements the ports.SchemaRepository interface.
	var repo ports.SchemaRepository = NewEmbeddedSchemaRepository(testSchemaFS, "testdata")

	// Verify it returns domain.Schema
	schema, err := repo.Get("test-schema")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify schema is the correct type by using it
	if schema.ID == "" {
		t.Error("expected schema.ID to be set")
	}
}
