package gh

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ghPRViewItem represents the JSON output format from gh pr view --json.
type ghPRViewItem struct {
	Number            int               `json:"number"`
	Title             string            `json:"title"`
	Body              string            `json:"body"`
	State             string            `json:"state"`
	Author            ghAuthor          `json:"author"`
	Labels            []ghLabel         `json:"labels"`
	Assignees         []ghAuthor        `json:"assignees"`
	ReviewRequests    []ghAuthor        `json:"reviewRequests"`
	Reviews           []ghReview        `json:"reviews"`
	StatusCheckRollup []ghStatusCheck   `json:"statusCheckRollup"`
	Comments          ghCommentsWrapper `json:"comments"`
	Additions         int               `json:"additions"`
	Deletions         int               `json:"deletions"`
	ChangedFiles      int               `json:"changedFiles"`
	Mergeable         string            `json:"mergeable"`
	CreatedAt         string            `json:"createdAt"`
	UpdatedAt         string            `json:"updatedAt"`
	URL               string            `json:"url"`
}

// ghReview represents a review in gh JSON output.
type ghReview struct {
	Author      ghAuthor `json:"author"`
	State       string   `json:"state"`
	Body        string   `json:"body,omitempty"`
	SubmittedAt string   `json:"submittedAt,omitempty"`
}

// ghStatusCheck represents a status check in gh JSON output.
type ghStatusCheck struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion,omitempty"`
	DetailsURL string `json:"detailsUrl,omitempty"`
}

// ghCommentsWrapper wraps the comments count from gh JSON output.
type ghCommentsWrapper struct {
	TotalCount int `json:"totalCount"`
}

// PRViewParser parses the output of 'gh pr view'.
type PRViewParser struct {
	schema domain.Schema
}

// NewPRViewParser creates a new PRViewParser with the gh-pr-view schema.
func NewPRViewParser() *PRViewParser {
	return &PRViewParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/gh-pr-view.json",
			"GitHub PR View Output",
			"object",
			map[string]domain.PropertySchema{
				"number":        {Type: "integer", Description: "PR number"},
				"title":         {Type: "string", Description: "PR title"},
				"body":          {Type: "string", Description: "PR description body"},
				"state":         {Type: "string", Description: "PR state (OPEN, CLOSED, MERGED)"},
				"author":        {Type: "object", Description: "PR author"},
				"labels":        {Type: "array", Description: "PR labels"},
				"assignees":     {Type: "array", Description: "PR assignees"},
				"reviewers":     {Type: "array", Description: "PR reviewers"},
				"reviews":       {Type: "array", Description: "PR reviews"},
				"checks":        {Type: "array", Description: "CI/CD checks"},
				"comments":      {Type: "integer", Description: "Number of comments"},
				"additions":     {Type: "integer", Description: "Lines added"},
				"deletions":     {Type: "integer", Description: "Lines deleted"},
				"changed_files": {Type: "integer", Description: "Number of changed files"},
				"mergeable":     {Type: "string", Description: "Mergeable state"},
				"created_at":    {Type: "string", Description: "Creation timestamp"},
				"updated_at":    {Type: "string", Description: "Last update timestamp"},
				"url":           {Type: "string", Description: "PR URL"},
			},
			[]string{"number", "title", "state", "author"},
		),
	}
}

// Parse reads gh pr view JSON output and returns structured data.
func (p *PRViewParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			"",
			0,
		), nil
	}

	raw := string(data)

	var item ghPRViewItem
	if err := json.Unmarshal(data, &item); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("parse JSON: %w", err),
			raw,
			0,
		), nil
	}

	// Convert to our domain types
	result := &PRViewResult{
		Number:       item.Number,
		Title:        item.Title,
		Body:         item.Body,
		State:        item.State,
		Author:       Author{Login: item.Author.Login, Name: item.Author.Name},
		Labels:       convertLabels(item.Labels),
		Assignees:    convertAuthors(item.Assignees),
		Reviewers:    convertAuthors(item.ReviewRequests),
		Reviews:      convertReviews(item.Reviews),
		Checks:       convertChecks(item.StatusCheckRollup),
		Comments:     item.Comments.TotalCount,
		Additions:    item.Additions,
		Deletions:    item.Deletions,
		ChangedFiles: item.ChangedFiles,
		Mergeable:    item.Mergeable,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
		URL:          item.URL,
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for gh pr view output.
func (p *PRViewParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PRViewParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "gh" {
		return false
	}
	if len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == "pr" && subcommands[1] == "view"
}

// convertAuthors converts gh authors to our domain type.
func convertAuthors(ghAuthors []ghAuthor) []Author {
	authors := make([]Author, 0, len(ghAuthors))
	for _, a := range ghAuthors {
		authors = append(authors, Author(a))
	}
	return authors
}

// convertReviews converts gh reviews to our domain type.
func convertReviews(ghReviews []ghReview) []Review {
	reviews := make([]Review, 0, len(ghReviews))
	for _, r := range ghReviews {
		reviews = append(reviews, Review{
			Author:      Author{Login: r.Author.Login, Name: r.Author.Name},
			State:       r.State,
			Body:        r.Body,
			SubmittedAt: r.SubmittedAt,
		})
	}
	return reviews
}

// convertChecks converts gh status checks to our domain type.
func convertChecks(ghChecks []ghStatusCheck) []Check {
	checks := make([]Check, 0, len(ghChecks))
	for _, c := range ghChecks {
		checks = append(checks, Check{
			Name:       c.Name,
			Status:     c.Status,
			Conclusion: c.Conclusion,
			URL:        c.DetailsURL,
		})
	}
	return checks
}
