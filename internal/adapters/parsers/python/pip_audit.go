package python

import (
	"bufio"
	"encoding/json"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing pip-audit output.
var (
	// vulnerabilityTablePattern matches vulnerability table rows.
	// Format: "package version id fix_versions"
	vulnerabilityTablePattern = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(.*)$`)

	// summaryPattern matches "Found N known vulnerabilities"
	summaryPattern = regexp.MustCompile(`Found (\d+) known vulnerabilit(?:y|ies)`)

	// tableHeaderPattern matches the header line to skip it
	tableHeaderPattern = regexp.MustCompile(`^Name\s+Version\s+ID`)

	// tableSeparatorPattern matches the separator line
	tableSeparatorPattern = regexp.MustCompile(`^-+\s+-+\s+-+`)
)

// pipAuditJSONEntry represents a single package entry in pip-audit JSON output.
type pipAuditJSONEntry struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Vulns   []pipAuditJSONVul `json:"vulns"`
}

// pipAuditJSONVul represents a vulnerability in pip-audit JSON output.
type pipAuditJSONVul struct {
	ID          string   `json:"id"`
	FixVersions []string `json:"fix_versions"`
	Aliases     []string `json:"aliases"`
	Description string   `json:"description"`
}

// PipAuditParser parses the output of 'pip-audit'.
type PipAuditParser struct {
	schema domain.Schema
}

// NewPipAuditParser creates a new PipAuditParser with the pip-audit schema.
func NewPipAuditParser() *PipAuditParser {
	return &PipAuditParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/pip-audit.json",
			"Pip-Audit Output",
			"object",
			map[string]domain.PropertySchema{
				"success":          {Type: "boolean", Description: "Whether audit found no vulnerabilities"},
				"vulnerabilities":  {Type: "array", Description: "List of vulnerabilities found"},
				"packages_scanned": {Type: "integer", Description: "Number of packages scanned"},
				"summary":          {Type: "object", Description: "Vulnerability count summary"},
			},
			[]string{"success", "vulnerabilities"},
		),
	}
}

// Parse reads pip-audit output and returns structured data.
func (p *PipAuditParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &PipAuditResult{
		Success:         true,
		Vulnerabilities: []PipVulnerability{},
	}

	// Try JSON format first
	if strings.HasPrefix(strings.TrimSpace(raw), "[") {
		parsePipAuditJSON(raw, result)
	} else {
		parsePipAuditTable(raw, result)
	}

	// Mark as unsuccessful if vulnerabilities found
	if len(result.Vulnerabilities) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parsePipAuditJSON parses JSON format output from pip-audit.
func parsePipAuditJSON(output string, result *PipAuditResult) {
	var entries []pipAuditJSONEntry
	if err := json.Unmarshal([]byte(output), &entries); err != nil {
		return
	}

	for _, entry := range entries {
		for _, vuln := range entry.Vulns {
			result.Vulnerabilities = append(result.Vulnerabilities, PipVulnerability{
				Name:        entry.Name,
				Version:     entry.Version,
				ID:          vuln.ID,
				Description: vuln.Description,
				FixVersions: vuln.FixVersions,
				Aliases:     vuln.Aliases,
			})
		}
	}

	result.Summary.Total = len(result.Vulnerabilities)
}

// parsePipAuditTable parses tabular format output from pip-audit.
func parsePipAuditTable(output string, result *PipAuditResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	inTable := false

	for scanner.Scan() {
		line := scanner.Text()

		// Skip header line
		if tableHeaderPattern.MatchString(line) {
			inTable = true
			continue
		}

		// Skip separator line
		if tableSeparatorPattern.MatchString(line) {
			continue
		}

		// Parse summary line
		if matches := summaryPattern.FindStringSubmatch(line); matches != nil {
			result.Summary.Total, _ = strconv.Atoi(matches[1])
			continue
		}

		// Parse vulnerability rows (only if we've seen the header)
		if inTable {
			parseVulnerabilityRow(line, result)
		}
	}
}

// parseVulnerabilityRow parses a single vulnerability row from pip-audit table output.
func parseVulnerabilityRow(line string, result *PipAuditResult) {
	matches := vulnerabilityTablePattern.FindStringSubmatch(line)
	if matches == nil {
		return
	}

	name := matches[1]
	version := matches[2]
	id := matches[3]
	fixVersions := strings.TrimSpace(matches[4])

	// Skip empty lines or lines that look like separators
	if name == "" || strings.HasPrefix(name, "-") {
		return
	}

	vuln := PipVulnerability{
		Name:    name,
		Version: version,
		ID:      id,
	}

	if fixVersions != "" {
		vuln.FixVersions = strings.Split(fixVersions, ",")
		for i, v := range vuln.FixVersions {
			vuln.FixVersions[i] = strings.TrimSpace(v)
		}
	}

	result.Vulnerabilities = append(result.Vulnerabilities, vuln)

	if result.Summary.Total == 0 {
		result.Summary.Total = len(result.Vulnerabilities)
	}
}

// Schema returns the JSON Schema for pip-audit output.
func (p *PipAuditParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PipAuditParser) Matches(cmd string, subcommands []string) bool {
	return cmd == pipAuditCommand
}
