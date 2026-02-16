package golang

import (
	"encoding/json"
	"reflect"
	"testing"
)

// Build Tests

func TestBuildType(t *testing.T) {
	t.Run("can be instantiated with all fields", func(t *testing.T) {
		build := Build{
			Success:  true,
			Packages: []string{"./cmd/...", "./internal/..."},
			Errors:   []BuildError{},
		}

		if !build.Success {
			t.Errorf("Success = %v, want true", build.Success)
		}
		if len(build.Packages) != 2 {
			t.Errorf("Packages length = %v, want 2", len(build.Packages))
		}
		if len(build.Errors) != 0 {
			t.Errorf("Errors length = %v, want 0", len(build.Errors))
		}
	})

	t.Run("failed build with errors", func(t *testing.T) {
		build := Build{
			Success:  false,
			Packages: []string{"./cmd/..."},
			Errors: []BuildError{
				{
					File:    "main.go",
					Line:    10,
					Column:  5,
					Message: "undefined: foo",
				},
			},
		}

		if build.Success {
			t.Errorf("Success = %v, want false", build.Success)
		}
		if len(build.Packages) != 1 {
			t.Errorf("Packages length = %v, want 1", len(build.Packages))
		}
		if len(build.Errors) != 1 {
			t.Errorf("Errors length = %v, want 1", len(build.Errors))
		}
		if build.Errors[0].File != "main.go" {
			t.Errorf("Error File = %v, want main.go", build.Errors[0].File)
		}
		if build.Errors[0].Line != 10 {
			t.Errorf("Error Line = %v, want 10", build.Errors[0].Line)
		}
		if build.Errors[0].Column != 5 {
			t.Errorf("Error Column = %v, want 5", build.Errors[0].Column)
		}
		if build.Errors[0].Message != "undefined: foo" {
			t.Errorf("Error Message = %v, want undefined: foo", build.Errors[0].Message)
		}
	})
}

func TestBuildErrorType(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		line    int
		column  int
		message string
	}{
		{"syntax error", "parser.go", 25, 10, "syntax error: unexpected token"},
		{"type error", "service.go", 100, 1, "cannot use int as string"},
		{"import error", "main.go", 5, 2, "could not import: pkg/foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			be := BuildError{
				File:    tt.file,
				Line:    tt.line,
				Column:  tt.column,
				Message: tt.message,
			}

			if be.File != tt.file {
				t.Errorf("File = %v, want %v", be.File, tt.file)
			}
			if be.Line != tt.line {
				t.Errorf("Line = %v, want %v", be.Line, tt.line)
			}
			if be.Column != tt.column {
				t.Errorf("Column = %v, want %v", be.Column, tt.column)
			}
			if be.Message != tt.message {
				t.Errorf("Message = %v, want %v", be.Message, tt.message)
			}
		})
	}
}

func TestBuildJSONMarshal(t *testing.T) {
	build := Build{
		Success:  false,
		Packages: []string{"./cmd/..."},
		Errors: []BuildError{
			{File: "main.go", Line: 10, Column: 5, Message: "undefined: foo"},
		},
	}

	data, err := json.Marshal(build)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify JSON structure matches expected property names
	expectedKeys := []string{"success", "packages", "errors"}
	for _, key := range expectedKeys {
		if _, exists := unmarshaled[key]; !exists {
			t.Errorf("Missing expected key in JSON: %s", key)
		}
	}

	// Verify values
	if unmarshaled["success"] != false {
		t.Errorf("success = %v, want false", unmarshaled["success"])
	}
}

