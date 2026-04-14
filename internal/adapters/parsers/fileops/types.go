// Package fileops provides parsers for file operation command output.
// This package is in the adapters layer and implements parsers for
// converting raw file operation output (ls, find, grep, etc.) into structured domain types.
//
// # Compact Format for Token Efficiency
//
// Several types in this package (MatchTuple, FileMatchGroup) use custom JSON
// marshaling to produce compact array-based output instead of verbose object
// format. This design choice optimizes for LLM token consumption:
//
// Object format (verbose):
//
//	{"file": "main.go", "line": 42, "content": "func main()"}
//
// Array format (compact):
//
//	["main.go", 1, [[42, "func main()"]]]
//
// The compact format reduces token usage by 60-80% for grep output with many
// matches, directly reducing API costs when results are fed to LLM APIs.
package fileops

import (
	"encoding/json"
	"errors"
	"fmt"
)

// LSEntry represents a single entry from ls output.
type LSEntry struct {
	// Name is the file or directory name.
	Name string `json:"name"`

	// Type indicates whether this is a file, directory, symlink, etc.
	// Valid values: file, directory, symlink, socket, fifo, block, char, unknown
	Type string `json:"type"`

	// Size is the file size in bytes (only for files).
	Size int64 `json:"size,omitempty"`

	// Permissions is the file permission string (e.g., "rwxr-xr-x").
	Permissions string `json:"permissions,omitempty"`

	// Owner is the file owner username.
	Owner string `json:"owner,omitempty"`

	// Group is the file group name.
	Group string `json:"group,omitempty"`

	// Modified is the last modification time in ISO 8601 format.
	Modified string `json:"modified,omitempty"`

	// Links is the number of hard links.
	Links int `json:"links,omitempty"`

	// Target is the symlink target path (only for symlinks).
	Target string `json:"target,omitempty"`
}

// LSOutput represents the structured output of 'ls'.
type LSOutput struct {
	// Entries is the list of directory entries.
	Entries []LSEntry `json:"entries"`

	// Total is the total block size (from ls -l header).
	Total int64 `json:"total,omitempty"`
}

// FindOutput represents the structured output of 'find'.
type FindOutput struct {
	// Files is the list of matching file paths.
	Files []string `json:"files"`

	// Count is the number of matches found.
	Count int `json:"count"`
}

// GrepMatch represents a single match from grep output.
type GrepMatch struct {
	// File is the file containing the match.
	File string `json:"file"`

	// Line is the line number of the match.
	Line int `json:"line,omitempty"`

	// Content is the matching line content.
	Content string `json:"content"`

	// Column is the column where the match starts (if available).
	Column int `json:"column,omitempty"`
}

// GrepOutput represents the structured output of 'grep'.
type GrepOutput struct {
	// Matches is the list of matches found.
	Matches []GrepMatch `json:"matches"`

	// Count is the total number of matches.
	Count int `json:"count"`

	// FilesMatched is the number of files with matches.
	FilesMatched int `json:"filesMatched"`
}

// RipgrepMatch represents a single match from rg output.
type RipgrepMatch struct {
	// File is the file containing the match.
	File string `json:"file"`

	// Line is the line number of the match.
	Line int `json:"line"`

	// Column is the column where the match starts.
	Column int `json:"column"`

	// Content is the matching line content.
	Content string `json:"content"`

	// MatchText is the specific text that matched.
	MatchText string `json:"matchText,omitempty"`
}

// RipgrepOutput represents the structured output of 'rg' (ripgrep).
type RipgrepOutput struct {
	// Matches is the list of matches found.
	Matches []RipgrepMatch `json:"matches"`

	// Count is the total number of matches.
	Count int `json:"count"`

	// FilesMatched is the number of files with matches.
	FilesMatched int `json:"filesMatched"`

	// Stats contains search statistics.
	Stats *RipgrepStats `json:"stats,omitempty"`
}

// RipgrepStats contains ripgrep search statistics.
type RipgrepStats struct {
	// FilesSearched is the number of files searched.
	FilesSearched int `json:"filesSearched"`

	// BytesSearched is the number of bytes searched.
	BytesSearched int64 `json:"bytesSearched"`

	// Duration is the search duration (if --stats was used).
	Duration string `json:"duration,omitempty"`
}

