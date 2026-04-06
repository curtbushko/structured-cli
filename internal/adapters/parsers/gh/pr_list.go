package gh

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ghPRListItem represents the JSON output format from gh pr list --json.
type ghPRListItem struct {
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	State       string    `json:"state"`
	Author      ghAuthor  `json:"author"`
	Labels      []ghLabel `json:"labels"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	URL         string    `json:"url"`
	HeadRefName string    `json:"headRefName"`
	BaseRefName string    `json:"baseRefName"`
	IsDraft     bool      `json:"isDraft"`
}

// ghAuthor represents the author in gh JSON output.
type ghAuthor struct {
	Login string `json:"login"`
	Name  string `json:"name,omitempty"`
}

// ghLabel represents a label in gh JSON output.
type ghLabel struct {
	Name        string `json:"name"`
	Color       string `json:"color,omitempty"`
	Description string `json:"description,omitempty"`
}

// PRListParser parses the output of 'gh pr list'.
type PRListParser struct {
	schema domain.Schema
}

// NewPRListParser creates a new PRListParser with the gh-pr-list schema.
func NewPRListParser() *PRListParser {
	return &PRListParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/gh-pr-list.json",
			"GitHub PR List Output",
			"object",
			map[string]domain.PropertySchema{
				"pull_requests": {Type: "array", Description: "List of pull requests"},
			},
			[]string{"pull_requests"},
		),
	}
}

// Parse reads gh pr list JSON output and returns structured data.
func (p *PRListParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			"",
			0,
		), nil
	}

	raw := string(data)

	var items []ghPRListItem
	if err := json.Unmarshal(data, &items); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("parse JSON: %w", err),
			raw,
			0,
		), nil
	}

	// Convert to our domain types
	result := &PRListResult{
		PullRequests: make([]PullRequest, 0, len(items)),
	}

	for _, item := range items {
		pr := PullRequest{
			Number:     item.Number,
			Title:      item.Title,
			State:      item.State,
			Author:     Author{Login: item.Author.Login, Name: item.Author.Name},
			Labels:     convertLabels(item.Labels),
			CreatedAt:  item.CreatedAt,
			UpdatedAt:  item.UpdatedAt,
			URL:        item.URL,
			HeadBranch: item.HeadRefName,
			BaseBranch: item.BaseRefName,
			Draft:      item.IsDraft,
		}
		result.PullRequests = append(result.PullRequests, pr)
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for gh pr list output.
func (p *PRListParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PRListParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "gh" {
		return false
	}
	if len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == "pr" && subcommands[1] == "list"
}

// convertLabels converts gh labels to our domain type.
func convertLabels(ghLabels []ghLabel) []Label {
	labels := make([]Label, 0, len(ghLabels))
	for _, l := range ghLabels {
		labels = append(labels, Label(l))
	}
	return labels
}
