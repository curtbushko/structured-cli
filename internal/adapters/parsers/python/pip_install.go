package python

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing pip install output.
var (
	// successfullyInstalledPattern matches "Successfully installed pkg-1.0.0 pkg2-2.0.0"
	successfullyInstalledPattern = regexp.MustCompile(`Successfully installed (.+)`)

	// packageVersionPattern matches "package-1.2.3" format
	packageVersionPattern = regexp.MustCompile(`^(.+?)-(\d+(?:\.\d+)*)(?:\.(?:post|dev)\d+)?$`)

	// alreadySatisfiedPattern matches "Requirement already satisfied: package"
	alreadySatisfiedPattern = regexp.MustCompile(`Requirement already satisfied: (\S+)`)

	// requirementsFilePattern matches "from -r requirements.txt (line N)"
	requirementsFilePattern = regexp.MustCompile(`\(from -r ([^\s]+) \(line \d+\)\)`)

	// warningPattern matches pip warnings
	warningPattern = regexp.MustCompile(`^WARNING:\s*(.+)$`)
)

// PipInstallParser parses the output of 'pip install'.
type PipInstallParser struct {
	schema domain.Schema
}

// NewPipInstallParser creates a new PipInstallParser with the pip-install schema.
func NewPipInstallParser() *PipInstallParser {
	return &PipInstallParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/pip-install.json",
			"Pip Install Output",
			"object",
			map[string]domain.PropertySchema{
				"success":            {Type: "boolean", Description: "Whether installation completed successfully"},
				"packages_installed": {Type: "array", Description: "List of packages installed"},
				"requirements_file":  {Type: "string", Description: "Requirements file used, if any"},
				"warnings":           {Type: "array", Description: "Installation warnings"},
				"already_satisfied":  {Type: "array", Description: "Packages already satisfied"},
			},
			[]string{"success", "packages_installed"},
		),
	}
}

// Parse reads pip install output and returns structured data.
func (p *PipInstallParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &PipInstallResult{
		Success:           true,
		PackagesInstalled: []InstalledPackage{},
		Warnings:          []string{},
		AlreadySatisfied:  []string{},
	}

	parsePipInstallOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parsePipInstallOutput extracts install information from the output.
func parsePipInstallOutput(output string, result *PipInstallResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Parse "Successfully installed" line
		if matches := successfullyInstalledPattern.FindStringSubmatch(line); matches != nil {
			packages := strings.Fields(matches[1])
			for _, pkg := range packages {
				if pkgMatch := packageVersionPattern.FindStringSubmatch(pkg); pkgMatch != nil {
					result.PackagesInstalled = append(result.PackagesInstalled, InstalledPackage{
						Name:    pkgMatch[1],
						Version: pkgMatch[2],
					})
				}
			}
		}

		// Parse "Requirement already satisfied"
		if matches := alreadySatisfiedPattern.FindStringSubmatch(line); matches != nil {
			result.AlreadySatisfied = append(result.AlreadySatisfied, matches[1])
		}

		// Parse requirements file reference
		if matches := requirementsFilePattern.FindStringSubmatch(line); matches != nil {
			if result.RequirementsFile == "" {
				result.RequirementsFile = matches[1]
			}
		}

		// Parse warnings
		if matches := warningPattern.FindStringSubmatch(line); matches != nil {
			result.Warnings = append(result.Warnings, matches[1])
		}
	}
}

// Schema returns the JSON Schema for pip install output.
func (p *PipInstallParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PipInstallParser) Matches(cmd string, subcommands []string) bool {
	// Match pip or pip3
	if cmd != "pip" && cmd != "pip3" {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	return subcommands[0] == "install"
}
