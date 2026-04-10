package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/curtbushko/structured-cli/internal/ports"
)

// buildThemeCommand creates the 'theme' cobra parent command with 'list' and 'select' subcommands.
func buildThemeCommand(provider ports.ThemeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "theme",
		Short:         "Manage themes for structured-cli output",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(buildThemeListCommand(provider))
	cmd.AddCommand(buildThemeSelectCommand(provider))

	return cmd
}

// buildThemeListCommand creates the 'theme list' subcommand.
func buildThemeListCommand(provider ports.ThemeProvider) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List available themes",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeThemeListCommand(provider, jsonOutput, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

// buildThemeSelectCommand creates the 'theme select' subcommand.
func buildThemeSelectCommand(provider ports.ThemeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "select <name>",
		Short:         "Select the active theme",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeThemeSelectCommand(provider, args[0], cmd.OutOrStdout())
		},
	}

	return cmd
}

// executeThemeListCommand lists all available themes from the provider.
func executeThemeListCommand(provider ports.ThemeProvider, jsonOutput bool, out io.Writer) error {
	themes := provider.ListThemes()

	if jsonOutput {
		enc := json.NewEncoder(out)
		return enc.Encode(themes)
	}

	for _, name := range themes {
		if _, err := fmt.Fprintln(out, name); err != nil {
			return fmt.Errorf("write theme name: %w", err)
		}
	}
	return nil
}

// executeThemeSelectCommand validates and persists a theme selection.
func executeThemeSelectCommand(provider ports.ThemeProvider, name string, out io.Writer) error {
	// Validate the theme name against available themes
	themes := provider.ListThemes()
	valid := false
	for _, t := range themes {
		if t == name {
			valid = true
			break
		}
	}

	if !valid {
		available := strings.Join(themes, ", ")
		return fmt.Errorf("unknown theme %q; available themes: %s", name, available)
	}

	// Persist the theme selection
	if err := provider.SetTheme(name); err != nil {
		return fmt.Errorf("select theme: %w", err)
	}

	if _, err := fmt.Fprintf(out, "Theme selected: %q\n", name); err != nil {
		return fmt.Errorf("write confirmation: %w", err)
	}
	return nil
}
