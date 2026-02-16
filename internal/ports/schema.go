package ports

import (
	"github.com/curtbushko/structured-cli/internal/domain"
)

// SchemaRepository provides access to JSON schemas.
// Schemas define the structure of parsed CLI output and are used
// for validation and documentation.
//
// Implementations may load schemas from embedded files, filesystem,
// or remote sources.
type SchemaRepository interface {
	// Get retrieves a schema by name.
	// The name typically matches the command pattern (e.g., "git-status").
	// Returns ErrSchemaNotFound if the schema does not exist.
	Get(name string) (domain.Schema, error)

	// List returns the names of all available schemas.
	// Useful for discovering supported commands or building documentation.
	List() []string
}
