package golang

import (
	"bufio"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// coveragePattern matches "coverage: X.X% of statements" in output.
var coveragePattern = regexp.MustCompile(`coverage:\s+(\d+(?:\.\d+)?)\%\s+of\s+statements`)

// TestEvent represents a single JSON event from 'go test -json' output.
// Each line of go test -json output is a TestEvent.
type TestEvent struct {
	// Time is when the event occurred.
	Time string `json:"Time"`

	// Action is the event type: run, pause, cont, pass, bench, fail, skip, output, start.
	Action string `json:"Action"`

	// Package is the package being tested.
	Package string `json:"Package"`

	// Test is the name of the test (empty for package-level events).
	Test string `json:"Test,omitempty"`

	// Output is the test output text (for output actions).
	Output string `json:"Output,omitempty"`

	// Elapsed is the time in seconds for pass/fail events.
	Elapsed float64 `json:"Elapsed,omitempty"`
}

// TestParser parses the output of 'go test -json'.
// It aggregates test events into structured TestResult.
type TestParser struct {
	schema domain.Schema
}

// NewTestParser creates a new TestParser with the go-test schema.
func NewTestParser() *TestParser {
	return &TestParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/go-test.json",
			"Go Test Output",
			"object",
			map[string]domain.PropertySchema{
				"passed":   {Type: "integer", Description: "Total number of tests that passed"},
				"failed":   {Type: "integer", Description: "Total number of tests that failed"},
				"skipped":  {Type: "integer", Description: "Total number of tests that were skipped"},
				"packages": {Type: "array", Description: "Per-package test results"},
			},
			[]string{"passed", "failed", "skipped", "packages"},
		),
	}
}

// testParseState holds intermediate state during parsing.
type testParseState struct {
	result          *TestResult
	packages        map[string]*TestPackage
	testOutput      map[string]string
	packageCoverage map[string]float64
}

// Parse reads go test -json output and returns structured TestResult.
func (p *TestParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	state := &testParseState{
		result: &TestResult{
			Passed:   0,
			Failed:   0,
			Skipped:  0,
			Packages: []TestPackage{},
		},
		packages:        make(map[string]*TestPackage),
		testOutput:      make(map[string]string),
		packageCoverage: make(map[string]float64),
	}

	if len(data) == 0 {
		return domain.NewParseResult(state.result, raw, 0), nil
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		state.processEvent(&event)
	}

	// Convert packages map to slice
	for _, pkg := range state.packages {
		state.result.Packages = append(state.result.Packages, *pkg)
	}

	// Build coverage information if we collected any
	state.finalizeCoverage()

	return domain.NewParseResult(state.result, raw, 0), nil
}

// finalizeCoverage builds the Coverage struct from collected package coverage.
func (s *testParseState) finalizeCoverage() {
	if len(s.packageCoverage) == 0 {
		return
	}

	coverage := &Coverage{
		Packages: make([]PackageCoverage, 0, len(s.packageCoverage)),
	}

	var totalCoverage float64
	for pkg, cov := range s.packageCoverage {
		coverage.Packages = append(coverage.Packages, PackageCoverage{
			Package:  pkg,
			Coverage: cov,
		})
		totalCoverage += cov
	}

	// Calculate average coverage
	coverage.Total = totalCoverage / float64(len(s.packageCoverage))

	s.result.Coverage = coverage
}

// processEvent handles a single test event.
func (s *testParseState) processEvent(event *TestEvent) {
	s.ensurePackage(event.Package)

	switch event.Action {
	case "output":
		s.handleOutput(event)
	case "pass":
		s.handlePass(event)
	case "fail":
		s.handleFail(event)
	case "skip":
		s.handleSkip(event)
	}
}

// ensurePackage creates the package entry if it doesn't exist.
func (s *testParseState) ensurePackage(pkg string) {
	if pkg == "" {
		return
	}
	if _, exists := s.packages[pkg]; !exists {
		s.packages[pkg] = &TestPackage{
			Package: pkg,
			Passed:  true, // Assume passed until we see a failure
			Tests:   []TestCase{},
		}
	}
}

// handleOutput accumulates test output and parses coverage lines.
func (s *testParseState) handleOutput(event *TestEvent) {
	if event.Test != "" {
		key := event.Package + "/" + event.Test
		s.testOutput[key] += event.Output
	}

	// Check for coverage line (package-level output)
	if event.Test == "" && event.Package != "" {
		if cov := parseCoverageLine(event.Output); cov >= 0 {
			s.packageCoverage[event.Package] = cov
		}
	}
}

// parseCoverageLine extracts coverage percentage from a coverage output line.
// Returns -1 if the line doesn't contain coverage information.
func parseCoverageLine(output string) float64 {
	matches := coveragePattern.FindStringSubmatch(output)
	if len(matches) < 2 {
		return -1
	}

	cov, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return -1
	}

	return cov
}

// handlePass processes a pass event.
func (s *testParseState) handlePass(event *TestEvent) {
	pkg := s.packages[event.Package]
	if event.Test != "" {
		s.result.Passed++
		key := event.Package + "/" + event.Test
		pkg.Tests = append(pkg.Tests, TestCase{
			Name:     event.Test,
			Package:  event.Package,
			Passed:   true,
			Duration: event.Elapsed,
			Output:   s.testOutput[key],
		})
	} else {
		pkg.Elapsed = event.Elapsed
	}
}

// handleFail processes a fail event.
func (s *testParseState) handleFail(event *TestEvent) {
	pkg := s.packages[event.Package]
	if event.Test != "" {
		s.result.Failed++
		key := event.Package + "/" + event.Test
		pkg.Tests = append(pkg.Tests, TestCase{
			Name:     event.Test,
			Package:  event.Package,
			Passed:   false,
			Duration: event.Elapsed,
			Output:   s.testOutput[key],
		})
		pkg.Passed = false
	} else {
		pkg.Elapsed = event.Elapsed
		pkg.Passed = false
	}
}

// handleSkip processes a skip event.
func (s *testParseState) handleSkip(event *TestEvent) {
	if event.Test != "" {
		s.result.Skipped++
		key := event.Package + "/" + event.Test
		pkg := s.packages[event.Package]
		pkg.Tests = append(pkg.Tests, TestCase{
			Name:     event.Test,
			Package:  event.Package,
			Passed:   true, // Skipped tests aren't failures
			Duration: event.Elapsed,
			Output:   s.testOutput[key],
		})
	}
}

// Schema returns the JSON Schema for go test output.
func (p *TestParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *TestParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "go" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "test"
}
