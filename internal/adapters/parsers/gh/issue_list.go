package gh

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ghIssueListItem represents the JSON output format from gh issue list --json.
type ghIssueListItem struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"`
	Author    ghAuthor   `json:"author"`
	Labels    []ghLabel  `json:"labels"`
	Assignees []ghAuthor `json:"assignees"`
	CreatedAt string     `json:"createdAt"`
	UpdatedAt string     `json:"updatedAt"`
	URL       string     `json:"url"`
	Comments  int        `json:"comments"`
}

// IssueListParser parses the output of 'gh issue list'.
type IssueListParser struct {
	schema domain.Schema
}

// NewIssueListParser creates a new IssueListParser with the gh-issue-list schema.
func NewIssueListParser() *IssueListParser {
	return &IssueListParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/gh-issue-list.json",
			"GitHub Issue List Output",
			"object",
			map[string]domain.PropertySchema{
				"issues": {Type: "array", Description: "List of issues"},
			},
			[]string{"issues"},
		),
	}
}

// Parse reads gh issue list JSON output and returns structured data.
func (p *IssueListParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			"",
			0,
		), nil
	}

	raw := string(data)

	var items []ghIssueListItem
	if err := json.Unmarshal(data, &items); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("parse JSON: %w", err),
			raw,
			0,
		), nil
	}

	// Convert to our domain types
	result := &IssueListResult{
		Issues: make([]Issue, 0, len(items)),
	}

	for _, item := range items {
		issue := Issue{
			Number:    item.Number,
			Title:     item.Title,
			State:     item.State,
			Author:    Author{Login: item.Author.Login, Name: item.Author.Name},
			Labels:    convertLabels(item.Labels),
			Assignees: convertAssignees(item.Assignees),
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
			URL:       item.URL,
			Comments:  item.Comments,
		}
		result.Issues = append(result.Issues, issue)
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for gh issue list output.
func (p *IssueListParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *IssueListParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "gh" {
		return false
	}
	if len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == "issue" && subcommands[1] == "list"
}

// convertAssignees converts gh authors to our domain type for assignees.
func convertAssignees(ghAssignees []ghAuthor) []Author {
	assignees := make([]Author, 0, len(ghAssignees))
	for _, a := range ghAssignees {
		assignees = append(assignees, Author(a))
	}
	return assignees
}
