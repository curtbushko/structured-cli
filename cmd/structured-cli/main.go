// Package main is the composition root for the structured-cli application.
// It wires all dependencies together and starts the application.
package main

import (
	"log"
	"os"

	"github.com/curtbushko/structured-cli/internal/adapters/cli"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/build"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/cargo"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/docker"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/golang"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/lint"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/makejust"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/npm"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/python"
	"github.com/curtbushko/structured-cli/internal/adapters/parsers/test"
	"github.com/curtbushko/structured-cli/internal/adapters/runner"
	"github.com/curtbushko/structured-cli/internal/adapters/tracking"
	"github.com/curtbushko/structured-cli/internal/application"
	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Environment variable to disable tracking.
const envNoTracking = "STRUCTURED_CLI_NO_TRACKING"

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

	// Register Python parsers
	registry.Register(python.NewPipInstallParser())
	registry.Register(python.NewPipAuditParser())
	registry.Register(python.NewUVInstallParser())
	registry.Register(python.NewUVRunParser())
	registry.Register(python.NewBlackParser())

	// Register Cargo parsers
	registry.Register(cargo.NewClippyParser())
	registry.Register(cargo.NewRunParser())
	registry.Register(cargo.NewAddParser())
	registry.Register(cargo.NewRemoveParser())
	registry.Register(cargo.NewFmtParser())
	registry.Register(cargo.NewDocParser())
	registry.Register(cargo.NewCheckParser())

	// Register Make/Just parsers
	registry.Register(makejust.NewMakeParser())
	registry.Register(makejust.NewJustParser())

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

	// Register file operations parsers
	registry.Register(fileops.NewLSParser())
	registry.Register(fileops.NewFindParser())
	registry.Register(fileops.NewGrepParser())
	registry.Register(fileops.NewRipgrepParser())
	registry.Register(fileops.NewFDParser())
	registry.Register(fileops.NewCatParser())
	registry.Register(fileops.NewHeadParser())
	registry.Register(fileops.NewTailParser())
	registry.Register(fileops.NewWCParser())
	registry.Register(fileops.NewDUParser())
	registry.Register(fileops.NewDFParser())

	// Create tracker (SQLite or NoOp based on environment)
	tracker := createTracker()
	defer func() {
		if err := tracker.Close(); err != nil {
			log.Printf("warning: failed to close tracker: %v", err)
		}
	}()

	// Create small filter with default patterns
	smallFilter := createSmallFilter()

	// Create CLI handler (inbound adapter) with tracker and small filter
	handler := cli.NewHandlerWithSmallFilter(execRunner, registry, tracker, smallFilter)

	// Execute the CLI and propagate exit code
	err := handler.Run()
	return domain.ExitCode(err)
}

// createTracker creates the appropriate tracker based on environment configuration.
// Returns NoOpTracker if STRUCTURED_CLI_NO_TRACKING is set, otherwise SQLiteTracker.
// Returns nil on SQLite creation errors (tracking disabled silently).
func createTracker() ports.Tracker {
	if os.Getenv(envNoTracking) != "" {
		return tracking.NewNoOpTracker()
	}

	// Get the default database path using XDG conventions
	dbPath := tracking.DatabasePath()

	tracker, err := tracking.NewSQLiteTracker(dbPath)
	if err != nil {
		// Log warning but don't fail - tracking is optional
		log.Printf("warning: failed to create tracker: %v", err)
		return tracking.NewNoOpTracker()
	}

	return tracker
}

// createSmallFilter creates a SmallFilter with default patterns for common commands.
// The filter detects terse output and returns compact JSON status.
func createSmallFilter() ports.SmallOutputFilter {
	filter := application.NewSmallFilter()

	// Register default patterns for common commands
	patterns := application.ToMinimalPatterns(application.DefaultPatterns())
	for _, p := range patterns {
		filter.RegisterPattern(p)
	}

	return filter
}
