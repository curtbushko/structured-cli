package git

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Push represents the structured output of 'git push'.
type Push struct {
	// Success indicates if the push was successful.
	Success bool `json:"success"`

	// Remote is the remote repository URL or name.
	Remote string `json:"remote"`

	// Branch is the branch that was pushed.
	Branch string `json:"branch"`

	// Commits is the number of commits pushed.
	Commits int `json:"commits"`

	// NewBranch indicates if a new branch was created on the remote.
	NewBranch bool `json:"newBranch"`

	// Errors contains any error messages.
	Errors []string `json:"errors,omitempty"`
}

// PushParser parses the output of 'git push'.
type PushParser struct {
	schema domain.Schema
	// Regex to parse remote URL: To <url>
	remoteRe *regexp.Regexp
	// Regex to parse branch update: hash..hash branch -> branch
	updateRe *regexp.Regexp
	// Regex to parse new branch: * [new branch] branch -> branch
	newBranchRe *regexp.Regexp
	// Regex to parse forced update: + hash...hash branch -> branch (forced update)
	forcedRe *regexp.Regexp
	// Regex to parse rejected: ! [rejected] branch -> branch (reason)
	rejectedRe *regexp.Regexp
}

// NewPushParser creates a new PushParser with the git-push schema.
func NewPushParser() *PushParser {
	return &PushParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-push.json",
			"Git Push Output",
			"object",
			map[string]domain.PropertySchema{
				"success":   {Type: "boolean", Description: "Whether push was successful"},
				"remote":    {Type: "string", Description: "Remote repository"},
				"branch":    {Type: "string", Description: "Branch pushed"},
				"commits":   {Type: "integer", Description: "Number of commits pushed"},
				"newBranch": {Type: "boolean", Description: "Whether a new branch was created"},
				"errors":    {Type: "array", Description: "Error messages"},
			},
			[]string{"success"},
		),
		remoteRe:    regexp.MustCompile(`^To\s+(.+)$`),
		updateRe:    regexp.MustCompile(`^\s*[a-f0-9]+\.\.[a-f0-9]+\s+(\S+)\s+->\s+(\S+)`),
		newBranchRe: regexp.MustCompile(`^\s*\*\s+\[new branch\]\s+(\S+)\s+->\s+(\S+)`),
		forcedRe:    regexp.MustCompile(`^\s*\+\s+[a-f0-9]+\.\.\.[a-f0-9]+\s+(\S+)\s+->\s+(\S+)`),
		rejectedRe:  regexp.MustCompile(`^\s*!\s+\[rejected\]\s+(\S+)\s+->\s+(\S+)`),
	}
}

// Parse reads git push output and returns structured data.
func (p *PushParser) Parse(r io.Reader) (domain.ParseResult, error) {
	push := &Push{
		Success: true, // Assume success unless we see an error
		Errors:  []string{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		p.parseLine(line, push)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(push, rawBuilder.String(), 0), nil
}

// parseLine parses a single line of git push output.
func (p *PushParser) parseLine(line string, push *Push) {
	if line == "" {
		return
	}

	switch {
	case strings.HasPrefix(line, "To "):
		// Remote URL
		if matches := p.remoteRe.FindStringSubmatch(line); matches != nil {
			push.Remote = matches[1]
		}

	case strings.Contains(line, "[new branch]"):
		// New branch created
		if matches := p.newBranchRe.FindStringSubmatch(line); matches != nil {
			push.Branch = matches[2]
			push.NewBranch = true
		}

	case strings.Contains(line, "[rejected]"):
		// Push rejected
		push.Success = false
		if matches := p.rejectedRe.FindStringSubmatch(line); matches != nil {
			push.Branch = matches[2]
		}

	case strings.Contains(line, "(forced update)"):
		// Forced push
		if matches := p.forcedRe.FindStringSubmatch(line); matches != nil {
			push.Branch = matches[2]
		}

	case p.updateRe.MatchString(line):
		// Normal update
		if matches := p.updateRe.FindStringSubmatch(line); matches != nil {
			push.Branch = matches[2]
		}

	case strings.HasPrefix(line, "error:"):
		push.Success = false
		push.Errors = append(push.Errors, line)

	case strings.HasPrefix(line, "fatal:"):
		push.Success = false
		push.Errors = append(push.Errors, line)

	case line == "Everything up-to-date":
		push.Success = true
	}
}

// Schema returns the JSON Schema for git push output.
func (p *PushParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PushParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "push"
}
