package git

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Checkout represents the structured output of 'git checkout'.
type Checkout struct {
	// Success indicates if the checkout was successful.
	Success bool `json:"success"`

	// Branch is the branch that was checked out.
	Branch string `json:"branch"`

	// NewBranch indicates if a new branch was created.
	NewBranch bool `json:"newBranch"`

	// Commit is the commit hash if in detached HEAD state.
	Commit string `json:"commit,omitempty"`
}

// CheckoutParser parses the output of 'git checkout'.
type CheckoutParser struct {
	schema domain.Schema
	// Regex to match: Switched to branch 'name'
	switchedRe *regexp.Regexp
	// Regex to match: Switched to a new branch 'name'
	newBranchRe *regexp.Regexp
	// Regex to match: Note: switching to 'commit'
	detachedRe *regexp.Regexp
	// Regex to match: Already on 'branch'
	alreadyOnRe *regexp.Regexp
	// Regex to match: HEAD is now at <hash>
	headNowRe *regexp.Regexp
}

// NewCheckoutParser creates a new CheckoutParser with the git-checkout schema.
func NewCheckoutParser() *CheckoutParser {
	return &CheckoutParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-checkout.json",
			"Git Checkout Output",
			"object",
			map[string]domain.PropertySchema{
				"success":   {Type: "boolean", Description: "Whether checkout was successful"},
				"branch":    {Type: "string", Description: "Branch checked out"},
				"newBranch": {Type: "boolean", Description: "Whether a new branch was created"},
				"commit":    {Type: "string", Description: "Commit hash if detached HEAD"},
			},
			[]string{"success"},
		),
		switchedRe:  regexp.MustCompile(`^Switched to branch '([^']+)'`),
		newBranchRe: regexp.MustCompile(`^Switched to a new branch '([^']+)'`),
		detachedRe:  regexp.MustCompile(`^Note: switching to '([^']+)'`),
		alreadyOnRe: regexp.MustCompile(`^Already on '([^']+)'`),
		headNowRe:   regexp.MustCompile(`^HEAD is now at ([a-f0-9]+)`),
	}
}

// Parse reads git checkout output and returns structured data.
func (p *CheckoutParser) Parse(r io.Reader) (domain.ParseResult, error) {
	checkout := &Checkout{
		Success: true,
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		p.parseLine(line, checkout)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(checkout, rawBuilder.String(), 0), nil
}

// parseLine parses a single line of git checkout output.
func (p *CheckoutParser) parseLine(line string, checkout *Checkout) {
	if line == "" {
		return
	}

	switch {
	case p.newBranchRe.MatchString(line):
		// Must check newBranchRe before switchedRe since both contain "Switched to"
		if matches := p.newBranchRe.FindStringSubmatch(line); matches != nil {
			checkout.Branch = matches[1]
			checkout.NewBranch = true
		}

	case p.switchedRe.MatchString(line):
		if matches := p.switchedRe.FindStringSubmatch(line); matches != nil {
			checkout.Branch = matches[1]
		}

	case p.detachedRe.MatchString(line):
		if matches := p.detachedRe.FindStringSubmatch(line); matches != nil {
			checkout.Commit = matches[1]
		}

	case p.alreadyOnRe.MatchString(line):
		if matches := p.alreadyOnRe.FindStringSubmatch(line); matches != nil {
			checkout.Branch = matches[1]
		}

	case p.headNowRe.MatchString(line):
		if matches := p.headNowRe.FindStringSubmatch(line); matches != nil {
			checkout.Commit = matches[1]
		}

	case strings.HasPrefix(line, "error:"):
		checkout.Success = false

	case strings.HasPrefix(line, "fatal:"):
		checkout.Success = false
	}
}

// Schema returns the JSON Schema for git checkout output.
func (p *CheckoutParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *CheckoutParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "checkout"
}
