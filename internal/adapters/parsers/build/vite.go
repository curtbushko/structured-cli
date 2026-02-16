package build

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// viteOutputPattern matches output file lines in Vite build output.
// Formats supported:
// - "dist/index.html                   0.46 kB │ gzip: 0.30 kB"
// - "dist/my-lib.js      0.08 kB / gzip: 0.07 kB"
var viteOutputPattern = regexp.MustCompile(`^([^\s]+)\s+(\d+(?:\.\d+)?)\s*kB\s*[│/]\s*gzip:\s*(\d+(?:\.\d+)?)\s*kB`)

// viteModulesPattern matches the modules transformed line.
// Format: "✓ 32 modules transformed."
var viteModulesPattern = regexp.MustCompile(`[✓]\s*(\d+)\s+modules?\s+transformed`)

// viteDurationPattern matches the build duration line.
// Formats: "✓ built in 1.90s" or "✓ built in 500ms"
var viteDurationPattern = regexp.MustCompile(`[✓]\s*built\s+in\s+(\d+(?:\.\d+)?)(ms|s)`)

// viteBuildFailedPattern matches the build failed line.
// Format: "x Build failed in 57ms"
var viteBuildFailedPattern = regexp.MustCompile(`x\s+Build\s+failed\s+in\s+(\d+(?:\.\d+)?)(ms|s)`)

// vitePluginErrorPattern matches plugin error lines.
// Format: "[vite:plugin-name] Error message"
var vitePluginErrorPattern = regexp.MustCompile(`^\[([^\]]+)\]\s+(.+)$`)

// viteRollupErrorPattern matches Rollup error lines.
// Format: "[ERROR] Error message"
var viteRollupErrorPattern = regexp.MustCompile(`^\[ERROR\]\s+(.+)$`)

// ViteParser parses the output of 'vite build' command.
type ViteParser struct {
	schema domain.Schema
}

// NewViteParser creates a new ViteParser with the vite build schema.
func NewViteParser() *ViteParser {
	return &ViteParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/vite-build.json",
			"Vite Build Output",
			"object",
			map[string]domain.PropertySchema{
				"success":  {Type: "boolean", Description: "Whether the build succeeded"},
				"errors":   {Type: "array", Description: "Build errors"},
				"warnings": {Type: "array", Description: "Build warnings"},
				"outputs":  {Type: "array", Description: "Generated output files"},
				"duration": {Type: "number", Description: "Build duration in milliseconds"},
				"modules":  {Type: "integer", Description: "Number of modules transformed"},
			},
			[]string{"success", "errors", "warnings", "outputs"},
		),
	}
}

// Parse reads vite build output and returns structured data.
func (p *ViteParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ViteResult{
		Success:  true,
		Errors:   []ViteError{},
		Warnings: []ViteWarning{},
		Outputs:  []ViteOutput{},
		Duration: 0,
		Modules:  0,
	}

	parseViteOutput(raw, result)

	// Success is false when errors are present
	if len(result.Errors) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// viteParseState holds the parsing state during line-by-line processing.
type viteParseState struct {
	inErrorBlock bool
	errorLines   []string
}

// parseViteOutput scans vite build output and extracts files, errors, duration, and modules.
func parseViteOutput(output string, result *ViteResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	state := &viteParseState{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parseLine(line, result, state)
	}

	// If we collected generic error lines, add them as a single error
	if len(state.errorLines) > 0 {
		result.Errors = append(result.Errors, ViteError{
			Message: strings.Join(state.errorLines, " "),
		})
	}
}

// parseLine processes a single line of vite build output.
func parseLine(line string, result *ViteResult, state *viteParseState) {
	// Check for "error during build:" which starts an error block
	if strings.Contains(line, "error during build") {
		state.inErrorBlock = true
		return
	}

	// If we're in an error block, handle error lines
	if state.inErrorBlock {
		parseErrorBlockLine(line, result, state)
		return
	}

	// Parse non-error lines
	parseNonErrorLine(line, result)
}

// parseErrorBlockLine processes a line within an error block.
func parseErrorBlockLine(line string, result *ViteResult, state *viteParseState) {
	// Check for plugin error format: [plugin-name] message
	if matches := vitePluginErrorPattern.FindStringSubmatch(line); matches != nil {
		result.Errors = append(result.Errors, ViteError{
			Plugin:  matches[1],
			Message: matches[2],
		})
		state.inErrorBlock = false
		state.errorLines = []string{}
		return
	}

	// Otherwise collect as generic error line
	if !strings.HasPrefix(line, "vite v") && !strings.HasPrefix(line, "✓") {
		state.errorLines = append(state.errorLines, line)
	}
}

// parseNonErrorLine processes a line outside of error blocks.
func parseNonErrorLine(line string, result *ViteResult) {
	// Check for build failed indicator
	if duration, ok := parseDuration(viteBuildFailedPattern, line); ok {
		result.Duration = duration
		return
	}

	// Check for [ERROR] format (Rollup errors)
	if matches := viteRollupErrorPattern.FindStringSubmatch(line); matches != nil {
		result.Errors = append(result.Errors, ViteError{
			Message: matches[1],
		})
		return
	}

	// Try to parse output file line
	if output, ok := parseOutputFile(line); ok {
		result.Outputs = append(result.Outputs, output)
		return
	}

	// Try to parse modules transformed line
	if matches := viteModulesPattern.FindStringSubmatch(line); matches != nil {
		modules, _ := strconv.Atoi(matches[1])
		result.Modules = modules
		return
	}

	// Try to parse duration line
	if duration, ok := parseDuration(viteDurationPattern, line); ok {
		result.Duration = duration
	}
}

// parseDuration extracts duration in milliseconds from a line using the given pattern.
func parseDuration(pattern *regexp.Regexp, line string) (float64, bool) {
	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		return 0, false
	}

	value, _ := strconv.ParseFloat(matches[1], 64)
	unit := matches[2]
	if unit == "s" {
		value *= 1000
	}
	return value, true
}

// parseOutputFile extracts output file information from a line.
func parseOutputFile(line string) (ViteOutput, bool) {
	matches := viteOutputPattern.FindStringSubmatch(line)
	if matches == nil {
		return ViteOutput{}, false
	}

	filePath := matches[1]
	sizeKB, _ := strconv.ParseFloat(matches[2], 64)
	gzipKB, _ := strconv.ParseFloat(matches[3], 64)

	// Convert kB to bytes (1 kB = 1024 bytes)
	return ViteOutput{
		Path:     filePath,
		Size:     int64(sizeKB * 1024),
		GzipSize: int64(gzipKB * 1024),
	}, true
}

// Schema returns the JSON Schema for vite build output.
func (p *ViteParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
// The vite parser matches "vite build" command.
func (p *ViteParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "vite" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "build"
}
