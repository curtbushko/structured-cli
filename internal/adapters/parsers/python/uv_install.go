package python

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing uv pip install output.
var (
	// uvInstalledPattern matches "+ package==version" or "+ package==version (cached)"
	uvInstalledPattern = regexp.MustCompile(`^\s*\+\s+(\S+)==(\S+?)(?:\s+\(cached\))?$`)

	// uvUninstalledPattern matches "- package==version"
	uvUninstalledPattern = regexp.MustCompile(`^\s*-\s+(\S+)==(\S+)$`)

	// uvCachedPattern matches packages with (cached) suffix
	uvCachedPattern = regexp.MustCompile(`\(cached\)$`)

	// uvAuditedPattern matches "Audited N package(s)"
	uvAuditedPattern = regexp.MustCompile(`^Audited \d+ packages?`)
)

// UVInstallParser parses the output of 'uv pip install'.
type UVInstallParser struct {
	schema domain.Schema
}

// NewUVInstallParser creates a new UVInstallParser with the uv-pip-install schema.
func NewUVInstallParser() *UVInstallParser {
	return &UVInstallParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/uv-pip-install.json",
			"UV Pip Install Output",
			"object",
			map[string]domain.PropertySchema{
				"success":              {Type: "boolean", Description: "Whether installation completed successfully"},
				"packages_installed":   {Type: "array", Description: "List of packages installed"},
				"packages_uninstalled": {Type: "array", Description: "List of packages uninstalled"},
				"cached":               {Type: "integer", Description: "Number of packages from cache"},
				"downloaded":           {Type: "integer", Description: "Number of packages downloaded"},
				"warnings":             {Type: "array", Description: "Installation warnings"},
				"already_satisfied":    {Type: "boolean", Description: "Requirements already satisfied"},
			},
			[]string{"success", "packages_installed"},
		),
	}
}

// Parse reads uv pip install output and returns structured data.
func (p *UVInstallParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &UVInstallResult{
		Success:             true,
		PackagesInstalled:   []InstalledPackage{},
		PackagesUninstalled: []string{},
		Warnings:            []string{},
	}

	parseUVInstallOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseUVInstallOutput extracts install information from uv output.
func parseUVInstallOutput(output string, result *UVInstallResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Check if requirements were already satisfied
		if uvAuditedPattern.MatchString(line) {
			result.AlreadySatisfied = true
			continue
		}

		// Parse installed packages
		if matches := uvInstalledPattern.FindStringSubmatch(line); matches != nil {
			pkg := InstalledPackage{
				Name:    matches[1],
				Version: matches[2],
			}
			result.PackagesInstalled = append(result.PackagesInstalled, pkg)

			// Check if cached
			if uvCachedPattern.MatchString(line) {
				result.Cached++
			} else {
				result.Downloaded++
			}
			continue
		}

		// Parse uninstalled packages
		if matches := uvUninstalledPattern.FindStringSubmatch(line); matches != nil {
			result.PackagesUninstalled = append(result.PackagesUninstalled, matches[1])
			continue
		}

		// Parse warnings
		if strings.HasPrefix(line, "warning:") || strings.HasPrefix(line, "Warning:") {
			result.Warnings = append(result.Warnings, strings.TrimPrefix(strings.TrimPrefix(line, "warning:"), "Warning:"))
		}
	}
}

// Schema returns the JSON Schema for uv pip install output.
func (p *UVInstallParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *UVInstallParser) Matches(cmd string, subcommands []string) bool {
	if cmd != uvCommand {
		return false
	}

	// Need at least "pip" and "install"
	if len(subcommands) < 2 {
		return false
	}

	return subcommands[0] == "pip" && subcommands[1] == "install"
}
