package domain

// FallbackResult represents the output from an unsupported command.
// When no parser is registered for a command, we still execute it and
// return the raw output wrapped in this structure. This enables JSON mode
// to always return valid JSON even for unrecognized commands.
//
// In JSON mode, this produces: {"raw": "...", "parsed": false, "exitCode": N}
// In passthrough mode, only the raw output is returned.
type FallbackResult struct {
	// Raw is the original unprocessed command output.
	Raw string `json:"raw"`

	// Parsed is always false for fallback results, indicating that
	// the output was not parsed into a structured format.
	Parsed bool `json:"parsed"`

	// ExitCode is the exit code returned by the command.
	ExitCode int `json:"exitCode"`
}

// NewFallbackResult creates a new FallbackResult with the given raw output
// and exit code. The Parsed field is always set to false.
func NewFallbackResult(raw string, exitCode int) FallbackResult {
	return FallbackResult{
		Raw:      raw,
		Parsed:   false,
		ExitCode: exitCode,
	}
}
