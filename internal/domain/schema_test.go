package domain

import (
	"reflect"
	"testing"
)

func TestSchemaFromJSON(t *testing.T) {
	tests := []struct {
		name      string
		jsonBytes []byte
		wantID    string
		wantTitle string
		wantType  string
		wantProps map[string]PropertySchema
		wantReq   []string
		wantErr   bool
	}{
		{
			name: "valid schema with all fields",
			jsonBytes: []byte(`{
				"$id": "git-status",
				"title": "Git Status Output",
				"type": "object",
				"properties": {
					"branch": {"type": "string"},
					"clean": {"type": "boolean"}
				},
				"required": ["branch"]
			}`),
			wantID:    "git-status",
			wantTitle: "Git Status Output",
			wantType:  "object",
			wantProps: map[string]PropertySchema{
				"branch": {Type: "string"},
				"clean":  {Type: "boolean"},
			},
			wantReq: []string{"branch"},
			wantErr: false,
		},
		{
			name: "valid schema with minimal fields",
			jsonBytes: []byte(`{
				"type": "object"
			}`),
			wantID:    "",
			wantTitle: "",
			wantType:  "object",
			wantProps: nil,
			wantReq:   nil,
			wantErr:   false,
		},
		{
			name: "array schema",
			jsonBytes: []byte(`{
				"$id": "file-list",
				"title": "File List",
				"type": "array",
				"items": {"type": "string"}
			}`),
			wantID:    "file-list",
			wantTitle: "File List",
			wantType:  "array",
			wantProps: nil,
			wantReq:   nil,
			wantErr:   false,
		},
		{
			name:      "invalid JSON",
			jsonBytes: []byte(`{invalid json}`),
			wantErr:   true,
		},
		{
			name:      "empty JSON",
			jsonBytes: []byte(``),
			wantErr:   true,
		},
		{
			name:      "nil bytes",
			jsonBytes: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SchemaFromJSON(tt.jsonBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("SchemaFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.ID != tt.wantID {
				t.Errorf("Schema.ID = %v, want %v", got.ID, tt.wantID)
			}
			if got.Title != tt.wantTitle {
				t.Errorf("Schema.Title = %v, want %v", got.Title, tt.wantTitle)
			}
			if got.Type != tt.wantType {
				t.Errorf("Schema.Type = %v, want %v", got.Type, tt.wantType)
			}
			if !reflect.DeepEqual(got.Properties, tt.wantProps) {
				t.Errorf("Schema.Properties = %v, want %v", got.Properties, tt.wantProps)
			}
			if !reflect.DeepEqual(got.Required, tt.wantReq) {
				t.Errorf("Schema.Required = %v, want %v", got.Required, tt.wantReq)
			}
		})
	}
}

func TestSchema_Raw(t *testing.T) {
	jsonBytes := []byte(`{"$id": "test", "type": "object"}`)
	schema, err := SchemaFromJSON(jsonBytes)
	if err != nil {
		t.Fatalf("SchemaFromJSON() error = %v", err)
	}

	raw := schema.Raw()
	if !reflect.DeepEqual(raw, jsonBytes) {
		t.Errorf("Schema.Raw() = %s, want %s", raw, jsonBytes)
	}
}

func TestSchema_ZeroValue(t *testing.T) {
	// Test that zero value is usable
	var schema Schema

	if schema.ID != "" {
		t.Errorf("Zero value ID should be empty, got %v", schema.ID)
	}
	if schema.Title != "" {
		t.Errorf("Zero value Title should be empty, got %v", schema.Title)
	}
	if schema.Type != "" {
		t.Errorf("Zero value Type should be empty, got %v", schema.Type)
	}
	if schema.Properties != nil {
		t.Errorf("Zero value Properties should be nil, got %v", schema.Properties)
	}
	if schema.Required != nil {
		t.Errorf("Zero value Required should be nil, got %v", schema.Required)
	}
	if schema.Raw() != nil {
		t.Errorf("Zero value Raw() should return nil, got %v", schema.Raw())
	}
}

func TestPropertySchema(t *testing.T) {
	tests := []struct {
		name     string
		propType string
		propDesc string
	}{
		{
			name:     "string property",
			propType: "string",
			propDesc: "A string value",
		},
		{
			name:     "integer property",
			propType: "integer",
			propDesc: "",
		},
		{
			name:     "boolean property",
			propType: "boolean",
			propDesc: "A boolean flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prop := PropertySchema{
				Type:        tt.propType,
				Description: tt.propDesc,
			}
			if prop.Type != tt.propType {
				t.Errorf("PropertySchema.Type = %v, want %v", prop.Type, tt.propType)
			}
			if prop.Description != tt.propDesc {
				t.Errorf("PropertySchema.Description = %v, want %v", prop.Description, tt.propDesc)
			}
		})
	}
}

func TestNewSchema(t *testing.T) {
	props := map[string]PropertySchema{
		"name": {Type: "string"},
	}
	required := []string{"name"}

	schema := NewSchema("test-id", "Test Schema", "object", props, required)

	if schema.ID != "test-id" {
		t.Errorf("Schema.ID = %v, want test-id", schema.ID)
	}
	if schema.Title != "Test Schema" {
		t.Errorf("Schema.Title = %v, want Test Schema", schema.Title)
	}
	if schema.Type != "object" {
		t.Errorf("Schema.Type = %v, want object", schema.Type)
	}
	if !reflect.DeepEqual(schema.Properties, props) {
		t.Errorf("Schema.Properties = %v, want %v", schema.Properties, props)
	}
	if !reflect.DeepEqual(schema.Required, required) {
		t.Errorf("Schema.Required = %v, want %v", schema.Required, required)
	}
}
