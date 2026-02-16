package domain

// ParseResult represents the structured output from parsing CLI command output.
// It contains the parsed data, raw output, exit code, and any error that occurred.
type ParseResult struct {
	// Data is the structured data extracted from the command output.
	// The type depends on the parser and schema used.
	Data any

	// Raw is the original unprocessed command output.
	Raw string

	// ExitCode is the exit code returned by the command.
	ExitCode int

	// Error holds any error that occurred during parsing.
	// If Error is non-nil, Data may be nil or incomplete.
	Error error
}

// NewParseResult creates a new ParseResult with the given data, raw output, and exit code.
// The Error field is set to nil, indicating successful parsing.
func NewParseResult(data any, raw string, exitCode int) ParseResult {
	return ParseResult{
		Data:     data,
		Raw:      raw,
		ExitCode: exitCode,
		Error:    nil,
	}
}

// NewParseResultWithError creates a new ParseResult representing a parsing failure.
// Data is set to nil since parsing failed.
func NewParseResultWithError(err error, raw string, exitCode int) ParseResult {
	return ParseResult{
		Data:     nil,
		Raw:      raw,
		ExitCode: exitCode,
		Error:    err,
	}
}

// Success returns true if the parsing was successful.
// A result is considered successful if there is no error and the exit code is 0.
func (r ParseResult) Success() bool {
	return r.Error == nil && r.ExitCode == 0
}
