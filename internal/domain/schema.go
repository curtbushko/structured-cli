package domain

import (
	"encoding/json"
	"fmt"
)

// Schema represents a JSON Schema that defines the structure of parsed CLI output.
// It contains the essential fields needed to describe and validate structured data.
type Schema struct {
	// ID is the unique identifier for the schema (JSON Schema $id)
	ID string

	// Title is the human-readable name for the schema
	Title string

	// Type is the JSON Schema type (object, array, string, etc.)
	Type string

	// Properties defines the schema for each property (for object type)
	Properties map[string]PropertySchema

	// Required lists the required property names (for object type)
	Required []string

	// raw stores the original JSON bytes for the Raw() method
	raw []byte
}

// PropertySchema defines the schema for a single property within an object schema.
type PropertySchema struct {
	// Type is the JSON Schema type for this property
	Type string `json:"type"`

	// Description provides documentation for this property
	Description string `json:"description,omitempty"`
}

// jsonSchema is used for JSON unmarshaling with proper field mapping.
type jsonSchema struct {
	ID         string                    `json:"$id"`
	Title      string                    `json:"title"`
	Type       string                    `json:"type"`
	Properties map[string]PropertySchema `json:"properties"`
	Required   []string                  `json:"required"`
}

// NewSchema creates a new Schema with the given fields.
func NewSchema(id, title, schemaType string, properties map[string]PropertySchema, required []string) Schema {
	return Schema{
		ID:         id,
		Title:      title,
		Type:       schemaType,
		Properties: properties,
		Required:   required,
	}
}

// SchemaFromJSON parses JSON bytes into a Schema.
// Returns ErrInvalidSchema if the JSON is invalid or cannot be parsed.
func SchemaFromJSON(data []byte) (Schema, error) {
	if len(data) == 0 {
		return Schema{}, fmt.Errorf("%w: empty input", ErrInvalidSchema)
	}

	var js jsonSchema
	if err := json.Unmarshal(data, &js); err != nil {
		return Schema{}, fmt.Errorf("%w: %w", ErrInvalidSchema, err)
	}

	return Schema{
		ID:         js.ID,
		Title:      js.Title,
		Type:       js.Type,
		Properties: js.Properties,
		Required:   js.Required,
		raw:        data,
	}, nil
}

// Raw returns the original JSON bytes used to create the schema.
// Returns nil if the schema was not created from JSON.
func (s Schema) Raw() []byte {
	return s.raw
}
