package git

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// BranchParser parses the output of 'git branch' command.
type BranchParser struct {
	schema domain.Schema
}

// trackingInfoRegex matches [origin/branch: ahead N, behind M] or [origin/branch]
var trackingInfoRegex = regexp.MustCompile(`\[([\w/\-\.]+)(?:\s*:\s*(.+?))?\]`)

// aheadBehindRegex matches "ahead N" or "behind N" or "gone"
var aheadBehindRegex = regexp.MustCompile(`(ahead|behind)\s+(\d+)`)

// NewBranchParser creates a new BranchParser with the git-branch schema.
func NewBranchParser() *BranchParser {
	return &BranchParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-branch.json",
			"Git Branch Output",
			"object",
			map[string]domain.PropertySchema{
				"branches": {Type: "array", Description: "List of branches"},
				"current":  {Type: "string", Description: "Name of the current branch"},
			},
			[]string{"branches", "current"},
		),
	}
}

// Parse reads git branch output and returns structured data.
func (p *BranchParser) Parse(r io.Reader) (domain.ParseResult, error) {
	branchList := &BranchList{
		Branches: []Branch{},
		Current:  "",
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		branch := p.parseBranchLine(line)
		if branch != nil {
			branchList.Branches = append(branchList.Branches, *branch)
			if branch.Current {
				branchList.Current = branch.Name
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(branchList, rawBuilder.String(), 0), nil
}

// parseBranchLine parses a single line of git branch output.
// Line formats:
// - Simple: "* main" or "  feature"
// - Verbose: "* main abc1234 commit message"
// - With tracking: "* main [origin/main: ahead 2, behind 1]"
// - Detached: "* (HEAD detached at abc123)"
func (p *BranchParser) parseBranchLine(line string) *Branch {
	if len(line) < 2 {
		return nil
	}

	branch := &Branch{}

	// Check if this is the current branch
	if line[0] == '*' {
		branch.Current = true
		line = line[2:] // Remove "* " prefix
	} else {
		branch.Current = false
		line = strings.TrimLeft(line, " ") // Remove leading spaces
	}

	// Handle detached HEAD case
	if strings.HasPrefix(line, "(HEAD detached") {
		// Extract the full detached HEAD name including closing paren
		endParen := strings.Index(line, ")")
		if endParen > 0 {
			branch.Name = line[:endParen+1]
		} else {
			branch.Name = line
		}
		return branch
	}

	// Check for tracking info in brackets
	trackingMatch := trackingInfoRegex.FindStringSubmatchIndex(line)
	if trackingMatch != nil {
		// Extract tracking info
		upstream := line[trackingMatch[2]:trackingMatch[3]]
		branch.Upstream = &upstream

		// Parse ahead/behind if present
		if trackingMatch[4] != -1 && trackingMatch[5] != -1 {
			trackingStatus := line[trackingMatch[4]:trackingMatch[5]]
			p.parseAheadBehind(trackingStatus, branch)
		}

		// Remove the tracking info from the line for further parsing
		line = strings.TrimSpace(line[:trackingMatch[0]]) + strings.TrimSpace(line[trackingMatch[1]:])
	}

	// Now parse what's left: branch_name [hash] [commit message]
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	branch.Name = parts[0]

	// Check if second part looks like a commit hash (7+ hex chars)
	if len(parts) >= 2 && isCommitHash(parts[1]) {
		branch.LastCommit = parts[1]
	}

	return branch
}

// parseAheadBehind parses "ahead N, behind M" or "gone" from tracking status.
func (p *BranchParser) parseAheadBehind(status string, branch *Branch) {
	matches := aheadBehindRegex.FindAllStringSubmatch(status, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		count, err := strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		switch match[1] {
		case "ahead":
			branch.Ahead = count
		case "behind":
			branch.Behind = count
		}
	}
}

// isCommitHash checks if a string looks like a git commit hash.
func isCommitHash(s string) bool {
	if len(s) < 7 {
		return false
	}
	for _, c := range s {
		isDigit := c >= '0' && c <= '9'
		isLowerHex := c >= 'a' && c <= 'f'
		isUpperHex := c >= 'A' && c <= 'F'
		if !isDigit && !isLowerHex && !isUpperHex {
			return false
		}
	}
	return true
}

// Schema returns the JSON Schema for git branch output.
func (p *BranchParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *BranchParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "branch"
}
