// Package main is the composition root for the structured-cli application.
// It wires all dependencies together and starts the application.
package main

import (
	"os"

	"github.com/curtbushko/structured-cli/internal/adapters/cli"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/build"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/docker"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/golang"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/lint"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/npm"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/test"
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

	// Register parsers
	registry.Register(git.NewStatusParser())
	registry.Register(git.NewLogParser())
	registry.Register(git.NewDiffParser())
	registry.Register(git.NewBranchParser())
	registry.Register(git.NewShowParser())
	registry.Register(git.NewAddParser())
	registry.Register(git.NewCommitParser())
	registry.Register(git.NewPushParser())
	registry.Register(git.NewPullParser())
	registry.Register(git.NewCheckoutParser())
	registry.Register(git.NewBlameParser())
	registry.Register(git.NewReflogParser())

	// Register Go parsers
	registry.Register(golang.NewBuildParser())
	registry.Register(golang.NewTestParser())
	registry.Register(golang.NewVetParser())
	registry.Register(golang.NewRunParser())
	registry.Register(golang.NewModTidyParser())
	registry.Register(golang.NewFmtParser())
	registry.Register(golang.NewGenerateParser())

	// Register build tool parsers
	registry.Register(build.NewTSCParser())
	registry.Register(build.NewESBuildParser())
	registry.Register(build.NewViteParser())
	registry.Register(build.NewWebpackParser())
	registry.Register(build.NewCargoParser())

	// Register linter parsers
	registry.Register(lint.NewESLintParser())
	registry.Register(lint.NewPrettierParser())
	registry.Register(lint.NewBiomeParser())
	registry.Register(lint.NewGolangCILintParser())
	registry.Register(lint.NewRuffParser())
	registry.Register(lint.NewMypyParser())

	// Register test runner parsers
	registry.Register(test.NewPytestParser())
	registry.Register(test.NewJestParser())
	registry.Register(test.NewVitestParser())
	registry.Register(test.NewMochaParser())
	registry.Register(test.NewCargoTestParser())

	// Register npm parsers
	registry.Register(npm.NewInstallParser())
	registry.Register(npm.NewAuditParser())
	registry.Register(npm.NewOutdatedParser())
	registry.Register(npm.NewListParser())
	registry.Register(npm.NewRunParser())
	registry.Register(npm.NewTestParser())
	registry.Register(npm.NewInitParser())

	// Register Docker parsers
	registry.Register(docker.NewPSParser())
	registry.Register(docker.NewBuildParser())
	registry.Register(docker.NewLogsParser())
	registry.Register(docker.NewImagesParser())
	registry.Register(docker.NewRunParser())
	registry.Register(docker.NewExecParser())
	registry.Register(docker.NewPullParser())
	registry.Register(docker.NewComposeUpParser())
	registry.Register(docker.NewComposeDownParser())
	registry.Register(docker.NewComposePSParser())

	// Create CLI handler (inbound adapter)
	handler := cli.NewHandler(execRunner, registry)

	// Execute the CLI and propagate exit code
	err := handler.Run()
	return domain.ExitCode(err)
}
