package build

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// webpackAssetPattern matches asset lines in webpack output.
// Format: "asset main.js 1.5 KiB [emitted] (name: main)"
// Format: "asset bundle.js 2.5 MiB [emitted] (name: main)"
// Format: "asset bundle.js 512 bytes [emitted] (name: main)"
var webpackAssetPattern = regexp.MustCompile(`^asset\s+(\S+)\s+(\d+(?:\.\d+)?)\s*(bytes|KiB|MiB)\s+\[([^\]]*)\](?:\s+\(name:\s*([^)]+)\))?`)

// webpackChunkPattern matches chunk lines in webpack output.
// Format: "chunk (runtime: main) main.js (main) 1.5 KiB [entry] [rendered]"
var webpackChunkPattern = regexp.MustCompile(`^chunk\s+\(runtime:\s*[^)]+\)\s+(\S+)\s+\(([^)]+)\)\s+(\d+(?:\.\d+)?)\s*(bytes|KiB|MiB)\s+(.*)`)

// webpackErrorPattern matches ERROR lines in webpack output.
// Format: "ERROR in ./src/index.js 10:5"
// Format: "ERROR in ./src/index.js"
var webpackErrorPattern = regexp.MustCompile(`^ERROR\s+in\s+(\S+)(?:\s+(\d+):(\d+))?`)

// webpackWarningPattern matches WARNING lines in webpack output.
// Format: "WARNING in ./src/utils.js 25:10"
// Format: "WARNING in ./src/utils.js"
var webpackWarningPattern = regexp.MustCompile(`^WARNING\s+in\s+(\S+)(?:\s+(\d+):(\d+))?`)

// webpackModulesPattern matches the modules count line.
// Format: "42 modules"
var webpackModulesPattern = regexp.MustCompile(`^(\d+)\s+modules?$`)

// webpackDurationPattern matches the compilation result line.
// Formats: "webpack 5.75.0 compiled successfully in 1234 ms"
// Formats: "webpack 5.75.0 compiled with 1 error in 234 ms"
// Formats: "webpack 5.75.0 compiled with 1 warning in 500 ms"
var webpackDurationPattern = regexp.MustCompile(`webpack\s+[\d.]+\s+compiled\s+(?:successfully|with\s+\d+\s+(?:error|warning)s?)\s+in\s+(\d+)\s+ms`)

// WebpackParser parses the output of 'webpack' bundler.
type WebpackParser struct {
	schema domain.Schema
}

// NewWebpackParser creates a new WebpackParser with the webpack schema.
func NewWebpackParser() *WebpackParser {
	return &WebpackParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/webpack.json",
			"Webpack Bundler Output",
			"object",
			map[string]domain.PropertySchema{
				"success":  {Type: "boolean", Description: "Whether the build succeeded"},
				"errors":   {Type: "array", Description: "Build errors"},
				"warnings": {Type: "array", Description: "Build warnings"},
				"assets":   {Type: "array", Description: "Generated output files"},
				"chunks":   {Type: "array", Description: "Webpack chunks"},
				"modules":  {Type: "integer", Description: "Number of modules processed"},
				"duration": {Type: "number", Description: "Build duration in milliseconds"},
			},
			[]string{"success", "errors", "warnings", "assets"},
		),
	}
}

// Parse reads webpack output and returns structured data.
func (p *WebpackParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &WebpackResult{
		Success:  true,
		Errors:   []WebpackError{},
		Warnings: []WebpackWarning{},
		Assets:   []WebpackAsset{},
		Chunks:   []WebpackChunk{},
		Modules:  0,
		Duration: 0,
	}

	parseWebpackOutput(raw, result)

	// Success is false when errors are present
	if len(result.Errors) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// webpackParseState holds the parsing state during line-by-line processing.
type webpackParseState struct {
	inErrorBlock   bool
	inWarningBlock bool
	currentFile    string
	currentLine    int
	currentColumn  int
	messageLines   []string
}

// parseWebpackOutput scans webpack output and extracts assets, chunks, errors, warnings, and duration.
func parseWebpackOutput(output string, result *WebpackResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	state := &webpackParseState{}

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		parseWebpackLine(trimmedLine, result, state)
	}

	// Finalize any pending error or warning
	finalizeErrorOrWarning(result, state)
}

