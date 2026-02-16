package npm

import (
	"strings"
	"testing"
)

func TestInitParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData InitResult
	}{
		{
			name:  "empty output",
			input: "",
			wantData: InitResult{
				Success:  true,
				Keywords: []string{},
			},
		},
		{
			name: "successful init",
			input: `Wrote to /path/to/project/package.json:

{
  "name": "myproject",
  "version": "1.0.0",
  "description": "My awesome project",
  "main": "index.js",
  "scripts": {
    "test": "jest"
  },
  "keywords": ["cli", "tool"],
  "author": "John Doe",
  "license": "MIT"
}
`,
			wantData: InitResult{
				Success:     true,
				PackageName: "myproject",
				Version:     "1.0.0",
				Description: "My awesome project",
				EntryPoint:  "index.js",
				TestCommand: "jest",
				Author:      "John Doe",
				License:     "MIT",
				Keywords:    []string{"cli", "tool"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewInitParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*InitResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *InitResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("InitResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.PackageName != tt.wantData.PackageName {
				t.Errorf("InitResult.PackageName = %v, want %v", got.PackageName, tt.wantData.PackageName)
			}

			if got.Version != tt.wantData.Version {
				t.Errorf("InitResult.Version = %v, want %v", got.Version, tt.wantData.Version)
			}

			if got.Description != tt.wantData.Description {
				t.Errorf("InitResult.Description = %v, want %v", got.Description, tt.wantData.Description)
			}

			if got.Author != tt.wantData.Author {
				t.Errorf("InitResult.Author = %v, want %v", got.Author, tt.wantData.Author)
			}

			if got.License != tt.wantData.License {
				t.Errorf("InitResult.License = %v, want %v", got.License, tt.wantData.License)
			}
		})
	}
}

func TestInitParser_WithError(t *testing.T) {
	input := `npm ERR! code ENOENT
npm ERR! syscall open
npm ERR! path /path/to/project/package.json
npm ERR! errno -2
`

	parser := NewInitParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*InitResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *InitResult", result.Data)
	}

	if got.Success {
		t.Error("InitResult.Success = true, want false when init failed")
	}
}

func TestInitParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches npm init",
			cmd:         "npm",
			subcommands: []string{"init"},
			want:        true,
		},
		{
			name:        "matches npm init -y",
			cmd:         "npm",
			subcommands: []string{"init", "-y"},
			want:        true,
		},
		{
			name:        "matches npm create",
			cmd:         "npm",
			subcommands: []string{"create"},
			want:        true,
		},
		{
			name:        "matches npm innit",
			cmd:         "npm",
			subcommands: []string{"innit"},
			want:        true,
		},
		{
			name:        "does not match npm install",
			cmd:         "npm",
			subcommands: []string{"install"},
			want:        false,
		},
		{
			name:        "does not match yarn init",
			cmd:         "yarn",
			subcommands: []string{"init"},
			want:        false,
		},
		{
			name:        "does not match empty",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewInitParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestInitParser_Schema(t *testing.T) {
	parser := NewInitParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema.ID should not be empty")
	}

	if schema.Title == "" {
		t.Error("Schema.Title should not be empty")
	}

	if schema.Type != "object" {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, "object")
	}

	requiredProps := []string{"success", "package_name", "version"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
