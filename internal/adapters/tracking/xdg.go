// Package tracking provides implementations of the Tracker interface for usage analytics.
package tracking

import (
	"os"
	"path/filepath"
)

// appName is the application name used in XDG directories.
const appName = "structured-cli"

// dbFileName is the name of the SQLite database file.
const dbFileName = "tracking.db"

// DataHome returns the XDG data home directory.
// It respects the XDG_DATA_HOME environment variable,
// falling back to ~/.local/share if not set.
func DataHome() string {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return dataHome
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home dir cannot be determined
		return filepath.Join(".", ".local", "share")
	}
	return filepath.Join(homeDir, ".local", "share")
}

// DatabasePath returns the full path to the tracking database.
// The path follows XDG Base Directory specification:
// $XDG_DATA_HOME/structured-cli/tracking.db
// or ~/.local/share/structured-cli/tracking.db if XDG_DATA_HOME is not set.
func DatabasePath() string {
	return filepath.Join(DataHome(), appName, dbFileName)
}
