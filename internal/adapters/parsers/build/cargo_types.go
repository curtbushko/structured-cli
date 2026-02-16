package build

// CargoResult represents the structured output of 'cargo build --message-format=json'.
// It captures build success, errors, warnings, compiled artifacts, and build scripts.
type CargoResult struct {
	// Success indicates whether the build completed without errors.
	Success bool `json:"success"`

	// Errors contains any compilation errors that occurred.
	Errors []CargoError `json:"errors"`

	// Warnings contains any compilation warnings that occurred.
	Warnings []CargoWarning `json:"warnings"`

	// Artifacts contains information about compiled artifacts.
	Artifacts []CargoArtifact `json:"artifacts"`

	// BuildScripts contains output from build script execution.
	BuildScripts []CargoBuildScript `json:"build_scripts"`
}

// CargoError represents a single compilation error from rustc.
type CargoError struct {
	// Message is the error message from the compiler.
	Message string `json:"message"`

	// Code is the error code (e.g., E0425).
	Code string `json:"code"`

	// File is the source file where the error occurred.
	File string `json:"file"`

	// Line is the line number where the error occurred.
	Line int `json:"line"`

	// Column is the column number where the error occurred.
	Column int `json:"column"`

	// Rendered is the human-readable rendered error message.
	Rendered string `json:"rendered"`
}

// CargoWarning represents a single compilation warning from rustc.
type CargoWarning struct {
	// Message is the warning message from the compiler.
	Message string `json:"message"`

	// Code is the warning/lint code (e.g., unused_variables).
	Code string `json:"code"`

	// File is the source file where the warning occurred.
	File string `json:"file"`

	// Line is the line number where the warning occurred.
	Line int `json:"line"`

	// Column is the column number where the warning occurred.
	Column int `json:"column"`

	// Rendered is the human-readable rendered warning message.
	Rendered string `json:"rendered"`
}

// CargoArtifact represents a compiled artifact from cargo build.
type CargoArtifact struct {
	// PackageID is the unique identifier for the package.
	PackageID string `json:"package_id"`

	// Target contains information about the compilation target.
	Target CargoTarget `json:"target"`

	// Profile contains the compilation profile settings.
	Profile CargoProfile `json:"profile"`

	// Features contains the enabled features for this artifact.
	Features []string `json:"features"`

	// Filenames contains the paths to generated artifact files.
	Filenames []string `json:"filenames"`

	// Executable is the path to the generated executable, if applicable.
	Executable string `json:"executable,omitempty"`

	// Fresh indicates whether the artifact was already up to date.
	Fresh bool `json:"fresh"`
}

// CargoTarget contains information about a build target.
type CargoTarget struct {
	// Kind is the target kind (e.g., ["lib"], ["bin"], ["example"]).
	Kind []string `json:"kind"`

	// CrateTypes is the list of crate types (e.g., ["lib"], ["bin"]).
	CrateTypes []string `json:"crate_types"`

	// Name is the name of the target.
	Name string `json:"name"`

	// SrcPath is the path to the target's source file.
	SrcPath string `json:"src_path"`

	// Edition is the Rust edition (e.g., "2018", "2021").
	Edition string `json:"edition"`
}

// CargoProfile contains compilation profile settings.
type CargoProfile struct {
	// OptLevel is the optimization level (e.g., "0", "1", "2", "3", "s", "z").
	OptLevel string `json:"opt_level"`

	// Debuginfo is the debug info level (0, 1, or 2).
	Debuginfo int `json:"debuginfo"`

	// DebugAssertions indicates whether debug assertions are enabled.
	DebugAssertions bool `json:"debug_assertions"`

	// OverflowChecks indicates whether overflow checks are enabled.
	OverflowChecks bool `json:"overflow_checks"`

	// Test indicates whether this is a test build.
	Test bool `json:"test"`
}

// CargoBuildScript represents output from a build script execution.
type CargoBuildScript struct {
	// PackageID is the unique identifier for the package.
	PackageID string `json:"package_id"`

	// LinkedLibs contains libraries to link.
	LinkedLibs []string `json:"linked_libs"`

	// LinkedPaths contains paths to search for libraries.
	LinkedPaths []string `json:"linked_paths"`

	// Cfgs contains configuration flags to pass to the compiler.
	Cfgs []string `json:"cfgs"`

	// Env contains environment variables to set.
	Env [][]string `json:"env"`

	// OutDir is the path to the build script's output directory.
	OutDir string `json:"out_dir"`
}
