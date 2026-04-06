package gh

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ghRunListItem represents the JSON output format from gh run list --json.
type ghRunListItem struct {
	DatabaseID   int64    `json:"databaseId"`
	DisplayTitle string   `json:"displayTitle"`
	Status       string   `json:"status"`
	Conclusion   *string  `json:"conclusion"`
	WorkflowName string   `json:"workflowName"`
	HeadBranch   string   `json:"headBranch"`
	HeadSha      string   `json:"headSha"`
	Event        string   `json:"event"`
	Actor        ghAuthor `json:"actor"`
	URL          string   `json:"url"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}

// RunListParser parses the output of 'gh run list'.
type RunListParser struct {
	schema domain.Schema
}

// NewRunListParser creates a new RunListParser with the gh-run-list schema.
func NewRunListParser() *RunListParser {
	return &RunListParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/gh-run-list.json",
			"GitHub Run List Output",
			"object",
			map[string]domain.PropertySchema{
				"workflow_runs": {Type: "array", Description: "List of workflow runs"},
			},
			[]string{"workflow_runs"},
		),
	}
}

// Parse reads gh run list JSON output and returns structured data.
func (p *RunListParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			"",
			0,
		), nil
	}

	raw := string(data)

	var items []ghRunListItem
	if err := json.Unmarshal(data, &items); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("parse JSON: %w", err),
			raw,
			0,
		), nil
	}

	// Convert to our domain types
	result := &RunListResult{
		WorkflowRuns: make([]WorkflowRun, 0, len(items)),
	}

	for _, item := range items {
		run := WorkflowRun{
			DatabaseID:   item.DatabaseID,
			DisplayTitle: item.DisplayTitle,
			Status:       item.Status,
			Conclusion:   item.Conclusion,
			WorkflowName: item.WorkflowName,
			HeadBranch:   item.HeadBranch,
			HeadSha:      item.HeadSha,
			Event:        item.Event,
			Actor:        Author{Login: item.Actor.Login, Name: item.Actor.Name},
			URL:          item.URL,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
		}
		result.WorkflowRuns = append(result.WorkflowRuns, run)
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for gh run list output.
func (p *RunListParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RunListParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "gh" {
		return false
	}
	if len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == "run" && subcommands[1] == "list"
}
