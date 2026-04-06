package gh

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ghRunViewItem represents the JSON output format from gh run view --json.
type ghRunViewItem struct {
	DatabaseID   int64    `json:"databaseId"`
	DisplayTitle string   `json:"displayTitle"`
	Status       string   `json:"status"`
	Conclusion   *string  `json:"conclusion"`
	WorkflowName string   `json:"workflowName"`
	HeadBranch   string   `json:"headBranch"`
	HeadSha      string   `json:"headSha"`
	Event        string   `json:"event"`
	Actor        ghAuthor `json:"actor"`
	Jobs         []ghJob  `json:"jobs"`
	URL          string   `json:"url"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
	RunStartedAt string   `json:"runStartedAt"`
}

// ghJob represents a job in gh JSON output.
type ghJob struct {
	DatabaseID  int64    `json:"databaseId"`
	Name        string   `json:"name"`
	Status      string   `json:"status"`
	Conclusion  *string  `json:"conclusion"`
	StartedAt   *string  `json:"startedAt"`
	CompletedAt *string  `json:"completedAt"`
	Steps       []ghStep `json:"steps"`
}

// ghStep represents a step in gh JSON output.
type ghStep struct {
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	Conclusion *string `json:"conclusion"`
	Number     int     `json:"number"`
}

// RunViewParser parses the output of 'gh run view'.
type RunViewParser struct {
	schema domain.Schema
}

// NewRunViewParser creates a new RunViewParser with the gh-run-view schema.
func NewRunViewParser() *RunViewParser {
	return &RunViewParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/gh-run-view.json",
			"GitHub Run View Output",
			"object",
			map[string]domain.PropertySchema{
				"database_id":    {Type: "integer", Description: "Run ID"},
				"display_title":  {Type: "string", Description: "Run display title"},
				"status":         {Type: "string", Description: "Run status"},
				"conclusion":     {Type: "string", Description: "Run conclusion"},
				"workflow_name":  {Type: "string", Description: "Workflow name"},
				"head_branch":    {Type: "string", Description: "Branch that triggered the run"},
				"head_sha":       {Type: "string", Description: "Commit SHA"},
				"event":          {Type: "string", Description: "Trigger event"},
				"actor":          {Type: "object", Description: "User who triggered the run"},
				"jobs":           {Type: "array", Description: "Jobs in the workflow run"},
				"url":            {Type: "string", Description: "Run URL"},
				"created_at":     {Type: "string", Description: "Creation timestamp"},
				"updated_at":     {Type: "string", Description: "Last update timestamp"},
				"run_started_at": {Type: "string", Description: "Run start timestamp"},
			},
			[]string{"database_id", "status", "jobs"},
		),
	}
}

// Parse reads gh run view JSON output and returns structured data.
func (p *RunViewParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			"",
			0,
		), nil
	}

	raw := string(data)

	var item ghRunViewItem
	if err := json.Unmarshal(data, &item); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("parse JSON: %w", err),
			raw,
			0,
		), nil
	}

	// Convert to our domain types
	result := &RunViewResult{
		DatabaseID:   item.DatabaseID,
		DisplayTitle: item.DisplayTitle,
		Status:       item.Status,
		Conclusion:   item.Conclusion,
		WorkflowName: item.WorkflowName,
		HeadBranch:   item.HeadBranch,
		HeadSha:      item.HeadSha,
		Event:        item.Event,
		Actor:        Author{Login: item.Actor.Login, Name: item.Actor.Name},
		Jobs:         convertJobs(item.Jobs),
		URL:          item.URL,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
		RunStartedAt: item.RunStartedAt,
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for gh run view output.
func (p *RunViewParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RunViewParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "gh" {
		return false
	}
	if len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == "run" && subcommands[1] == "view"
}

// convertJobs converts gh jobs to our domain type.
func convertJobs(ghJobs []ghJob) []Job {
	jobs := make([]Job, 0, len(ghJobs))
	for _, j := range ghJobs {
		job := Job{
			DatabaseID:  j.DatabaseID,
			Name:        j.Name,
			Status:      j.Status,
			Conclusion:  j.Conclusion,
			StartedAt:   j.StartedAt,
			CompletedAt: j.CompletedAt,
			Steps:       convertSteps(j.Steps),
		}
		jobs = append(jobs, job)
	}
	return jobs
}

// convertSteps converts gh steps to our domain type.
func convertSteps(ghSteps []ghStep) []Step {
	steps := make([]Step, 0, len(ghSteps))
	for _, s := range ghSteps {
		steps = append(steps, Step(s))
	}
	return steps
}
