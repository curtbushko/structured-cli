package git

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// StatusParser parses the output of 'git status --porcelain=v2 --branch'.
type StatusParser struct {
	schema domain.Schema
}

// NewStatusParser creates a new StatusParser with the git-status schema.
func NewStatusParser() *StatusParser {
	return &StatusParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-status.json",
			"Git Status Output",
			"object",
			map[string]domain.PropertySchema{
				"branch":    {Type: "string", Description: "Current branch name"},
				"upstream":  {Type: "string", Description: "Upstream branch name"},
				"ahead":     {Type: "integer", Description: "Commits ahead of upstream"},
				"behind":    {Type: "integer", Description: "Commits behind upstream"},
				"staged":    {Type: "array", Description: "Files staged for commit"},
				"modified":  {Type: "array", Description: "Files modified in worktree"},
				"deleted":   {Type: "array", Description: "Files deleted in worktree"},
				"untracked": {Type: "array", Description: "Untracked files"},
				"conflicts": {Type: "array", Description: "Files with merge conflicts"},
				"clean":     {Type: "boolean", Description: "True if working tree is clean"},
			},
			[]string{"branch", "staged", "modified", "deleted", "untracked", "conflicts", "clean"},
		),
	}
}

// Parse reads git status porcelain v2 output and returns structured data.
func (p *StatusParser) Parse(r io.Reader) (domain.ParseResult, error) {
	status := &Status{
		Staged:    []StagedFile{},
		Modified:  []string{},
		Deleted:   []string{},
		Untracked: []string{},
		Conflicts: []string{},
		Clean:     true,
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		if err := p.parseLine(line, status); err != nil {
			return domain.NewParseResultWithError(
				fmt.Errorf("parse line: %w", err),
				rawBuilder.String(),
				0,
			), nil
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			rawBuilder.String(),
			0,
		), nil
	}

	// Determine if repository is clean
	status.Clean = len(status.Staged) == 0 &&
		len(status.Modified) == 0 &&
		len(status.Deleted) == 0 &&
		len(status.Untracked) == 0 &&
		len(status.Conflicts) == 0

	return domain.NewParseResult(status, rawBuilder.String(), 0), nil
}

// Schema returns the JSON Schema for git status output.
func (p *StatusParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *StatusParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "git" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "status"
}

// parseLine parses a single line of porcelain v2 output.
func (p *StatusParser) parseLine(line string, status *Status) error {
	if len(line) == 0 {
		return nil
	}

	switch {
	case strings.HasPrefix(line, "# branch.head "):
		status.Branch = strings.TrimPrefix(line, "# branch.head ")

	case strings.HasPrefix(line, "# branch.upstream "):
		upstream := strings.TrimPrefix(line, "# branch.upstream ")
		status.Upstream = &upstream

	case strings.HasPrefix(line, "# branch.ab "):
		if err := p.parseAheadBehind(line, status); err != nil {
			return err
		}

	case strings.HasPrefix(line, "# branch.oid "):
		// OID is just metadata, skip

	case strings.HasPrefix(line, "1 "):
		// Ordinary changed entry
		return p.parseOrdinaryEntry(line, status)

	case strings.HasPrefix(line, "2 "):
		// Renamed or copied entry
		return p.parseRenamedEntry(line, status)

	case strings.HasPrefix(line, "u "):
		// Unmerged entry (conflict)
		return p.parseUnmergedEntry(line, status)

	case strings.HasPrefix(line, "? "):
		// Untracked file
		path := strings.TrimPrefix(line, "? ")
		status.Untracked = append(status.Untracked, path)

		// Note: "! " prefix (ignored files) is intentionally not handled
		// as they are not relevant to status output
	}

	return nil
}

// parseAheadBehind parses "# branch.ab +N -M" line.
func (p *StatusParser) parseAheadBehind(line string, status *Status) error {
	// Format: "# branch.ab +<ahead> -<behind>"
	parts := strings.Fields(line)
	if len(parts) != 4 {
		return fmt.Errorf("invalid branch.ab format: %s", line)
	}

	aheadStr := strings.TrimPrefix(parts[2], "+")
	ahead, err := strconv.Atoi(aheadStr)
	if err != nil {
		return fmt.Errorf("parse ahead count: %w", err)
	}
	status.Ahead = ahead

	behindStr := strings.TrimPrefix(parts[3], "-")
	behind, err := strconv.Atoi(behindStr)
	if err != nil {
		return fmt.Errorf("parse behind count: %w", err)
	}
	status.Behind = behind

	return nil
}

// parseOrdinaryEntry parses a "1 XY ..." line (ordinary changed entry).
// Format: 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
func (p *StatusParser) parseOrdinaryEntry(line string, status *Status) error {
	parts := strings.Fields(line)
	if len(parts) < 9 {
		return fmt.Errorf("invalid ordinary entry: %s", line)
	}

	xy := parts[1]
	path := parts[8]

	// X = staged status, Y = worktree status
	x := xy[0]
	y := xy[1]

	// Handle staged changes (X column)
	if x != '.' {
		staged := StagedFile{File: path, Status: statusCodeToString(x)}
		status.Staged = append(status.Staged, staged)
	}

	// Handle worktree changes (Y column)
	switch y {
	case 'M':
		status.Modified = append(status.Modified, path)
	case 'D':
		status.Deleted = append(status.Deleted, path)
	}

	return nil
}

// parseRenamedEntry parses a "2 XY ..." line (renamed/copied entry).
// Format: 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <X><score> <path><TAB><origPath>
func (p *StatusParser) parseRenamedEntry(line string, status *Status) error {
	parts := strings.Fields(line)
	if len(parts) < 10 {
		return fmt.Errorf("invalid renamed entry: %s", line)
	}

	xy := parts[1]
	x := xy[0]

	// The path is in the last field, which may contain a tab separator
	// Find the path portion after the score (R100, C100, etc.)
	pathParts := parts[9]
	// Split by tab to get new path and original path
	paths := strings.Split(pathParts, "\t")
	newPath := paths[0]

	// Handle staged renames/copies (X column)
	switch x {
	case 'R':
		staged := StagedFile{File: newPath, Status: statusRenamed}
		status.Staged = append(status.Staged, staged)
	case 'C':
		staged := StagedFile{File: newPath, Status: statusCopied}
		status.Staged = append(status.Staged, staged)
	}

	return nil
}

// parseUnmergedEntry parses a "u XY ..." line (unmerged/conflict entry).
// Format: u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
func (p *StatusParser) parseUnmergedEntry(line string, status *Status) error {
	parts := strings.Fields(line)
	if len(parts) < 11 {
		return fmt.Errorf("invalid unmerged entry: %s", line)
	}

	path := parts[10]
	status.Conflicts = append(status.Conflicts, path)

	return nil
}

// statusCodeToString converts a git status code to a human-readable string.
func statusCodeToString(code byte) string {
	switch code {
	case 'A':
		return statusAdded
	case 'M':
		return statusModified
	case 'D':
		return statusDeleted
	case 'R':
		return statusRenamed
	case 'C':
		return statusCopied
	default:
		return statusUnknown
	}
}
