package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Compile-time check that SchemaValidator implements ports.SchemaValidator.
var _ ports.SchemaValidator = (*SchemaValidator)(nil)

func TestSchemaValidator_Validate(t *testing.T) {
	t.Run("valid data passes validation", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.NewSchema(
			"https://example.com/test.json",
			"Test Schema",
			"object",
			map[string]domain.PropertySchema{
				"name":    {Type: "string", Description: "The name"},
				"count":   {Type: "integer", Description: "A count"},
				"enabled": {Type: "boolean", Description: "Is enabled"},
			},
			[]string{"name"},
		)

		data := map[string]any{
			"name":    "test",
			"count":   42,
			"enabled": true,
		}

		// Act
		err := validator.Validate(data, schema)

		// Assert
		require.NoError(t, err)
	})

	t.Run("missing required field fails validation", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.NewSchema(
			"https://example.com/test.json",
			"Test Schema",
			"object",
			map[string]domain.PropertySchema{
				"name":  {Type: "string", Description: "The name"},
				"count": {Type: "integer", Description: "A count"},
			},
			[]string{"name", "count"},
		)

		data := map[string]any{
			"name": "test",
			// missing "count"
		}

		// Act
		err := validator.Validate(data, schema)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrSchemaValidation)
		assert.Contains(t, err.Error(), "count")
	})

	t.Run("wrong type fails validation", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.NewSchema(
			"https://example.com/test.json",
			"Test Schema",
			"object",
			map[string]domain.PropertySchema{
				"name":  {Type: "string", Description: "The name"},
				"count": {Type: "integer", Description: "A count"},
			},
			[]string{"name"},
		)

		data := map[string]any{
			"name":  "test",
			"count": "not-an-integer", // wrong type
		}

		// Act
		err := validator.Validate(data, schema)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrSchemaValidation)
		assert.Contains(t, err.Error(), "count")
	})

	t.Run("nil data fails validation for object schema", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.NewSchema(
			"https://example.com/test.json",
			"Test Schema",
			"object",
			map[string]domain.PropertySchema{
				"name": {Type: "string", Description: "The name"},
			},
			[]string{"name"},
		)

		// Act
		err := validator.Validate(nil, schema)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrSchemaValidation)
	})

	t.Run("struct with matching fields passes validation", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.NewSchema(
			"https://example.com/test.json",
			"Test Schema",
			"object",
			map[string]domain.PropertySchema{
				"name":    {Type: "string", Description: "The name"},
				"success": {Type: "boolean", Description: "Success flag"},
			},
			[]string{"name", "success"},
		)

		type TestResult struct {
			Name    string `json:"name"`
			Success bool   `json:"success"`
		}
		data := &TestResult{
			Name:    "test",
			Success: true,
		}

		// Act
		err := validator.Validate(data, schema)

		// Assert
		require.NoError(t, err)
	})

	t.Run("empty schema passes any data", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.Schema{} // empty schema

		data := map[string]any{
			"anything": "goes",
		}

		// Act
		err := validator.Validate(data, schema)

		// Assert
		require.NoError(t, err)
	})

	t.Run("array type validates array data", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.NewSchema(
			"https://example.com/test.json",
			"Test Array Schema",
			"array",
			nil, // no properties for array type
			nil,
		)

		data := []string{"item1", "item2"}

		// Act
		err := validator.Validate(data, schema)

		// Assert
		require.NoError(t, err)
	})

	t.Run("non-array data fails array schema", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.NewSchema(
			"https://example.com/test.json",
			"Test Array Schema",
			"array",
			nil,
			nil,
		)

		data := map[string]any{"not": "array"}

		// Act
		err := validator.Validate(data, schema)

		// Assert
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrSchemaValidation)
	})

	t.Run("number type accepts float64", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.NewSchema(
			"https://example.com/test.json",
			"Test Schema",
			"object",
			map[string]domain.PropertySchema{
				"value": {Type: "number", Description: "A number"},
			},
			[]string{"value"},
		)

		data := map[string]any{
			"value": 3.14,
		}

		// Act
		err := validator.Validate(data, schema)

		// Assert
		require.NoError(t, err)
	})

	t.Run("integer type accepts int", func(t *testing.T) {
		// Arrange
		validator := NewSchemaValidator()
		schema := domain.NewSchema(
			"https://example.com/test.json",
			"Test Schema",
			"object",
			map[string]domain.PropertySchema{
				"count": {Type: "integer", Description: "A count"},
			},
			[]string{"count"},
		)

		data := map[string]any{
			"count": 42,
		}

		// Act
		err := validator.Validate(data, schema)

		// Assert
		require.NoError(t, err)
	})
}
