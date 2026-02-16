// Package makejust provides parsers for the make and just build tools.
// It parses output from GNU make and the just command runner to produce
// structured results indicating success, errors, targets/recipes, and commands.
package makejust

// Result represents the structured output of 'make' command.
// It captures build success, target, duration, errors, and related metadata.
type Result struct {
	// Success indicates whether the make command completed without errors.
	Success bool `json:"success"`

	// Target is the target that was built (empty for default target).
	Target string `json:"target,omitempty"`

	// Duration is the time taken to build in milliseconds.
	Duration int64 `json:"duration,omitempty"`

	// Error contains the error message if the build failed.
	Error string `json:"error,omitempty"`

	// ExitCode is the exit code from make.
	ExitCode int `json:"exit_code"`

	// Targets contains the list of available targets (for make -p or parsing).
	Targets []Target `json:"targets,omitempty"`

	// Commands contains the commands that would be run (for dry run mode -n).
	Commands []string `json:"commands,omitempty"`
}

// Target represents a single make target.
type Target struct {
	// Name is the target name.
	Name string `json:"name"`

	// Description is the target description if available.
	Description string `json:"description,omitempty"`

	// Dependencies lists the target's prerequisites.
	Dependencies []string `json:"dependencies,omitempty"`
}

// JustResult represents the structured output of 'just' command runner.
// It captures recipe execution results, listings, and dry run output.
type JustResult struct {
	// Success indicates whether the just command completed without errors.
	Success bool `json:"success"`

	// Recipe is the recipe that was executed (empty for default).
	Recipe string `json:"recipe,omitempty"`

	// Duration is the time taken to execute in milliseconds.
	Duration int64 `json:"duration,omitempty"`

	// Error contains the error message if execution failed.
	Error string `json:"error,omitempty"`

	// ExitCode is the exit code from just.
	ExitCode int `json:"exit_code"`

	// Recipes contains the list of available recipes (for just --list).
	Recipes []JustRecipe `json:"recipes,omitempty"`

	// Commands contains the commands that would be run (for dry run mode -n).
	Commands []string `json:"commands,omitempty"`
}

// JustRecipe represents a single just recipe.
type JustRecipe struct {
	// Name is the recipe name.
	Name string `json:"name"`

	// Description is the recipe description from comments.
	Description string `json:"description,omitempty"`

	// Parameters contains the recipe's parameters.
	Parameters []JustParameter `json:"parameters,omitempty"`
}

// JustParameter represents a parameter for a just recipe.
type JustParameter struct {
	// Name is the parameter name.
	Name string `json:"name"`

	// Default is the default value if any.
	Default string `json:"default,omitempty"`
}
