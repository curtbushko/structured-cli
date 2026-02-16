package npm

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing npm install output
var (
	// addedPattern matches "added X packages" or "added X packages, ..."
	addedPattern = regexp.MustCompile(`added (\d+) packages?`)

	// removedPattern matches "removed X packages"
	removedPattern = regexp.MustCompile(`removed (\d+) packages?`)

	// changedPattern matches "changed X packages"
	changedPattern = regexp.MustCompile(`changed (\d+) packages?`)

	// auditedPattern matches "audited X packages"
	auditedPattern = regexp.MustCompile(`audited (\d+) packages?`)

	// fundingPattern matches "X packages are looking for funding"
	fundingPattern = regexp.MustCompile(`(\d+) packages? (?:are|is) looking for funding`)

	// vulnerabilitiesPattern matches "X vulnerabilities (1 low, 2 moderate, ...)"
	vulnerabilitiesPattern = regexp.MustCompile(`(\d+) vulnerabilit(?:y|ies)`)

	// vulnDetailPattern matches individual vulnerability counts like "1 low"
	vulnDetailPattern = regexp.MustCompile(`(\d+) (low|moderate|high|critical)`)
)

// InstallParser parses the output of 'npm install'.
type InstallParser struct {
	schema domain.Schema
}

// NewInstallParser creates a new InstallParser with the npm-install schema.
func NewInstallParser() *InstallParser {
	return &InstallParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/npm-install.json",
			"NPM Install Output",
			"object",
			map[string]domain.PropertySchema{
				"success":          {Type: "boolean", Description: "Whether installation completed successfully"},
				"packages_added":   {Type: "integer", Description: "Number of packages added"},
				"packages_removed": {Type: "integer", Description: "Number of packages removed"},
				"packages_changed": {Type: "integer", Description: "Number of packages changed"},
				"packages_audited": {Type: "integer", Description: "Number of packages audited"},
				"funding":          {Type: "integer", Description: "Number of packages seeking funding"},
				"vulnerabilities":  {Type: "object", Description: "Vulnerability summary"},
				"warnings":         {Type: "array", Description: "Installation warnings"},
			},
			[]string{"success", "packages_added", "vulnerabilities"},
		),
	}
}

// Parse reads npm install output and returns structured data.
func (p *InstallParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &InstallResult{
		Success:  true,
		Warnings: []string{},
	}

	parseInstallOutput(raw, result)

	// If vulnerabilities exist, mark as not successful
	if result.Vulnerabilities.Total > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseInstallOutput extracts install information from the output.
func parseInstallOutput(output string, result *InstallResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Parse added packages
		if matches := addedPattern.FindStringSubmatch(line); matches != nil {
			result.PackagesAdded, _ = strconv.Atoi(matches[1])
		}

		// Parse removed packages
		if matches := removedPattern.FindStringSubmatch(line); matches != nil {
			result.PackagesRemoved, _ = strconv.Atoi(matches[1])
		}

		// Parse changed packages
		if matches := changedPattern.FindStringSubmatch(line); matches != nil {
			result.PackagesChanged, _ = strconv.Atoi(matches[1])
		}

		// Parse audited packages
		if matches := auditedPattern.FindStringSubmatch(line); matches != nil {
			result.PackagesAudited, _ = strconv.Atoi(matches[1])
		}

		// Parse funding
		if matches := fundingPattern.FindStringSubmatch(line); matches != nil {
			result.Funding, _ = strconv.Atoi(matches[1])
		}

		// Parse vulnerabilities
		if matches := vulnerabilitiesPattern.FindStringSubmatch(line); matches != nil {
			result.Vulnerabilities.Total, _ = strconv.Atoi(matches[1])
			parseVulnerabilityDetails(line, &result.Vulnerabilities)
		}

		// Check for warnings (lines starting with "npm WARN")
		if strings.HasPrefix(line, "npm WARN") {
			result.Warnings = append(result.Warnings, strings.TrimPrefix(line, "npm WARN "))
		}
	}
}

// parseVulnerabilityDetails extracts vulnerability counts by severity.
func parseVulnerabilityDetails(line string, vuln *VulnerabilitySummary) {
	matches := vulnDetailPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		count, _ := strconv.Atoi(match[1])
		switch match[2] {
		case "low":
			vuln.Low = count
		case "moderate":
			vuln.Moderate = count
		case "high":
			vuln.High = count
		case "critical":
			vuln.Critical = count
		}
	}
}

// Schema returns the JSON Schema for npm install output.
func (p *InstallParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *InstallParser) Matches(cmd string, subcommands []string) bool {
	if cmd != npmCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// npm install, npm i, or npm ci
	switch subcommands[0] {
	case "install", "i", "ci":
		return true
	default:
		return false
	}
}
