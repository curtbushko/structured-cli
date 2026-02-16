// Package main is the composition root for the structured-cli application.
// It wires all dependencies together and starts the application.
package main

import (
	"os"

	"github.com/curtbushko/structured-cli/internal/adapters/cli"
	"github.com/curtbushko/structured-cli/internal/adapters/runner"
	"github.com/curtbushko/structured-cli/internal/application"
	"github.com/curtbushko/structured-cli/internal/domain"
)

func main() {
	os.Exit(run())
}

// run wires all dependencies and executes the CLI.
// Returns the exit code to propagate to the OS.
func run() int {
	// Create adapters (infrastructure layer)
	execRunner := runner.NewExecRunner()

	// Create application services
	registry := application.NewInMemoryParserRegistry()
	// Note: No parsers registered yet - all commands will use passthrough mode

	// Create CLI handler (inbound adapter)
	handler := cli.NewHandler(execRunner, registry)

	// Execute the CLI and propagate exit code
	err := handler.Run()
	return domain.ExitCode(err)
}
