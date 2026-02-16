package npm

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing npm audit output
var (
	// vulnerabilityHeaderPattern matches package name and version constraint
	// e.g., "lodash  <4.17.21"
	vulnerabilityHeaderPattern = regexp.MustCompile(`^(\S+)\s+(<[^>]+|>=?[^<]+|[*])$`)

	// severityPattern matches "Severity: high"
	severityPattern = regexp.MustCompile(`^Severity:\s+(low|moderate|high|critical)$`)

	// auditSummaryPattern matches "X vulnerabilities (1 low, 2 moderate, ...)"
	auditSummaryPattern = regexp.MustCompile(`^(\d+)\s+vulnerabilit(?:y|ies)`)

	// foundVulnPattern matches "found X vulnerabilities"
	foundVulnPattern = regexp.MustCompile(`found\s+(\d+)\s+vulnerabilit(?:y|ies)`)
)

// AuditParser parses the output of 'npm audit'.
type AuditParser struct {
	schema domain.Schema
}

// NewAuditParser creates a new AuditParser with the npm-audit schema.
func NewAuditParser() *AuditParser {
	return &AuditParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/npm-audit.json",
			"NPM Audit Output",
			"object",
			map[string]domain.PropertySchema{
				"success":         {Type: "boolean", Description: "Whether audit found no vulnerabilities"},
				"vulnerabilities": {Type: "array", Description: "List of vulnerabilities found"},
				"summary":         {Type: "object", Description: "Vulnerability count summary"},
				"metadata":        {Type: "object", Description: "Audit metadata"},
			},
			[]string{"success", "vulnerabilities", "summary"},
		),
	}
}

// Parse reads npm audit output and returns structured data.
func (p *AuditParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &AuditResult{
		Success:         true,
		Vulnerabilities: []Vulnerability{},
	}

	parseAuditOutput(raw, result)

	// If vulnerabilities exist, mark as not successful
	if result.Summary.Total > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// auditParseState tracks state while parsing audit output.
type auditParseState struct {
	currentVuln     *Vulnerability
	inVulnerability bool
}

// parseAuditOutput extracts audit information from the output.
func parseAuditOutput(output string, result *AuditResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	state := &auditParseState{}

	for scanner.Scan() {
		line := scanner.Text()
		parseAuditLine(line, result, state)
	}

	// Save last vulnerability
	if state.currentVuln != nil {
		result.Vulnerabilities = append(result.Vulnerabilities, *state.currentVuln)
	}
}

// parseAuditLine processes a single line of audit output.
func parseAuditLine(line string, result *AuditResult, state *auditParseState) {
	// Check for "found X vulnerabilities" (clean audit output)
	if matches := foundVulnPattern.FindStringSubmatch(line); matches != nil {
		result.Summary.Total, _ = strconv.Atoi(matches[1])
		return
	}

	// Check for vulnerability header (package name and version)
	if matches := vulnerabilityHeaderPattern.FindStringSubmatch(line); matches != nil {
		if state.currentVuln != nil {
			result.Vulnerabilities = append(result.Vulnerabilities, *state.currentVuln)
		}
		state.currentVuln = &Vulnerability{
			Name:  matches[1],
			Range: matches[2],
		}
		state.inVulnerability = true
		return
	}

	// Process vulnerability details
	if state.inVulnerability && state.currentVuln != nil {
		if parseVulnerabilityLine(line, state) {
			return
		}
	}

	// Check for vulnerability summary line
	if matches := auditSummaryPattern.FindStringSubmatch(line); matches != nil {
		result.Summary.Total, _ = strconv.Atoi(matches[1])
		parseVulnerabilityDetails(line, &result.Summary)
	}
}

// parseVulnerabilityLine processes a line within a vulnerability block.
// Returns true if the line was handled.
func parseVulnerabilityLine(line string, state *auditParseState) bool {
	// Check for severity
	if matches := severityPattern.FindStringSubmatch(line); matches != nil {
		state.currentVuln.Severity = matches[1]
		return true
	}

	// Check for title/description
	if isTitleLine(line, state.currentVuln) {
		state.currentVuln.Title = strings.TrimSpace(line)
		return true
	}

	// Check for URL
	if strings.HasPrefix(line, "https://") || strings.HasPrefix(line, "http://") {
		state.currentVuln.URL = strings.TrimSpace(line)
		return true
	}

	// Check for fix available
	if strings.Contains(line, "fix available") {
		state.currentVuln.FixAvailable = true
		return true
	}

	// Check for node_modules path (end of vulnerability block)
	if strings.HasPrefix(line, "node_modules/") {
		state.inVulnerability = false
		return true
	}

	return false
}

// isTitleLine checks if a line is a valid title for the vulnerability.
func isTitleLine(line string, vuln *Vulnerability) bool {
	if vuln.Title != "" {
		return false
	}
	if line == "" {
		return false
	}
	if strings.HasPrefix(line, "http") {
		return false
	}
	if strings.HasPrefix(line, "fix available") {
		return false
	}
	if strings.HasPrefix(line, "node_modules") {
		return false
	}
	if strings.HasPrefix(line, "Severity:") {
		return false
	}
	return true
}

// Schema returns the JSON Schema for npm audit output.
func (p *AuditParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *AuditParser) Matches(cmd string, subcommands []string) bool {
	if cmd != npmCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	return subcommands[0] == "audit"
}
