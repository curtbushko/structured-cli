package kubectl

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regular expressions for parsing kubectl logs output.
var (
	// RFC3339 timestamp pattern used by --timestamps flag
	logTimestampPattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z)\s+(.*)$`)

	// Log level pattern - matches common log level prefixes
	logLevelPattern = regexp.MustCompile(`^(DEBUG|INFO|WARN(?:ING)?|ERROR|FATAL|TRACE)\s+(.*)$`)

	// Combined timestamp and log level pattern
	timestampLevelPattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z)\s+(DEBUG|INFO|WARN(?:ING)?|ERROR|FATAL|TRACE)\s+(.*)$`)
)

// LogsResult represents the structured output of 'kubectl logs'.
type LogsResult struct {
	// Container is the container name (when specified with -c flag).
	Container string `json:"container,omitempty"`

	// Lines is the list of log lines.
	Lines []LogLine `json:"lines"`

	// LineCount is the total number of log lines.
	LineCount int `json:"line_count"`
}

// LogLine represents a single log line.
type LogLine struct {
	// Timestamp is the log timestamp (when --timestamps is used).
	Timestamp string `json:"timestamp,omitempty"`

	// Level is the log level (INFO, WARN, ERROR, etc.) if detected.
	Level string `json:"level,omitempty"`

	// Message is the log message content.
	Message string `json:"message"`
}

// LogsParser parses the output of 'kubectl logs'.
type LogsParser struct {
	schema domain.Schema
}

// NewLogsParser creates a new LogsParser.
func NewLogsParser() *LogsParser {
	return &LogsParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/kubectl-logs.json",
			"Kubectl Logs Output",
			"object",
			map[string]domain.PropertySchema{
				"container":  {Type: "string", Description: "Container name"},
				"lines":      {Type: "array", Description: "List of log lines"},
				"line_count": {Type: "integer", Description: "Total number of log lines"},
			},
			[]string{"lines", "line_count"},
		),
	}
}

// Parse reads kubectl logs output and returns structured data.
func (p *LogsParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)
	result := p.parseLogsOutput(raw)

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for kubectl logs output.
func (p *LogsParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *LogsParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdKubectl || len(subcommands) < 1 {
		return false
	}
	return subcommands[0] == "logs"
}

// parseLogsOutput parses the logs output into structured data.
func (p *LogsParser) parseLogsOutput(raw string) *LogsResult {
	result := &LogsResult{
		Lines: []LogLine{},
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines at end of output
		if line == "" && !scanner.Scan() {
			break
		}
		if line == "" {
			// Put back the line we peeked
			line = scanner.Text()
		}

		logLine := p.parseLine(line)
		result.Lines = append(result.Lines, logLine)
	}

	result.LineCount = len(result.Lines)
	return result
}

// parseLine parses a single log line.
func (p *LogsParser) parseLine(line string) LogLine {
	logLine := LogLine{}

	// Try to match timestamp + level first
	if matches := timestampLevelPattern.FindStringSubmatch(line); matches != nil {
		logLine.Timestamp = matches[1]
		logLine.Level = normalizeLogLevel(matches[2])
		logLine.Message = matches[3]
		return logLine
	}

	// Try to match just timestamp
	if matches := logTimestampPattern.FindStringSubmatch(line); matches != nil {
		logLine.Timestamp = matches[1]
		remaining := matches[2]

		// Check if remaining has a log level
		if levelMatches := logLevelPattern.FindStringSubmatch(remaining); levelMatches != nil {
			logLine.Level = normalizeLogLevel(levelMatches[1])
			logLine.Message = levelMatches[2]
		} else {
			logLine.Message = remaining
		}
		return logLine
	}

	// Try to match just log level
	if matches := logLevelPattern.FindStringSubmatch(line); matches != nil {
		logLine.Level = normalizeLogLevel(matches[1])
		logLine.Message = matches[2]
		return logLine
	}

	// Plain message
	logLine.Message = line
	return logLine
}

// normalizeLogLevel normalizes log level variants to standard form.
func normalizeLogLevel(level string) string {
	if level == "WARNING" {
		return "WARN"
	}
	return level
}
