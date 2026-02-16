// Package fileops provides parsers for file operation command output.
// This package is in the adapters layer and implements parsers for
// converting raw file operation output (ls, find, grep, etc.) into structured domain types.
package fileops

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