// FDOutput represents the structured output of 'fd'.
type FDOutput struct {
	// Files is the list of matching file paths.
	Files []string `json:"files"`

	// Count is the number of matches found.
	Count int `json:"count"`
}

// CatOutput represents the structured output of 'cat'.
type CatOutput struct {
	// Content is the file content.
	Content string `json:"content"`

	// Lines is the line count (if -n flag was used).
	Lines int `json:"lines,omitempty"`

	// Bytes is the byte count.
	Bytes int `json:"bytes"`
}

// HeadTailOutput represents the structured output of 'head' or 'tail'.
type HeadTailOutput struct {
	// Content is the file content (first/last N lines).
	Content string `json:"content"`

	// Lines is the list of lines read.
	Lines []string `json:"lines"`

	// LineCount is the number of lines returned.
	LineCount int `json:"lineCount"`
}

// WCStats represents word count statistics for a single file.
type WCStats struct {
	// File is the filename.
	File string `json:"file"`

	// Lines is the number of newline characters.
	Lines int `json:"lines"`

	// Words is the number of words.
	Words int `json:"words"`

	// Chars is the number of characters.
	Chars int `json:"chars"`

	// Bytes is the number of bytes.
	Bytes int `json:"bytes"`
}

// WCOutput represents the structured output of 'wc'.
type WCOutput struct {
	// Files is the list of file statistics.
	Files []WCStats `json:"files"`

	// Total contains totals when multiple files are counted.
	Total *WCStats `json:"total,omitempty"`
}

// DUEntry represents disk usage for a single path.
type DUEntry struct {
	// Path is the file or directory path.
	Path string `json:"path"`

	// Size is the disk usage in bytes.
	Size int64 `json:"size"`

	// SizeHuman is the human-readable size (e.g., "4.5M").
	SizeHuman string `json:"sizeHuman,omitempty"`
}

// DUOutput represents the structured output of 'du'.
type DUOutput struct {
	// Entries is the list of disk usage entries.
	Entries []DUEntry `json:"entries"`

	// Total is the total disk usage in bytes.
	Total int64 `json:"total,omitempty"`

	// TotalHuman is the human-readable total size.
	TotalHuman string `json:"totalHuman,omitempty"`
}

// DFEntry represents disk space for a single filesystem.
type DFEntry struct {
	// Filesystem is the filesystem name/device.
	Filesystem string `json:"filesystem"`

	// Type is the filesystem type (e.g., "ext4", "tmpfs").
	Type string `json:"type,omitempty"`

	// Size is the total size in bytes.
	Size int64 `json:"size"`

	// Used is the used space in bytes.
	Used int64 `json:"used"`

	// Available is the available space in bytes.
	Available int64 `json:"available"`

	// UsePercent is the usage percentage.
	UsePercent float64 `json:"usePercent"`

	// MountedOn is the mount point.
	MountedOn string `json:"mountedOn"`

	// SizeHuman is the human-readable total size.
	SizeHuman string `json:"sizeHuman,omitempty"`

	// UsedHuman is the human-readable used size.
	UsedHuman string `json:"usedHuman,omitempty"`

	// AvailableHuman is the human-readable available size.
	AvailableHuman string `json:"availableHuman,omitempty"`
}

// DFOutput represents the structured output of 'df'.
type DFOutput struct {
	// Filesystems is the list of filesystem entries.
	Filesystems []DFEntry `json:"filesystems"`
}

// MatchTuple represents a single grep match as a [line, content] tuple.
//
// This type uses custom JSON marshaling to produce a compact 2-element array
// instead of an object with named fields. This saves ~30 tokens per match
// compared to {"line": N, "content": "..."} format.
//
// JSON format: [lineNumber, "matchingContent"]
// Example: [42, "func main() {"]
//
// The positional array format is unambiguous because:
//   - Element 0 is always the line number (integer)
//   - Element 1 is always the content (string)
type MatchTuple struct {
	// Line is the 1-based line number where the match was found.
	Line int
	// Content is the full text of the matching line.
	Content string
}

// MarshalJSON implements json.Marshaler for MatchTuple.
// Marshals as a 2-element array: [line, "content"].
func (m MatchTuple) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{m.Line, m.Content})
}

