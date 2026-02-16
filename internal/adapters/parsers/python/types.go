// Package python provides parsers for Python tool command output.
// This package is in the adapters layer and implements parsers for
// converting raw Python tool command output into structured domain types.
package python

// uvCommand is the uv command name constant.
const uvCommand = "uv"

// blackCommand is the black command name constant.
const blackCommand = "black"

// pipAuditCommand is the pip-audit command name constant.
const pipAuditCommand = "pip-audit"

// PipInstallResult represents the structured output of 'pip install'.
type PipInstallResult struct {
	// Success indicates whether the installation completed successfully.
	Success bool `json:"success"`

	// PackagesInstalled is the list of packages that were installed.
	PackagesInstalled []InstalledPackage `json:"packages_installed"`

	// RequirementsFile is the requirements file used, if any.
	RequirementsFile string `json:"requirements_file,omitempty"`

	// Warnings contains any warnings during installation.
	Warnings []string `json:"warnings"`

	// AlreadySatisfied contains packages that were already installed.
	AlreadySatisfied []string `json:"already_satisfied"`
}

// InstalledPackage represents a single installed package.
type InstalledPackage struct {
	// Name is the package name.
	Name string `json:"name"`

	// Version is the installed version.
	Version string `json:"version"`
}

// PipAuditResult represents the structured output of 'pip-audit'.
type PipAuditResult struct {
	// Success indicates whether the audit found no vulnerabilities.
	Success bool `json:"success"`

	// Vulnerabilities is the list of vulnerabilities found.
	Vulnerabilities []PipVulnerability `json:"vulnerabilities"`

	// PackagesScanned is the number of packages scanned.
	PackagesScanned int `json:"packages_scanned"`

	// Summary contains vulnerability count summary.
	Summary VulnerabilitySummary `json:"summary"`
}

// PipVulnerability represents a single vulnerability found by pip-audit.
type PipVulnerability struct {
	// Name is the package name with the vulnerability.
	Name string `json:"name"`

	// Version is the installed version with the vulnerability.
	Version string `json:"version"`

	// ID is the vulnerability identifier (e.g., CVE, GHSA).
	ID string `json:"id"`

	// Description is the vulnerability description.
	Description string `json:"description"`

	// FixVersions contains the versions that fix this vulnerability.
	FixVersions []string `json:"fix_versions,omitempty"`

	// Aliases contains other identifiers for this vulnerability.
	Aliases []string `json:"aliases,omitempty"`
}

// VulnerabilitySummary contains counts of vulnerabilities.
type VulnerabilitySummary struct {
	// Total is the total number of vulnerabilities found.
	Total int `json:"total"`

	// FixAvailable is the number of vulnerabilities with fixes available.
	FixAvailable int `json:"fix_available"`
}

// UVInstallResult represents the structured output of 'uv pip install'.
type UVInstallResult struct {
	// Success indicates whether the installation completed successfully.
	Success bool `json:"success"`

	// PackagesInstalled is the list of packages that were installed.
	PackagesInstalled []InstalledPackage `json:"packages_installed"`

	// PackagesUninstalled is the list of packages that were uninstalled.
	PackagesUninstalled []string `json:"packages_uninstalled"`

	// Cached is the number of packages installed from cache.
	Cached int `json:"cached"`

	// Downloaded is the number of packages downloaded.
	Downloaded int `json:"downloaded"`

	// Warnings contains any warnings during installation.
	Warnings []string `json:"warnings"`

	// AlreadySatisfied indicates requirements were already satisfied.
	AlreadySatisfied bool `json:"already_satisfied"`
}

// UVRunResult represents the structured output of 'uv run'.
type UVRunResult struct {
	// Success indicates whether the command ran successfully.
	Success bool `json:"success"`

	// Script is the script or command that was run.
	Script string `json:"script"`

	// Output is the command output.
	Output string `json:"output"`

	// InstalledPackages lists any packages installed before running.
	InstalledPackages []InstalledPackage `json:"installed_packages,omitempty"`

	// ExitCode is the exit code from the command.
	ExitCode int `json:"exit_code"`
}

// BlackResult represents the structured output of 'black --check'.
type BlackResult struct {
	// Success indicates whether all files are properly formatted.
	Success bool `json:"success"`

	// FilesChecked is the number of files checked.
	FilesChecked int `json:"files_checked"`

	// FilesWouldReformat lists files that would be reformatted.
	FilesWouldReformat []string `json:"files_would_reformat"`

	// FilesUnchanged is the number of files that are already formatted.
	FilesUnchanged int `json:"files_unchanged"`

	// Errors contains any errors encountered during checking.
	Errors []BlackError `json:"errors"`
}

// BlackError represents an error encountered by black.
type BlackError struct {
	// File is the file where the error occurred.
	File string `json:"file"`

	// Message is the error message.
	Message string `json:"message"`
}