// parseWebpackLine processes a single line of webpack output.
func parseWebpackLine(line string, result *WebpackResult, state *webpackParseState) {
	// Check for new ERROR block
	if matches := webpackErrorPattern.FindStringSubmatch(line); matches != nil {
		// Finalize previous error/warning if any
		finalizeErrorOrWarning(result, state)

		state.inErrorBlock = true
		state.inWarningBlock = false
		state.currentFile = matches[1]
		state.currentLine = 0
		state.currentColumn = 0
		state.messageLines = []string{}

		if matches[2] != "" {
			state.currentLine, _ = strconv.Atoi(matches[2])
		}
		if matches[3] != "" {
			state.currentColumn, _ = strconv.Atoi(matches[3])
		}
		return
	}

	// Check for new WARNING block
	if matches := webpackWarningPattern.FindStringSubmatch(line); matches != nil {
		// Finalize previous error/warning if any
		finalizeErrorOrWarning(result, state)

		state.inWarningBlock = true
		state.inErrorBlock = false
		state.currentFile = matches[1]
		state.currentLine = 0
		state.currentColumn = 0
		state.messageLines = []string{}

		if matches[2] != "" {
			state.currentLine, _ = strconv.Atoi(matches[2])
		}
		if matches[3] != "" {
			state.currentColumn, _ = strconv.Atoi(matches[3])
		}
		return
	}

	// If in error or warning block, collect message lines
	if state.inErrorBlock || state.inWarningBlock {
		// Check if this line starts a new section (asset, chunk, webpack line)
		if webpackAssetPattern.MatchString(line) ||
			webpackChunkPattern.MatchString(line) ||
			strings.HasPrefix(line, "webpack ") ||
			webpackModulesPattern.MatchString(line) {
			finalizeErrorOrWarning(result, state)
			// Fall through to parse the line normally
		} else {
			// Collect as message line
			state.messageLines = append(state.messageLines, line)
			return
		}
	}

	// Parse non-error/warning lines
	parseWebpackNonErrorLine(line, result)
}

// finalizeErrorOrWarning creates an error or warning from collected state.
func finalizeErrorOrWarning(result *WebpackResult, state *webpackParseState) {
	if state.inErrorBlock && state.currentFile != "" {
		message := strings.Join(state.messageLines, " ")
		result.Errors = append(result.Errors, WebpackError{
			File:    state.currentFile,
			Line:    state.currentLine,
			Column:  state.currentColumn,
			Message: strings.TrimSpace(message),
		})
	}

	if state.inWarningBlock && state.currentFile != "" {
		message := strings.Join(state.messageLines, " ")
		result.Warnings = append(result.Warnings, WebpackWarning{
			File:    state.currentFile,
			Line:    state.currentLine,
			Column:  state.currentColumn,
			Message: strings.TrimSpace(message),
		})
	}

	// Reset state
	state.inErrorBlock = false
	state.inWarningBlock = false
	state.currentFile = ""
	state.currentLine = 0
	state.currentColumn = 0
	state.messageLines = []string{}
}

// parseWebpackNonErrorLine processes lines outside error/warning blocks.
func parseWebpackNonErrorLine(line string, result *WebpackResult) {
	// Try to parse asset line
	if matches := webpackAssetPattern.FindStringSubmatch(line); matches != nil {
		name := matches[1]
		sizeValue, _ := strconv.ParseFloat(matches[2], 64)
		sizeUnit := matches[3]
		flags := matches[4]
		chunkName := matches[5]

		size := convertToBytes(sizeValue, sizeUnit)
		emitted := strings.Contains(flags, "emitted")

		var chunkNames []string
		if chunkName != "" {
			chunkNames = []string{chunkName}
		}

		result.Assets = append(result.Assets, WebpackAsset{
			Name:       name,
			Size:       size,
			Emitted:    emitted,
			ChunkNames: chunkNames,
		})
		return
	}

	// Try to parse chunk line
	if matches := webpackChunkPattern.FindStringSubmatch(line); matches != nil {
		name := matches[1]
		chunkName := matches[2]
		sizeValue, _ := strconv.ParseFloat(matches[3], 64)
		sizeUnit := matches[4]
		flags := matches[5]

		size := convertToBytes(sizeValue, sizeUnit)
		entry := strings.Contains(flags, "[entry]")
		initial := strings.Contains(flags, "[initial]")
		rendered := strings.Contains(flags, "[rendered]")

		result.Chunks = append(result.Chunks, WebpackChunk{
			Name:      name,
			ChunkName: chunkName,
			Size:      size,
			Entry:     entry,
			Initial:   initial,
			Rendered:  rendered,
		})
		return
	}

	// Try to parse modules count
	if matches := webpackModulesPattern.FindStringSubmatch(line); matches != nil {
		modules, _ := strconv.Atoi(matches[1])
		result.Modules = modules
		return
	}

	// Try to parse duration
	if matches := webpackDurationPattern.FindStringSubmatch(line); matches != nil {
		duration, _ := strconv.ParseFloat(matches[1], 64)
		result.Duration = duration
	}
}

// convertToBytes converts a size value with unit to bytes.
func convertToBytes(value float64, unit string) int64 {
	switch unit {
	case "bytes":
		return int64(value)
	case "KiB":
		return int64(value * 1024)
	case "MiB":
		return int64(value * 1024 * 1024)
	default:
		return int64(value)
	}
}

// Schema returns the JSON Schema for webpack output.
func (p *WebpackParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
// The webpack parser matches the "webpack" command with any subcommands/flags.
func (p *WebpackParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "webpack"
}
