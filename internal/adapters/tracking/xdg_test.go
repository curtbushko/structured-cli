package tracking_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/tracking"
)

func TestDataHome_Default(t *testing.T) {
	// Save original value and unset
	originalValue := os.Getenv("XDG_DATA_HOME")
	os.Unsetenv("XDG_DATA_HOME")
	t.Cleanup(func() {
		if originalValue != "" {
			os.Setenv("XDG_DATA_HOME", originalValue)
		}
	})

	got := tracking.DataHome()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	want := filepath.Join(homeDir, ".local", "share")

	if got != want {
		t.Errorf("DataHome() = %q, want %q", got, want)
	}
}

func TestDataHome_Custom(t *testing.T) {
	// Save original value
	originalValue := os.Getenv("XDG_DATA_HOME")
	customPath := "/custom/data/path"
	os.Setenv("XDG_DATA_HOME", customPath)
	t.Cleanup(func() {
		if originalValue != "" {
			os.Setenv("XDG_DATA_HOME", originalValue)
		} else {
			os.Unsetenv("XDG_DATA_HOME")
		}
	})

	got := tracking.DataHome()

	if got != customPath {
		t.Errorf("DataHome() = %q, want %q", got, customPath)
	}
}

func TestDatabasePath(t *testing.T) {
	// Save original value and unset for predictable test
	originalValue := os.Getenv("XDG_DATA_HOME")
	os.Unsetenv("XDG_DATA_HOME")
	t.Cleanup(func() {
		if originalValue != "" {
			os.Setenv("XDG_DATA_HOME", originalValue)
		}
	})

	got := tracking.DatabasePath()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	want := filepath.Join(homeDir, ".local", "share", "structured-cli", "tracking.db")

	if got != want {
		t.Errorf("DatabasePath() = %q, want %q", got, want)
	}
}
