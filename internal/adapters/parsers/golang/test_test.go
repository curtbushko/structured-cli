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
	requiredProps := []string{"passed", "failed", "skipped", "packages_passed", "packages_failed", "packages"}
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

func TestTestParser_CachedTests(t *testing.T) {
	// go test -json output with cached test results
	// Cached tests show "ok" with "(cached)" and only emit package-level pass event
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/cached"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/cached","Output":"ok  \tgithub.com/example/cached\t(cached)\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"pass","Package":"github.com/example/cached","Elapsed":0}`

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

	// Verify package is marked as passed
	if len(got.Packages) != 1 {
		t.Fatalf("TestResult.Packages length = %d, want 1", len(got.Packages))
	}

	pkg := got.Packages[0]
	if pkg.Package != "github.com/example/cached" {
		t.Errorf("Package.Package = %q, want %q", pkg.Package, "github.com/example/cached")
	}
	if !pkg.Passed {
		t.Error("Package.Passed = false, want true for cached test")
	}
	if !pkg.Cached {
		t.Error("Package.Cached = false, want true for cached test results")
	}

	// Verify PackagesPassed includes the cached package
	if got.PackagesPassed != 1 {
		t.Errorf("TestResult.PackagesPassed = %d, want 1", got.PackagesPassed)
	}
	if got.PackagesFailed != 0 {
		t.Errorf("TestResult.PackagesFailed = %d, want 0", got.PackagesFailed)
	}
}

