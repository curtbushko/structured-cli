// Package domain contains the core business entities for the structured-cli.
// This layer has NO external dependencies - only stdlib.
package domain

import (
	"strings"
)

// Command represents a CLI command with its name, subcommands, and arguments.
// For example, "git log --oneline" would be:
// - Name: "git"
// - Subcommands: ["log"]
// - Args: ["--oneline"]
type Command struct {
	// Name is the base command (e.g., "git", "kubectl", "ls")
	Name string

	// Subcommands are the subcommand chain (e.g., ["config", "view"] for "kubectl config view")
	Subcommands []string

	// Args are the flags and arguments passed to the command
	Args []string
}

// NewCommand creates a new Command with the given name, subcommands, and arguments.
func NewCommand(name string, subcommands []string, args []string) Command {
	return Command{
		Name:        name,
		Subcommands: subcommands,
		Args:        args,
	}
}

// CommandFromArgs parses a slice of command-line arguments into a Command.
// The first element is the command name, subsequent elements that don't start
// with "-" are treated as subcommands until an argument starting with "-" is found,
// at which point all remaining elements are treated as args.
func CommandFromArgs(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, ErrEmptyCommand
	}

	name := args[0]
	var subcommands []string
	var cmdArgs []string

	// Parse remaining args: subcommands come before flags
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			// Once we hit a flag, everything else is args
			cmdArgs = args[i:]
			break
		}
		subcommands = append(subcommands, arg)
	}

	// If no flags were found, check if there are remaining positional args
	if cmdArgs == nil && len(args) > 1 {
		// All remaining were subcommands, no args
		subcommands = args[1:]
	}

	// Normalize empty slices to nil for consistent comparison
	if len(subcommands) == 0 {
		subcommands = nil
	}
	if len(cmdArgs) == 0 {
		cmdArgs = nil
	}

	return Command{
		Name:        name,
		Subcommands: subcommands,
		Args:        cmdArgs,
	}, nil
}

// FullCommand returns the complete command string including name, subcommands, and args.
func (c Command) FullCommand() string {
	parts := make([]string, 0, 1+len(c.Subcommands)+len(c.Args))
	parts = append(parts, c.Name)
	parts = append(parts, c.Subcommands...)
	parts = append(parts, c.Args...)
	return strings.Join(parts, " ")
}

// CommandSpec represents a specification for a CLI command that can be parsed.
// It defines what command and subcommand combination this spec handles.
type CommandSpec struct {
	// Name is the base command name (e.g., "git")
	Name string

	// Subcommand is the specific subcommand this spec handles (e.g., "status")
	Subcommand string

	// Description provides human-readable documentation for this command
	Description string
}
