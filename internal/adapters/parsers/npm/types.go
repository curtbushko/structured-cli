// Package npm provides parsers for npm command output.
// This package is in the adapters layer and implements parsers for
// converting raw npm command output into structured domain types.
package npm

// npmCommand is the npm command name constant.
const npmCommand = "npm"

// InstallResult represents the structured output of 'npm install'.
type InstallResult struct {
	// Success indicates whether the installation completed successfully.
	Success bool `json:"success"`

	// PackagesAdded is the number of packages added.
	PackagesAdded int `json:"packages_added"`

	// PackagesRemoved is the number of packages removed.
	PackagesRemoved int `json:"packages_removed"`

	// PackagesChanged is the number of packages changed.
	PackagesChanged int `json:"packages_changed"`

	// PackagesAudited is the number of packages audited.
	PackagesAudited int `json:"packages_audited"`

	// Funding is the number of packages seeking funding.
	Funding int `json:"funding"`

	// Vulnerabilities contains vulnerability summary.
	Vulnerabilities VulnerabilitySummary `json:"vulnerabilities"`

	// Warnings contains any warnings during installation.
	Warnings []string `json:"warnings"`
}

// VulnerabilitySummary contains counts of vulnerabilities by severity.
type VulnerabilitySummary struct {
	// Total is the total number of vulnerabilities.
	Total int `json:"total"`

	// Low is the number of low severity vulnerabilities.
	Low int `json:"low"`

	// Moderate is the number of moderate severity vulnerabilities.
	Moderate int `json:"moderate"`

	// High is the number of high severity vulnerabilities.
	High int `json:"high"`

	// Critical is the number of critical severity vulnerabilities.
	Critical int `json:"critical"`
}

// AuditResult represents the structured output of 'npm audit'.
type AuditResult struct {
	// Success indicates whether the audit found no vulnerabilities.
	Success bool `json:"success"`

	// Vulnerabilities is the list of vulnerabilities found.
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`

	// Summary contains the vulnerability count summary.
	Summary VulnerabilitySummary `json:"summary"`

	// Metadata contains audit metadata.
	Metadata AuditMetadata `json:"metadata"`
}

// AuditMetadata contains metadata about the audit.
type AuditMetadata struct {
	// TotalDependencies is the number of dependencies scanned.
	TotalDependencies int `json:"total_dependencies"`

	// DevDependencies is the number of dev dependencies scanned.
	DevDependencies int `json:"dev_dependencies"`

	// OptionalDependencies is the number of optional dependencies scanned.
	OptionalDependencies int `json:"optional_dependencies"`
}

// Vulnerability represents a single security vulnerability.
type Vulnerability struct {
	// Name is the package name with the vulnerability.
	Name string `json:"name"`

	// Severity is the vulnerability severity (low, moderate, high, critical).
	Severity string `json:"severity"`

	// Title is a short description of the vulnerability.
	Title string `json:"title"`

	// URL is the URL for more information about the vulnerability.
	URL string `json:"url"`

	// Via contains the dependency path that introduces this vulnerability.
	Via []string `json:"via"`

	// Range is the affected version range.
	Range string `json:"range"`

	// FixAvailable indicates whether a fix is available.
	FixAvailable bool `json:"fix_available"`
}

// OutdatedResult represents the structured output of 'npm outdated'.
type OutdatedResult struct {
	// Success indicates whether all packages are up to date.
	Success bool `json:"success"`

	// Packages is the list of outdated packages.
	Packages []OutdatedPackage `json:"packages"`
}

// OutdatedPackage represents a single outdated package.
type OutdatedPackage struct {
	// Name is the package name.
	Name string `json:"name"`

	// Current is the currently installed version.
	Current string `json:"current"`

	// Wanted is the wanted version (based on semver range).
	Wanted string `json:"wanted"`

	// Latest is the latest available version.
	Latest string `json:"latest"`

	// Location is where the package is in the dependency tree.
	Location string `json:"location"`

	// Type is the dependency type (dependencies, devDependencies, etc.).
	Type string `json:"type"`
}

// ListResult represents the structured output of 'npm list'.
type ListResult struct {
	// Success indicates whether the list command succeeded.
	Success bool `json:"success"`

	// Name is the root package name.
	Name string `json:"name"`

	// Version is the root package version.
	Version string `json:"version"`

	// Dependencies is the list of direct dependencies.
	Dependencies []ListDependency `json:"dependencies"`

	// Problems contains any problems found.
	Problems []string `json:"problems"`
}

// ListDependency represents a dependency in the list output.
type ListDependency struct {
	// Name is the package name.
	Name string `json:"name"`

	// Version is the installed version.
	Version string `json:"version"`

	// Resolved is the resolved URL.
	Resolved string `json:"resolved,omitempty"`

	// Dependencies contains nested dependencies (for tree view).
	Dependencies []ListDependency `json:"dependencies,omitempty"`
}

// RunResult represents the structured output of 'npm run'.
type RunResult struct {
	// Success indicates whether the script ran successfully.
	Success bool `json:"success"`

	// Script is the script name that was run.
	Script string `json:"script"`

	// Output is the script output.
	Output string `json:"output"`

	// ExitCode is the exit code from the script.
	ExitCode int `json:"exit_code"`
}

// TestResult represents the structured output of 'npm test'.
type TestResult struct {
	// Success indicates whether tests passed.
	Success bool `json:"success"`

	// Output is the test output.
	Output string `json:"output"`

	// ExitCode is the exit code from the test.
	ExitCode int `json:"exit_code"`
}

// InitResult represents the structured output of 'npm init'.
type InitResult struct {
	// Success indicates whether initialization succeeded.
	Success bool `json:"success"`

	// PackageName is the created package name.
	PackageName string `json:"package_name"`

	// Version is the initial version.
	Version string `json:"version"`

	// Description is the package description.
	Description string `json:"description"`

	// EntryPoint is the main entry point.
	EntryPoint string `json:"entry_point"`

	// TestCommand is the test command.
	TestCommand string `json:"test_command"`

	// Repository is the git repository URL.
	Repository string `json:"repository"`

	// Keywords are the package keywords.
	Keywords []string `json:"keywords"`

	// Author is the package author.
	Author string `json:"author"`

	// License is the package license.
	License string `json:"license"`
}