// UnmarshalJSON implements json.Unmarshaler for MatchTuple.
// Expects a 2-element array: [line, "content"].
func (m *MatchTuple) UnmarshalJSON(data []byte) error {
	var arr []any
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) != 2 {
		return fmt.Errorf("MatchTuple: expected 2 elements, got %d", len(arr))
	}
	line, ok := arr[0].(float64)
	if !ok {
		return errors.New("MatchTuple: line must be a number")
	}
	content, ok := arr[1].(string)
	if !ok {
		return errors.New("MatchTuple: content must be a string")
	}
	m.Line = int(line)
	m.Content = content
	return nil
}

// FileMatchGroup represents all matches in a single file as a compact array.
//
// This type uses custom JSON marshaling to produce a 3-element array instead
// of an object. By grouping matches by file and using the array format, the
// filename appears only once per file rather than repeated for each match.
//
// JSON format: ["filename", totalMatchCount, [[line, content], ...]]
// Example: ["main.go", 3, [[10, "import fmt"], [42, "func main()"], [50, "fmt.Println"]]]
//
// Token efficiency breakdown:
//   - Object format: ~45 tokens per match (filename repeated)
//   - Array format: ~15 tokens per file + ~8 tokens per match
//   - For 10 matches: 450 tokens (object) vs 95 tokens (array) = 79% savings
//
// The Count field preserves the original match count even when Matches is
// truncated, allowing consumers to know how many matches were omitted.
type FileMatchGroup struct {
	// Filename is the path to the file containing matches.
	Filename string
	// Count is the total number of matches found in this file (may exceed len(Matches) if truncated).
	Count int
	// Matches is the list of match tuples (may be truncated per maxMatchesPerFile).
	Matches []MatchTuple
}

// MarshalJSON implements json.Marshaler for FileMatchGroup.
// Marshals as a 3-element array: ["filename", count, [[line, content], ...]].
func (f FileMatchGroup) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{f.Filename, f.Count, f.Matches})
}

// UnmarshalJSON implements json.Unmarshaler for FileMatchGroup.
// Expects a 3-element array: ["filename", count, [[line, content], ...]].
func (f *FileMatchGroup) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	if len(arr) != 3 {
		return fmt.Errorf("FileMatchGroup: expected 3 elements, got %d", len(arr))
	}
	if err := json.Unmarshal(arr[0], &f.Filename); err != nil {
		return fmt.Errorf("FileMatchGroup: filename: %w", err)
	}
	if err := json.Unmarshal(arr[1], &f.Count); err != nil {
		return fmt.Errorf("FileMatchGroup: count: %w", err)
	}
	if err := json.Unmarshal(arr[2], &f.Matches); err != nil {
		return fmt.Errorf("FileMatchGroup: matches: %w", err)
	}
	return nil
}

// GrepOutputCompact represents grep output in a compact array-based format.
//
// This is the top-level output structure for the grep parser. It provides
// summary statistics (Total, Files) alongside the detailed Results, allowing
// consumers to quickly assess the scope of matches without iterating.
//
// The compact format is specifically designed for LLM consumption where token
// efficiency directly impacts API costs and context window utilization. The
// array-based Results format (via FileMatchGroup and MatchTuple) reduces
// token usage by 60-80% compared to traditional object-per-match formats.
//
// Example output:
//
//	{
//	  "total": 25,
//	  "files": 3,
//	  "results": [
//	    ["main.go", 10, [[1, "package main"], [5, "func main()"]]],
//	    ["util.go", 15, [[12, "func helper()"]]]
//	  ],
//	  "truncated": true
//	}
type GrepOutputCompact struct {
	// Total is the total number of matches found across all files (before truncation).
	Total int `json:"total"`
	// Files is the number of unique files containing matches.
	Files int `json:"files"`
	// Results contains the grouped matches per file in compact array format.
	Results []FileMatchGroup `json:"results"`
	// Truncated is true if results were limited due to maxMatchesPerFile or maxTotalMatches.
	Truncated bool `json:"truncated"`
}

// File type constants for ls output.
const (
	TypeFile      = "file"
	TypeDirectory = "directory"
	TypeSymlink   = "symlink"
	TypeSocket    = "socket"
	TypeFIFO      = "fifo"
	TypeBlock     = "block"
	TypeChar      = "char"
	TypeUnknown   = "unknown"
)
