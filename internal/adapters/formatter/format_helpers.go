// Package formatter provides rendering adapters for formatted statistics display.
// This adapter package renders statistics using lipgloss styling with flair theme colors.
package formatter

import (
	"fmt"
	"strconv"
	"time"
)

const (
	billion  = 1_000_000_000
	million  = 1_000_000
	thousand = 1_000
)

// FormatLargeNumber formats an integer with K/M/B suffixes for readability.
// Values below 1000 are returned as-is.
func FormatLargeNumber(n int) string {
	switch {
	case n >= billion:
		return fmt.Sprintf("%.1fB", float64(n)/float64(billion))
	case n >= million:
		return fmt.Sprintf("%.1fM", float64(n)/float64(million))
	case n >= thousand:
		return fmt.Sprintf("%.1fK", float64(n)/float64(thousand))
	default:
		return strconv.Itoa(n)
	}
}

// FormatDuration formats a time.Duration into a human-readable string.
// Examples: "12m20s", "696ms", "1h5m", "<1ms".
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "0ms"
	}

	if d < time.Millisecond {
		return "<1ms"
	}

	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	switch {
	case hours > 0 && minutes > 0:
		return fmt.Sprintf("%dh%dm", hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh", hours)
	case minutes > 0 && seconds > 0:
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	case minutes > 0:
		return fmt.Sprintf("%dm", minutes)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// TruncateCommand truncates a command string to maxLen characters,
// adding "..." suffix if truncated.
func TruncateCommand(cmd string, maxLen int) string {
	if len(cmd) <= maxLen {
		return cmd
	}

	ellipsis := "..."
	if maxLen <= len(ellipsis) {
		return ellipsis[:maxLen]
	}

	return cmd[:maxLen-len(ellipsis)] + ellipsis
}