func TestBuildTypeMatchesExpected(t *testing.T) {
	build := Build{}
	v := reflect.TypeOf(build)

	expectedFields := map[string]string{
		"Success":  "bool",
		"Packages": "[]string",
		"Errors":   "[]golang.BuildError",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

func TestBuildErrorTypeMatchesExpected(t *testing.T) {
	be := BuildError{}
	v := reflect.TypeOf(be)

	expectedFields := map[string]string{
		"File":    "string",
		"Line":    "int",
		"Column":  "int",
		"Message": "string",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

// Test Tests

func TestTestResultType(t *testing.T) {
	t.Run("can be instantiated with all fields", func(t *testing.T) {
		result := TestResult{
			Passed:   10,
			Failed:   2,
			Skipped:  1,
			Packages: []TestPackage{},
		}

		if result.Passed != 10 {
			t.Errorf("Passed = %v, want 10", result.Passed)
		}
		if result.Failed != 2 {
			t.Errorf("Failed = %v, want 2", result.Failed)
		}
		if len(result.Packages) != 0 {
			t.Errorf("Packages length = %v, want 0", len(result.Packages))
		}
		if result.Skipped != 1 {
			t.Errorf("Skipped = %v, want 1", result.Skipped)
		}
	})
}

func TestTestPackageType(t *testing.T) {
	pkg := TestPackage{
		Package: "github.com/example/pkg",
		Passed:  true,
		Elapsed: 1.234,
		Tests: []TestCase{
			{
				Name:     "TestSomething",
				Package:  "github.com/example/pkg",
				Passed:   true,
				Duration: 0.5,
				Output:   "PASS",
			},
		},
	}

	if pkg.Package != "github.com/example/pkg" {
		t.Errorf("Package = %v, want github.com/example/pkg", pkg.Package)
	}
	if !pkg.Passed {
		t.Errorf("Passed = %v, want true", pkg.Passed)
	}
	if pkg.Elapsed != 1.234 {
		t.Errorf("Elapsed = %v, want 1.234", pkg.Elapsed)
	}
	if len(pkg.Tests) != 1 {
		t.Errorf("Tests length = %v, want 1", len(pkg.Tests))
	}
}

func TestTestCaseType(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		pkg      string
		passed   bool
		duration float64
		output   string
	}{
		{"passing test", "TestAdd", "pkg/math", true, 0.001, "PASS"},
		{"failing test", "TestSub", "pkg/math", false, 0.002, "expected 5, got 3"},
		{"slow test", "TestIntegration", "pkg/integration", true, 5.0, "PASS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := TestCase{
				Name:     tt.testName,
				Package:  tt.pkg,
				Passed:   tt.passed,
				Duration: tt.duration,
				Output:   tt.output,
			}

			if tc.Name != tt.testName {
				t.Errorf("Name = %v, want %v", tc.Name, tt.testName)
			}
			if tc.Package != tt.pkg {
				t.Errorf("Package = %v, want %v", tc.Package, tt.pkg)
			}
			if tc.Passed != tt.passed {
				t.Errorf("Passed = %v, want %v", tc.Passed, tt.passed)
			}
			if tc.Duration != tt.duration {
				t.Errorf("Duration = %v, want %v", tc.Duration, tt.duration)
			}
			if tc.Output != tt.output {
				t.Errorf("Output = %v, want %v", tc.Output, tt.output)
			}
		})
	}
}

func TestTestResultJSONMarshal(t *testing.T) {
	result := TestResult{
		Passed:  5,
		Failed:  1,
		Skipped: 0,
		Packages: []TestPackage{
			{
				Package: "pkg/example",
				Passed:  false,
				Elapsed: 2.5,
				Tests: []TestCase{
					{Name: "TestFoo", Package: "pkg/example", Passed: true, Duration: 0.1, Output: "PASS"},
				},
			},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expectedKeys := []string{"passed", "failed", "skipped", "packages"}
	for _, key := range expectedKeys {
		if _, exists := unmarshaled[key]; !exists {
			t.Errorf("Missing expected key in JSON: %s", key)
		}
	}
}

func TestTestResultTypeMatchesExpected(t *testing.T) {
	tr := TestResult{}
	v := reflect.TypeOf(tr)

	expectedFields := map[string]string{
		"Passed":   "int",
		"Failed":   "int",
		"Skipped":  "int",
		"Packages": "[]golang.TestPackage",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

func TestTestPackageTypeMatchesExpected(t *testing.T) {
	tp := TestPackage{}
	v := reflect.TypeOf(tp)

	expectedFields := map[string]string{
		"Package": "string",
		"Passed":  "bool",
		"Elapsed": "float64",
		"Tests":   "[]golang.TestCase",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

func TestTestCaseTypeMatchesExpected(t *testing.T) {
	tc := TestCase{}
	v := reflect.TypeOf(tc)

	expectedFields := map[string]string{
		"Name":     "string",
		"Package":  "string",
		"Passed":   "bool",
		"Duration": "float64",
		"Output":   "string",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

// Coverage Tests

func TestCoverageType(t *testing.T) {
	coverage := Coverage{
		Total: 85.5,
		Packages: []PackageCoverage{
			{
				Package:  "pkg/example",
				Coverage: 90.0,
			},
		},
	}

	if coverage.Total != 85.5 {
		t.Errorf("Total = %v, want 85.5", coverage.Total)
	}
	if len(coverage.Packages) != 1 {
		t.Errorf("Packages length = %v, want 1", len(coverage.Packages))
	}
	if coverage.Packages[0].Package != "pkg/example" {
		t.Errorf("Package = %v, want pkg/example", coverage.Packages[0].Package)
	}
	if coverage.Packages[0].Coverage != 90.0 {
		t.Errorf("Coverage = %v, want 90.0", coverage.Packages[0].Coverage)
	}
}

func TestCoverageJSONMarshal(t *testing.T) {
	coverage := Coverage{
		Total: 75.0,
		Packages: []PackageCoverage{
			{Package: "pkg/a", Coverage: 80.0},
			{Package: "pkg/b", Coverage: 70.0},
		},
	}

	data, err := json.Marshal(coverage)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expectedKeys := []string{"total", "packages"}
	for _, key := range expectedKeys {
		if _, exists := unmarshaled[key]; !exists {
			t.Errorf("Missing expected key in JSON: %s", key)
		}
	}
}

func TestCoverageTypeMatchesExpected(t *testing.T) {
	c := Coverage{}
	v := reflect.TypeOf(c)

	expectedFields := map[string]string{
		"Total":    "float64",
		"Packages": "[]golang.PackageCoverage",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

func TestPackageCoverageTypeMatchesExpected(t *testing.T) {
	pc := PackageCoverage{}
	v := reflect.TypeOf(pc)

	expectedFields := map[string]string{
		"Package":  "string",
		"Coverage": "float64",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

// Vet Tests

func TestVetResultType(t *testing.T) {
	t.Run("clean vet result", func(t *testing.T) {
		vet := VetResult{
			Issues: []VetIssue{},
		}

		if len(vet.Issues) != 0 {
			t.Errorf("Issues length = %v, want 0", len(vet.Issues))
		}
	})

	t.Run("vet result with issues", func(t *testing.T) {
		vet := VetResult{
			Issues: []VetIssue{
				{
					File:    "main.go",
					Line:    15,
					Column:  10,
					Message: "printf: format string not verified",
				},
			},
		}

		if len(vet.Issues) != 1 {
			t.Errorf("Issues length = %v, want 1", len(vet.Issues))
		}
		if vet.Issues[0].File != "main.go" {
			t.Errorf("File = %v, want main.go", vet.Issues[0].File)
		}
	})
}

func TestVetIssueType(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		line    int
		column  int
		message string
	}{
		{"printf error", "main.go", 10, 5, "printf: format string not verified"},
		{"unreachable code", "helper.go", 25, 1, "unreachable code"},
		{"shadow warning", "service.go", 50, 10, "declaration of err shadows declaration"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vi := VetIssue{
				File:    tt.file,
				Line:    tt.line,
				Column:  tt.column,
				Message: tt.message,
			}

			if vi.File != tt.file {
				t.Errorf("File = %v, want %v", vi.File, tt.file)
			}
			if vi.Line != tt.line {
				t.Errorf("Line = %v, want %v", vi.Line, tt.line)
			}
			if vi.Column != tt.column {
				t.Errorf("Column = %v, want %v", vi.Column, tt.column)
			}
			if vi.Message != tt.message {
				t.Errorf("Message = %v, want %v", vi.Message, tt.message)
			}
		})
	}
}

func TestVetResultJSONMarshal(t *testing.T) {
	vet := VetResult{
		Issues: []VetIssue{
			{File: "main.go", Line: 10, Column: 5, Message: "issue here"},
		},
	}

	data, err := json.Marshal(vet)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, exists := unmarshaled["issues"]; !exists {
		t.Error("Missing expected key in JSON: issues")
	}
}

func TestVetResultTypeMatchesExpected(t *testing.T) {
	vr := VetResult{}
	v := reflect.TypeOf(vr)

	field, fieldOK := v.FieldByName("Issues")
	if !fieldOK {
		t.Error("Missing field: Issues")
		return
	}

	actualType := field.Type.String()
	if actualType != "[]golang.VetIssue" {
		t.Errorf("Field Issues has type %s, want []golang.VetIssue", actualType)
	}
}

func TestVetIssueTypeMatchesExpected(t *testing.T) {
	vi := VetIssue{}
	v := reflect.TypeOf(vi)

	expectedFields := map[string]string{
		"File":    "string",
		"Line":    "int",
		"Column":  "int",
		"Message": "string",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

// RunResult Tests

func TestRunResultType(t *testing.T) {
	t.Run("successful run", func(t *testing.T) {
		run := RunResult{
			ExitCode: 0,
			Stdout:   "Hello, World!",
			Stderr:   "",
		}

		if run.ExitCode != 0 {
			t.Errorf("ExitCode = %v, want 0", run.ExitCode)
		}
		if run.Stdout != "Hello, World!" {
			t.Errorf("Stdout = %v, want Hello, World!", run.Stdout)
		}
		if run.Stderr != "" {
			t.Errorf("Stderr = %v, want empty", run.Stderr)
		}
	})

	t.Run("failed run", func(t *testing.T) {
		run := RunResult{
			ExitCode: 1,
			Stdout:   "",
			Stderr:   "error: something went wrong",
		}

		if run.ExitCode != 1 {
			t.Errorf("ExitCode = %v, want 1", run.ExitCode)
		}
		if run.Stdout != "" {
			t.Errorf("Stdout = %v, want empty", run.Stdout)
		}
		if run.Stderr != "error: something went wrong" {
			t.Errorf("Stderr = %v, want error message", run.Stderr)
		}
	})
}

func TestRunResultJSONMarshal(t *testing.T) {
	run := RunResult{
		ExitCode: 0,
		Stdout:   "output",
		Stderr:   "",
	}

	data, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expectedKeys := []string{"exitCode", "stdout", "stderr"}
	for _, key := range expectedKeys {
		if _, exists := unmarshaled[key]; !exists {
			t.Errorf("Missing expected key in JSON: %s", key)
		}
	}
}

func TestRunResultTypeMatchesExpected(t *testing.T) {
	rr := RunResult{}
	v := reflect.TypeOf(rr)

	expectedFields := map[string]string{
		"ExitCode": "int",
		"Stdout":   "string",
		"Stderr":   "string",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

// ModTidyResult Tests

func TestModTidyResultType(t *testing.T) {
	t.Run("no changes", func(t *testing.T) {
		mod := ModTidyResult{
			Added:   []string{},
			Removed: []string{},
		}

		if len(mod.Added) != 0 {
			t.Errorf("Added length = %v, want 0", len(mod.Added))
		}
		if len(mod.Removed) != 0 {
			t.Errorf("Removed length = %v, want 0", len(mod.Removed))
		}
	})

	t.Run("with changes", func(t *testing.T) {
		mod := ModTidyResult{
			Added:   []string{"github.com/new/dep v1.0.0"},
			Removed: []string{"github.com/old/dep v0.5.0"},
		}

		if len(mod.Added) != 1 {
			t.Errorf("Added length = %v, want 1", len(mod.Added))
		}
		if len(mod.Removed) != 1 {
			t.Errorf("Removed length = %v, want 1", len(mod.Removed))
		}
		if mod.Added[0] != "github.com/new/dep v1.0.0" {
			t.Errorf("Added[0] = %v, want github.com/new/dep v1.0.0", mod.Added[0])
		}
		if mod.Removed[0] != "github.com/old/dep v0.5.0" {
			t.Errorf("Removed[0] = %v, want github.com/old/dep v0.5.0", mod.Removed[0])
		}
	})
}

func TestModTidyResultJSONMarshal(t *testing.T) {
	mod := ModTidyResult{
		Added:   []string{"pkg/a"},
		Removed: []string{"pkg/b"},
	}

	data, err := json.Marshal(mod)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expectedKeys := []string{"added", "removed"}
	for _, key := range expectedKeys {
		if _, exists := unmarshaled[key]; !exists {
			t.Errorf("Missing expected key in JSON: %s", key)
		}
	}
}

func TestModTidyResultTypeMatchesExpected(t *testing.T) {
	mtr := ModTidyResult{}
	v := reflect.TypeOf(mtr)

	expectedFields := map[string]string{
		"Added":   "[]string",
		"Removed": "[]string",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

// FmtResult Tests

func TestFmtResultType(t *testing.T) {
	t.Run("no unformatted files", func(t *testing.T) {
		fmt := FmtResult{
			Unformatted: []string{},
		}

		if len(fmt.Unformatted) != 0 {
			t.Errorf("Unformatted length = %v, want 0", len(fmt.Unformatted))
		}
	})

	t.Run("with unformatted files", func(t *testing.T) {
		fmt := FmtResult{
			Unformatted: []string{"main.go", "helper.go"},
		}

		if len(fmt.Unformatted) != 2 {
			t.Errorf("Unformatted length = %v, want 2", len(fmt.Unformatted))
		}
		if fmt.Unformatted[0] != "main.go" {
			t.Errorf("Unformatted[0] = %v, want main.go", fmt.Unformatted[0])
		}
	})
}

func TestFmtResultJSONMarshal(t *testing.T) {
	fmt := FmtResult{
		Unformatted: []string{"file.go"},
	}

	data, err := json.Marshal(fmt)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, exists := unmarshaled["unformatted"]; !exists {
		t.Error("Missing expected key in JSON: unformatted")
	}
}

func TestFmtResultTypeMatchesExpected(t *testing.T) {
	fr := FmtResult{}
	v := reflect.TypeOf(fr)

	field, fieldOK := v.FieldByName("Unformatted")
	if !fieldOK {
		t.Error("Missing field: Unformatted")
		return
	}

	actualType := field.Type.String()
	if actualType != "[]string" {
		t.Errorf("Field Unformatted has type %s, want []string", actualType)
	}
}

// GenerateResult Tests

func TestGenerateResultType(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		gen := GenerateResult{
			Success:   true,
			Generated: []string{"mock_service.go", "mock_repo.go"},
		}

		if !gen.Success {
			t.Errorf("Success = %v, want true", gen.Success)
		}
		if len(gen.Generated) != 2 {
			t.Errorf("Generated length = %v, want 2", len(gen.Generated))
		}
	})

	t.Run("failed generation", func(t *testing.T) {
		gen := GenerateResult{
			Success:   false,
			Generated: []string{},
		}

		if gen.Success {
			t.Errorf("Success = %v, want false", gen.Success)
		}
		if len(gen.Generated) != 0 {
			t.Errorf("Generated length = %v, want 0", len(gen.Generated))
		}
	})
}

func TestGenerateResultJSONMarshal(t *testing.T) {
	gen := GenerateResult{
		Success:   true,
		Generated: []string{"gen.go"},
	}

	data, err := json.Marshal(gen)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	expectedKeys := []string{"success", "generated"}
	for _, key := range expectedKeys {
		if _, exists := unmarshaled[key]; !exists {
			t.Errorf("Missing expected key in JSON: %s", key)
		}
	}
}

func TestGenerateResultTypeMatchesExpected(t *testing.T) {
	gr := GenerateResult{}
	v := reflect.TypeOf(gr)

	expectedFields := map[string]string{
		"Success":   "bool",
		"Generated": "[]string",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}
