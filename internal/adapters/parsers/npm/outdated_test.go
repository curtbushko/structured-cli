package npm

import (
	"strings"
	"testing"
)

func TestOutdatedParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData OutdatedResult
	}{
		{
			name:  "empty output means all up to date",
			input: "",
			wantData: OutdatedResult{
				Success:  true,
				Packages: []OutdatedPackage{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewOutdatedParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*OutdatedResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *OutdatedResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("OutdatedResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Packages) != len(tt.wantData.Packages) {
				t.Errorf("OutdatedResult.Packages length = %d, want %d", len(got.Packages), len(tt.wantData.Packages))
			}
		})
	}
}

func TestOutdatedParser_WithOutdatedPackages(t *testing.T) {
	// npm outdated output format
	input := `Package   Current  Wanted  Latest  Location              Depended by
lodash    4.17.19  4.17.21 4.17.21 node_modules/lodash   myproject
express   4.17.0   4.18.2  4.18.2  node_modules/express  myproject
`

	parser := NewOutdatedParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*OutdatedResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *OutdatedResult", result.Data)
	}

	if got.Success {
		t.Error("OutdatedResult.Success = true, want false when packages are outdated")
	}

	if len(got.Packages) != 2 {
		t.Fatalf("OutdatedResult.Packages length = %d, want 2", len(got.Packages))
	}

	// Check first package
	wantPkg := OutdatedPackage{
		Name:     "lodash",
		Current:  "4.17.19",
		Wanted:   "4.17.21",
		Latest:   "4.17.21",
		Location: "node_modules/lodash",
	}

	if got.Packages[0].Name != wantPkg.Name {
		t.Errorf("Package[0].Name = %q, want %q", got.Packages[0].Name, wantPkg.Name)
	}

	if got.Packages[0].Current != wantPkg.Current {
		t.Errorf("Package[0].Current = %q, want %q", got.Packages[0].Current, wantPkg.Current)
	}

	if got.Packages[0].Wanted != wantPkg.Wanted {
		t.Errorf("Package[0].Wanted = %q, want %q", got.Packages[0].Wanted, wantPkg.Wanted)
	}

	if got.Packages[0].Latest != wantPkg.Latest {
		t.Errorf("Package[0].Latest = %q, want %q", got.Packages[0].Latest, wantPkg.Latest)
	}

	if got.Packages[0].Location != wantPkg.Location {
		t.Errorf("Package[0].Location = %q, want %q", got.Packages[0].Location, wantPkg.Location)
	}
}

func TestOutdatedParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches npm outdated",
			cmd:         "npm",
			subcommands: []string{"outdated"},
			want:        true,
		},
		{
			name:        "does not match npm install",
			cmd:         "npm",
			subcommands: []string{"install"},
			want:        false,
		},
		{
			name:        "does not match yarn outdated",
			cmd:         "yarn",
			subcommands: []string{"outdated"},
			want:        false,
		},
		{
			name:        "does not match empty",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewOutdatedParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestOutdatedParser_Schema(t *testing.T) {
	parser := NewOutdatedParser()
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

	requiredProps := []string{"success", "packages"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
