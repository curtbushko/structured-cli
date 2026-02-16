// Package schema provides implementations of the SchemaRepository port.
// This package is in the adapters layer and implements the ports.SchemaRepository
// interface using Go's embed.FS for embedded schema files.
package schema

import (
	"embed"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// ErrSchemaNotFound is returned when a schema cannot be found by name.
var ErrSchemaNotFound = errors.New("schema not found")

// Compile-time check that EmbeddedSchemaRepository implements ports.SchemaRepository.
var _ ports.SchemaRepository = (*EmbeddedSchemaRepository)(nil)

// EmbeddedSchemaRepository implements ports.SchemaRepository using embed.FS.
// It loads JSON Schema files from an embedded filesystem, providing
// access to schemas without external file dependencies.
type EmbeddedSchemaRepository struct {
	fs      embed.FS
	baseDir string
	schemas map[string]domain.Schema
}

// NewEmbeddedSchemaRepository creates a new EmbeddedSchemaRepository.
// The fs parameter should be an embedded filesystem containing JSON schema files.
// The baseDir parameter specifies the directory within the embedded FS to search.
func NewEmbeddedSchemaRepository(fs embed.FS, baseDir string) *EmbeddedSchemaRepository {
	repo := &EmbeddedSchemaRepository{
		fs:      fs,
		baseDir: baseDir,
		schemas: make(map[string]domain.Schema),
	}
	repo.loadSchemas()
	return repo
}

// loadSchemas reads all JSON schema files from the embedded filesystem
// and parses them into Schema objects.
func (r *EmbeddedSchemaRepository) loadSchemas() {
	entries, err := r.fs.ReadDir(r.baseDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}

		filePath := path.Join(r.baseDir, name)
		data, err := r.fs.ReadFile(filePath)
		if err != nil {
			continue
		}

		schema, err := domain.SchemaFromJSON(data)
		if err != nil {
			continue
		}

		// Remove .json extension for the schema name
		schemaName := strings.TrimSuffix(name, ".json")
		r.schemas[schemaName] = schema
	}
}

// Get retrieves a schema by name.
// The name should not include the .json extension.
// Returns ErrSchemaNotFound if the schema does not exist.
func (r *EmbeddedSchemaRepository) Get(name string) (domain.Schema, error) {
	schema, ok := r.schemas[name]
	if !ok {
		return domain.Schema{}, fmt.Errorf("%w: %s", ErrSchemaNotFound, name)
	}
	return schema, nil
}

// List returns the names of all available schemas.
// Names do not include the .json extension.
func (r *EmbeddedSchemaRepository) List() []string {
	names := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		names = append(names, name)
	}
	return names
}
