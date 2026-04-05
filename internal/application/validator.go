// Package application contains the business logic and use cases for structured-cli.
package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// JSON Schema type constants.
const (
	schemaTypeObject  = "object"
	schemaTypeArray   = "array"
	schemaTypeString  = "string"
	schemaTypeNumber  = "number"
	schemaTypeInteger = "integer"
	schemaTypeBoolean = "boolean"
)

// Compile-time check that SchemaValidator implements ports.SchemaValidator.
var _ ports.SchemaValidator = (*SchemaValidator)(nil)

// SchemaValidator validates parsed data against JSON schemas.
// It provides basic type checking and required field validation
// based on the domain.Schema type.
type SchemaValidator struct{}

// NewSchemaValidator creates a new SchemaValidator.
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

// Validate checks that data conforms to the given schema.
// Returns nil if validation passes, or a wrapped ErrSchemaValidation if it fails.
//
// The validator performs:
// - Type checking (object, array, string, number, integer, boolean)
// - Required field validation for object types
// - Property type validation for object fields
func (v *SchemaValidator) Validate(data any, schema domain.Schema) error {
	// Empty schema accepts any data
	if schema.Type == "" && schema.ID == "" {
		return nil
	}

	// Check top-level type
	if err := v.validateType(data, schema.Type); err != nil {
		return errors.Join(domain.ErrSchemaValidation, err)
	}

	// For object types, validate properties and required fields
	if schema.Type == schemaTypeObject {
		return v.validateObject(data, schema)
	}

	return nil
}

// validateType checks if data matches the expected JSON Schema type.
func (v *SchemaValidator) validateType(data any, schemaType string) error {
	if schemaType == "" {
		return nil
	}

	if data == nil {
		return fmt.Errorf("expected %s, got nil", schemaType)
	}

	dataMap, isMap := v.toMap(data)

	switch schemaType {
	case schemaTypeObject:
		if !isMap {
			return fmt.Errorf("expected object, got %T", data)
		}
	case schemaTypeArray:
		if !v.isArray(data) {
			return fmt.Errorf("expected array, got %T", data)
		}
	case schemaTypeString:
		if _, ok := data.(string); !ok {
			return fmt.Errorf("expected string, got %T", data)
		}
	case schemaTypeNumber:
		if !v.isNumber(data) {
			return fmt.Errorf("expected number, got %T", data)
		}
	case schemaTypeInteger:
		if !v.isInteger(data) {
			return fmt.Errorf("expected integer, got %T", data)
		}
	case schemaTypeBoolean:
		if _, ok := data.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", data)
		}
	}

	// If it's a map/object, we already checked it
	_ = dataMap

	return nil
}

// validateObject validates an object against schema properties and required fields.
func (v *SchemaValidator) validateObject(data any, schema domain.Schema) error {
	dataMap, ok := v.toMap(data)
	if !ok {
		return fmt.Errorf("%w: expected object, got %T", domain.ErrSchemaValidation, data)
	}

	// Check required fields
	for _, required := range schema.Required {
		if _, exists := dataMap[required]; !exists {
			return fmt.Errorf("%w: missing required field: %s", domain.ErrSchemaValidation, required)
		}
	}

	// Validate property types
	for propName, propSchema := range schema.Properties {
		propValue, exists := dataMap[propName]
		if !exists {
			continue // Optional field not present is OK
		}

		if err := v.validatePropertyType(propValue, propSchema.Type, propName); err != nil {
			return errors.Join(domain.ErrSchemaValidation, err)
		}
	}

	return nil
}

// validatePropertyType validates a single property value against its expected type.
func (v *SchemaValidator) validatePropertyType(value any, expectedType, propName string) error {
	if value == nil {
		// null is valid for any type unless we add null checking
		return nil
	}

	switch expectedType {
	case schemaTypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("field %s: expected string, got %T", propName, value)
		}
	case schemaTypeNumber:
		if !v.isNumber(value) {
			return fmt.Errorf("field %s: expected number, got %T", propName, value)
		}
	case schemaTypeInteger:
		if !v.isInteger(value) {
			return fmt.Errorf("field %s: expected integer, got %T", propName, value)
		}
	case schemaTypeBoolean:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("field %s: expected boolean, got %T", propName, value)
		}
	case schemaTypeArray:
		if !v.isArray(value) {
			return fmt.Errorf("field %s: expected array, got %T", propName, value)
		}
	case schemaTypeObject:
		if _, ok := v.toMap(value); !ok {
			return fmt.Errorf("field %s: expected object, got %T", propName, value)
		}
	}

	return nil
}

// toMap converts data to a map[string]any.
// It handles both map types and structs (via JSON marshaling).
func (v *SchemaValidator) toMap(data any) (map[string]any, bool) {
	if data == nil {
		return nil, false
	}

	// Direct map
	if m, ok := data.(map[string]any); ok {
		return m, true
	}

	// Try reflection for struct types
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, false
		}
		val = val.Elem()
	}

	if val.Kind() == reflect.Struct {
		// Convert struct to map via JSON for consistent field naming
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, false
		}

		var m map[string]any
		if err := json.Unmarshal(jsonBytes, &m); err != nil {
			return nil, false
		}
		return m, true
	}

	return nil, false
}

// isArray checks if data is an array or slice type.
func (v *SchemaValidator) isArray(data any) bool {
	if data == nil {
		return false
	}

	val := reflect.ValueOf(data)
	kind := val.Kind()
	return kind == reflect.Array || kind == reflect.Slice
}

// isNumber checks if data is a numeric type (int, float, etc.).
func (v *SchemaValidator) isNumber(data any) bool {
	if data == nil {
		return false
	}

	switch data.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	case json.Number:
		return true
	}

	return false
}

// isInteger checks if data is an integer type or a whole number.
func (v *SchemaValidator) isInteger(data any) bool {
	if data == nil {
		return false
	}

	switch v := data.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32:
		return float32(int(v)) == v
	case float64:
		return float64(int(v)) == v
	case json.Number:
		_, err := v.Int64()
		return err == nil
	}

	return false
}
