package golang

import (
	"strings"
	"testing"
)

func TestTestParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches go test",
			cmd:         "go",
			subcommands: []string{"test"},
			want:        true,
		},
		{
			name:        "matches go test with path",
			cmd:         "go",
			subcommands: []string{"test", "./..."},
			want:        true,
		},
		{
			name:        "matches go test with flags",
			cmd:         "go",
			subcommands: []string{"test", "-v", "-json"},
			want:        true,
		},
		{
			name:        "does not match go build",
			cmd:         "go",
			subcommands: []string{"build"},
			want:        false,
		},
		{
			name:        "does not match git",
			cmd:         "git",
			subcommands: []string{"test"},
			want:        false,
		},
		{
			name:        "does not match go without subcommand",
			cmd:         "go",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{"test"},
			want:        false,
		},
	}

	parser := NewTestParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestTestParser_Schema(t *testing.T) {
	parser := NewTestParser()
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

	// Verify required properties exist
	requiredProps := []string{"passed", "failed", "skipped", "packages"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestTestParser_AllPass(t *testing.T) {
	// go test -json output for all tests passing
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestOne"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestOne","Output":"=== RUN   TestOne\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestOne","Output":"--- PASS: TestOne (0.01s)\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestOne","Elapsed":0.01}
{"Time":"2024-01-15T10:00:01Z","Action":"run","Package":"github.com/example/pkg","Test":"TestTwo"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestTwo","Output":"=== RUN   TestTwo\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestTwo","Output":"--- PASS: TestTwo (0.02s)\n"}
{"Time":"2024-01-15T10:00:02Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestTwo","Elapsed":0.02}
{"Time":"2024-01-15T10:00:02Z","Action":"output","Package":"github.com/example/pkg","Output":"PASS\n"}
{"Time":"2024-01-15T10:00:02Z","Action":"pass","Package":"github.com/example/pkg","Elapsed":0.05}`

	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	// Verify counts
	if got.Passed != 2 {
		t.Errorf("TestResult.Passed = %d, want 2", got.Passed)
	}
	if got.Failed != 0 {
		t.Errorf("TestResult.Failed = %d, want 0", got.Failed)
	}
	if got.Skipped != 0 {
		t.Errorf("TestResult.Skipped = %d, want 0", got.Skipped)
	}

	// Verify packages
	if len(got.Packages) != 1 {
		t.Fatalf("TestResult.Packages length = %d, want 1", len(got.Packages))
	}

	pkg := got.Packages[0]
	if pkg.Package != "github.com/example/pkg" {
		t.Errorf("Package.Package = %q, want %q", pkg.Package, "github.com/example/pkg")
	}
	if !pkg.Passed {
		t.Error("Package.Passed = false, want true")
	}
	if len(pkg.Tests) != 2 {
		t.Errorf("Package.Tests length = %d, want 2", len(pkg.Tests))
	}
}

func TestTestParser_SomeFail(t *testing.T) {
	// go test -json output with some failures
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestPass"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass","Output":"=== RUN   TestPass\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass","Output":"--- PASS: TestPass (0.01s)\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestPass","Elapsed":0.01}
{"Time":"2024-01-15T10:00:01Z","Action":"run","Package":"github.com/example/pkg","Test":"TestFail"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFail","Output":"=== RUN   TestFail\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFail","Output":"    test_test.go:15: expected 1, got 2\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFail","Output":"--- FAIL: TestFail (0.02s)\n"}
{"Time":"2024-01-15T10:00:02Z","Action":"fail","Package":"github.com/example/pkg","Test":"TestFail","Elapsed":0.02}
{"Time":"2024-01-15T10:00:02Z","Action":"output","Package":"github.com/example/pkg","Output":"FAIL\n"}
{"Time":"2024-01-15T10:00:02Z","Action":"fail","Package":"github.com/example/pkg","Elapsed":0.05}`

	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	// Verify counts
	if got.Passed != 1 {
		t.Errorf("TestResult.Passed = %d, want 1", got.Passed)
	}
	if got.Failed != 1 {
		t.Errorf("TestResult.Failed = %d, want 1", got.Failed)
	}
	if got.Skipped != 0 {
		t.Errorf("TestResult.Skipped = %d, want 0", got.Skipped)
	}

	// Verify package marked as failed
	if len(got.Packages) != 1 {
		t.Fatalf("TestResult.Packages length = %d, want 1", len(got.Packages))
	}

	pkg := got.Packages[0]
	if pkg.Passed {
		t.Error("Package.Passed = true, want false when tests fail")
	}

	// Verify failed test captures output
	var failedTest *TestCase
	for i := range pkg.Tests {
		if pkg.Tests[i].Name == "TestFail" {
			failedTest = &pkg.Tests[i]
			break
		}
	}
	if failedTest == nil {
		t.Fatal("Failed to find TestFail in package tests")
	}
	if failedTest.Passed {
		t.Error("TestFail.Passed = true, want false")
	}
	if !strings.Contains(failedTest.Output, "expected 1, got 2") {
		t.Errorf("TestFail.Output = %q, want to contain 'expected 1, got 2'", failedTest.Output)
	}
}

func TestTestParser_SkippedTests(t *testing.T) {
	// go test -json output with skipped tests
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestSkipped"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestSkipped","Output":"=== RUN   TestSkipped\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestSkipped","Output":"    test_test.go:20: skipping in short mode\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestSkipped","Output":"--- SKIP: TestSkipped (0.00s)\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"skip","Package":"github.com/example/pkg","Test":"TestSkipped","Elapsed":0.00}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Output":"PASS\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Elapsed":0.01}`

	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	// Verify counts
	if got.Passed != 0 {
		t.Errorf("TestResult.Passed = %d, want 0", got.Passed)
	}
	if got.Failed != 0 {
		t.Errorf("TestResult.Failed = %d, want 0", got.Failed)
	}
	if got.Skipped != 1 {
		t.Errorf("TestResult.Skipped = %d, want 1", got.Skipped)
	}
}

func TestTestParser_MultiplePackages(t *testing.T) {
	// go test -json output with multiple packages
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg1"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg1","Test":"TestOne"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg1","Test":"TestOne","Elapsed":0.01}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg1","Elapsed":0.02}
{"Time":"2024-01-15T10:00:01Z","Action":"start","Package":"github.com/example/pkg2"}
{"Time":"2024-01-15T10:00:01Z","Action":"run","Package":"github.com/example/pkg2","Test":"TestTwo"}
{"Time":"2024-01-15T10:00:02Z","Action":"pass","Package":"github.com/example/pkg2","Test":"TestTwo","Elapsed":0.01}
{"Time":"2024-01-15T10:00:02Z","Action":"pass","Package":"github.com/example/pkg2","Elapsed":0.02}`

	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	// Verify counts
	if got.Passed != 2 {
		t.Errorf("TestResult.Passed = %d, want 2", got.Passed)
	}

	// Verify packages
	if len(got.Packages) != 2 {
		t.Fatalf("TestResult.Packages length = %d, want 2", len(got.Packages))
	}
}

func TestTestParser_EmptyOutput(t *testing.T) {
	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	// Empty output should result in zero counts
	if got.Passed != 0 {
		t.Errorf("TestResult.Passed = %d, want 0", got.Passed)
	}
	if got.Failed != 0 {
		t.Errorf("TestResult.Failed = %d, want 0", got.Failed)
	}
	if got.Skipped != 0 {
		t.Errorf("TestResult.Skipped = %d, want 0", got.Skipped)
	}
	if len(got.Packages) != 0 {
		t.Errorf("TestResult.Packages length = %d, want 0", len(got.Packages))
	}
}

func TestTestParser_TestDuration(t *testing.T) {
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestSlow"}
{"Time":"2024-01-15T10:00:05Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestSlow","Elapsed":5.123}
{"Time":"2024-01-15T10:00:05Z","Action":"pass","Package":"github.com/example/pkg","Elapsed":5.2}`

	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	if len(got.Packages) != 1 {
		t.Fatalf("TestResult.Packages length = %d, want 1", len(got.Packages))
	}

	pkg := got.Packages[0]
	if len(pkg.Tests) != 1 {
		t.Fatalf("Package.Tests length = %d, want 1", len(pkg.Tests))
	}

	// Check test duration
	if pkg.Tests[0].Duration != 5.123 {
		t.Errorf("Test.Duration = %f, want 5.123", pkg.Tests[0].Duration)
	}

	// Check package elapsed time
	if pkg.Elapsed != 5.2 {
		t.Errorf("Package.Elapsed = %f, want 5.2", pkg.Elapsed)
	}
}

func TestTestParser_WithCoverage(t *testing.T) {
	// go test -json -cover output with coverage line
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestOne"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestOne","Elapsed":0.01}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Output":"coverage: 85.5% of statements\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Elapsed":0.02}`

	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	// Verify coverage is populated
	if got.Coverage == nil {
		t.Fatal("TestResult.Coverage = nil, want non-nil")
	}

	if got.Coverage.Total != 85.5 {
		t.Errorf("Coverage.Total = %f, want 85.5", got.Coverage.Total)
	}

	// Verify per-package coverage
	if len(got.Coverage.Packages) != 1 {
		t.Fatalf("Coverage.Packages length = %d, want 1", len(got.Coverage.Packages))
	}

	pkgCov := got.Coverage.Packages[0]
	if pkgCov.Package != "github.com/example/pkg" {
		t.Errorf("PackageCoverage.Package = %q, want %q", pkgCov.Package, "github.com/example/pkg")
	}
	if pkgCov.Coverage != 85.5 {
		t.Errorf("PackageCoverage.Coverage = %f, want 85.5", pkgCov.Coverage)
	}
}

func TestTestParser_PerPackageCoverage(t *testing.T) {
	// go test -json -cover output with multiple packages
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg1"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg1","Test":"TestOne"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg1","Test":"TestOne","Elapsed":0.01}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg1","Output":"coverage: 75.0% of statements\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg1","Elapsed":0.02}
{"Time":"2024-01-15T10:00:01Z","Action":"start","Package":"github.com/example/pkg2"}
{"Time":"2024-01-15T10:00:01Z","Action":"run","Package":"github.com/example/pkg2","Test":"TestTwo"}
{"Time":"2024-01-15T10:00:02Z","Action":"pass","Package":"github.com/example/pkg2","Test":"TestTwo","Elapsed":0.01}
{"Time":"2024-01-15T10:00:02Z","Action":"output","Package":"github.com/example/pkg2","Output":"coverage: 92.3% of statements\n"}
{"Time":"2024-01-15T10:00:02Z","Action":"pass","Package":"github.com/example/pkg2","Elapsed":0.02}`

	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	// Verify coverage is populated
	if got.Coverage == nil {
		t.Fatal("TestResult.Coverage = nil, want non-nil")
	}

	// Verify per-package coverage has 2 packages
	if len(got.Coverage.Packages) != 2 {
		t.Fatalf("Coverage.Packages length = %d, want 2", len(got.Coverage.Packages))
	}

	// Find packages by name
	pkgCoverages := make(map[string]float64)
	for _, pc := range got.Coverage.Packages {
		pkgCoverages[pc.Package] = pc.Coverage
	}

	if cov, ok := pkgCoverages["github.com/example/pkg1"]; !ok {
		t.Error("Coverage.Packages missing github.com/example/pkg1")
	} else if cov != 75.0 {
		t.Errorf("pkg1 coverage = %f, want 75.0", cov)
	}

	if cov, ok := pkgCoverages["github.com/example/pkg2"]; !ok {
		t.Error("Coverage.Packages missing github.com/example/pkg2")
	} else if cov != 92.3 {
		t.Errorf("pkg2 coverage = %f, want 92.3", cov)
	}

	// Verify total is average of both packages
	expectedTotal := (75.0 + 92.3) / 2.0
	if got.Coverage.Total != expectedTotal {
		t.Errorf("Coverage.Total = %f, want %f (average)", got.Coverage.Total, expectedTotal)
	}
}

func TestTestParser_NoCoverage(t *testing.T) {
	// go test -json output WITHOUT -cover flag (no coverage output)
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestOne"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestOne","Elapsed":0.01}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Elapsed":0.02}`

	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	// Coverage should be nil when not present
	if got.Coverage != nil {
		t.Errorf("TestResult.Coverage = %v, want nil when no coverage output", got.Coverage)
	}
}