func TestTestParser_MixedCachedAndNewTests(t *testing.T) {
	// go test -json output with mix of cached and new tests
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/cached"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/cached","Output":"ok  \tgithub.com/example/cached\t(cached)\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"pass","Package":"github.com/example/cached","Elapsed":0}
{"Time":"2024-01-15T10:00:01Z","Action":"start","Package":"github.com/example/new"}
{"Time":"2024-01-15T10:00:01Z","Action":"run","Package":"github.com/example/new","Test":"TestNew"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/new","Test":"TestNew","Output":"=== RUN   TestNew\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/new","Test":"TestNew","Output":"--- PASS: TestNew (0.01s)\n"}
{"Time":"2024-01-15T10:00:02Z","Action":"pass","Package":"github.com/example/new","Test":"TestNew","Elapsed":0.01}
{"Time":"2024-01-15T10:00:02Z","Action":"output","Package":"github.com/example/new","Output":"PASS\n"}
{"Time":"2024-01-15T10:00:02Z","Action":"pass","Package":"github.com/example/new","Elapsed":0.02}`

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

	// Verify both packages present
	if len(got.Packages) != 2 {
		t.Fatalf("TestResult.Packages length = %d, want 2", len(got.Packages))
	}

	// Find packages by name
	pkgByName := make(map[string]*TestPackage)
	for i := range got.Packages {
		pkgByName[got.Packages[i].Package] = &got.Packages[i]
	}

	// Check cached package
	cachedPkg, ok := pkgByName["github.com/example/cached"]
	if !ok {
		t.Fatal("Missing github.com/example/cached package")
	}
	if !cachedPkg.Passed {
		t.Error("cached Package.Passed = false, want true")
	}
	if !cachedPkg.Cached {
		t.Error("cached Package.Cached = false, want true")
	}

	// Check new package
	newPkg, ok := pkgByName["github.com/example/new"]
	if !ok {
		t.Fatal("Missing github.com/example/new package")
	}
	if !newPkg.Passed {
		t.Error("new Package.Passed = false, want true")
	}
	if newPkg.Cached {
		t.Error("new Package.Cached = true, want false")
	}

	// Verify counts
	if got.Passed != 1 {
		t.Errorf("TestResult.Passed = %d, want 1 (only new test)", got.Passed)
	}
	if got.PackagesPassed != 2 {
		t.Errorf("TestResult.PackagesPassed = %d, want 2 (both packages passed)", got.PackagesPassed)
	}
	if got.PackagesFailed != 0 {
		t.Errorf("TestResult.PackagesFailed = %d, want 0", got.PackagesFailed)
	}
}

func TestTestParser_PassingTestsNoOutput(t *testing.T) {
	// go test -json output with all tests passing
	// For token savings, passing tests should NOT have output populated
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestPass1"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass1","Output":"=== RUN   TestPass1\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass1","Output":"--- PASS: TestPass1 (0.01s)\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestPass1","Elapsed":0.01}
{"Time":"2024-01-15T10:00:01Z","Action":"run","Package":"github.com/example/pkg","Test":"TestPass2"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass2","Output":"=== RUN   TestPass2\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass2","Output":"    some debug output\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass2","Output":"--- PASS: TestPass2 (0.02s)\n"}
{"Time":"2024-01-15T10:00:02Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestPass2","Elapsed":0.02}
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

	// Verify that passing tests do NOT have output (token savings)
	if len(got.Packages) != 1 {
		t.Fatalf("TestResult.Packages length = %d, want 1", len(got.Packages))
	}

	pkg := got.Packages[0]
	for _, test := range pkg.Tests {
		if test.Output != "" {
			t.Errorf("Passing test %q has Output = %q, want empty (token savings)", test.Name, test.Output)
		}
	}
}

func TestTestParser_FailingTestsWithOutput(t *testing.T) {
	// go test -json output with failing tests
	// Failing tests SHOULD retain full output for debugging
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestFail"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFail","Output":"=== RUN   TestFail\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFail","Output":"    test_test.go:15: expected 1, got 2\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFail","Output":"    test_test.go:16: additional error details\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFail","Output":"--- FAIL: TestFail (0.02s)\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"fail","Package":"github.com/example/pkg","Test":"TestFail","Elapsed":0.02}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Output":"FAIL\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"fail","Package":"github.com/example/pkg","Elapsed":0.05}`

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
	if got.Failed != 1 {
		t.Errorf("TestResult.Failed = %d, want 1", got.Failed)
	}

	// Verify that failing test DOES have output for debugging
	if len(got.Packages) != 1 {
		t.Fatalf("TestResult.Packages length = %d, want 1", len(got.Packages))
	}

	pkg := got.Packages[0]
	if len(pkg.Tests) != 1 {
		t.Fatalf("Package.Tests length = %d, want 1", len(pkg.Tests))
	}

	failedTest := pkg.Tests[0]
	if failedTest.Output == "" {
		t.Error("Failing test has empty Output, want error details")
	}
	if !strings.Contains(failedTest.Output, "expected 1, got 2") {
		t.Errorf("Failing test Output = %q, want to contain 'expected 1, got 2'", failedTest.Output)
	}
	if !strings.Contains(failedTest.Output, "additional error details") {
		t.Errorf("Failing test Output = %q, want to contain 'additional error details'", failedTest.Output)
	}
}

func TestTestParser_MixedPassFailOutput(t *testing.T) {
	// go test -json output with mixed passing and failing tests
	// Only failing tests should have output populated
	input := `{"Time":"2024-01-15T10:00:00Z","Action":"start","Package":"github.com/example/pkg"}
{"Time":"2024-01-15T10:00:00Z","Action":"run","Package":"github.com/example/pkg","Test":"TestPass"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass","Output":"=== RUN   TestPass\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass","Output":"    some logging output\n"}
{"Time":"2024-01-15T10:00:00Z","Action":"output","Package":"github.com/example/pkg","Test":"TestPass","Output":"--- PASS: TestPass (0.01s)\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"pass","Package":"github.com/example/pkg","Test":"TestPass","Elapsed":0.01}
{"Time":"2024-01-15T10:00:01Z","Action":"run","Package":"github.com/example/pkg","Test":"TestFail"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFail","Output":"=== RUN   TestFail\n"}
{"Time":"2024-01-15T10:00:01Z","Action":"output","Package":"github.com/example/pkg","Test":"TestFail","Output":"    error: something went wrong\n"}
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

	if len(got.Packages) != 1 {
		t.Fatalf("TestResult.Packages length = %d, want 1", len(got.Packages))
	}

	pkg := got.Packages[0]
	if len(pkg.Tests) != 2 {
		t.Fatalf("Package.Tests length = %d, want 2", len(pkg.Tests))
	}

	// Find tests by name
	testByName := make(map[string]*TestCase)
	for i := range pkg.Tests {
		testByName[pkg.Tests[i].Name] = &pkg.Tests[i]
	}

	// Verify passing test has NO output (token savings)
	passTest, ok := testByName["TestPass"]
	if !ok {
		t.Fatal("Missing TestPass in package tests")
	}
	if passTest.Output != "" {
		t.Errorf("Passing TestPass has Output = %q, want empty (token savings)", passTest.Output)
	}

	// Verify failing test HAS output for debugging
	failTest, ok := testByName["TestFail"]
	if !ok {
		t.Fatal("Missing TestFail in package tests")
	}
	if failTest.Output == "" {
		t.Error("Failing TestFail has empty Output, want error details")
	}
	if !strings.Contains(failTest.Output, "error: something went wrong") {
		t.Errorf("Failing TestFail Output = %q, want to contain 'error: something went wrong'", failTest.Output)
	}
}

func TestTestParser_SkippedTestsNoOutput(t *testing.T) {
	// go test -json output with skipped tests
	// Skipped tests should NOT have output populated (similar to passing)
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
	if got.Skipped != 1 {
		t.Errorf("TestResult.Skipped = %d, want 1", got.Skipped)
	}

	if len(got.Packages) != 1 {
		t.Fatalf("TestResult.Packages length = %d, want 1", len(got.Packages))
	}

	pkg := got.Packages[0]
	if len(pkg.Tests) != 1 {
		t.Fatalf("Package.Tests length = %d, want 1", len(pkg.Tests))
	}

	// Verify skipped test has NO output (token savings)
	skippedTest := pkg.Tests[0]
	if skippedTest.Output != "" {
		t.Errorf("Skipped test has Output = %q, want empty (token savings)", skippedTest.Output)
	}
}
