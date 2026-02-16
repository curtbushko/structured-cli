package docker

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing docker logs output.
var (
	// timestampPattern matches ISO 8601 timestamps at the start of log lines.
	timestampPattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+Z)\s+(.*)$`)
)

// LogsParser parses the output of 'docker logs'.
type LogsParser struct {
	schema domain.Schema
}

// NewLogsParser creates a new LogsParser with the docker-logs schema.
func NewLogsParser() *LogsParser {
	return &LogsParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-logs.json",
			"Docker Logs Output",
			"object",
			map[string]domain.PropertySchema{
				"success":      {Type: "boolean", Description: "Whether the command completed successfully"},
				"container_id": {Type: "string", Description: "Container ID"},
				"lines":        {Type: "array", Description: "Log lines"},
				"total_lines":  {Type: "integer", Description: "Total number of log lines"},
			},
			[]string{"success", "lines", "total_lines"},
		),
	}
}

// Parse reads docker logs output and returns structured data.
func (p *LogsParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &LogsResult{
		Success: true,
		Lines:   []LogLine{},
	}

	parseLogsOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseLogsOutput extracts log information from the output.
func parseLogsOutput(output string, result *LogsResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		logLine := LogLine{
			Stream: "stdout", // Default stream
		}

		// Check for timestamp prefix
		if matches := timestampPattern.FindStringSubmatch(line); matches != nil {
			logLine.Timestamp = matches[1]
			logLine.Message = matches[2]
		} else {
			logLine.Message = line
		}

		result.Lines = append(result.Lines, logLine)
		result.TotalLines++
	}
}

// Schema returns the JSON Schema for docker logs output.
func (p *LogsParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *LogsParser) Matches(cmd string, subcommands []string) bool {
	if cmd != dockerCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// docker logs
	if subcommands[0] == subLogs {
		return true
	}

	// docker container logs
	if len(subcommands) >= 2 && subcommands[0] == subContainer && subcommands[1] == subLogs {
		return true
	}

	return false
}
