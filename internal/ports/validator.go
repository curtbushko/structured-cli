package ports

import (
	"github.com/curtbushko/structured-cli/internal/domain"
)

// SchemaValidator validates data against a JSON schema.
// It ensures that parsed output conforms to the expected structure.
//
// Implementations may use different validation libraries or strategies.
type SchemaValidator interface {
	// Validate checks that data conforms to the given schema.
	// Returns nil if validation passes, or a ValidationError if it fails.
	Validate(data any, schema domain.Schema) error
}
