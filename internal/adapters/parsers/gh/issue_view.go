package gh

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ghIssueView represents the JSON output format from gh issue view --json.
type ghIssueView struct {
	Number         int               `json:"number"`
	Title          string            `json:"title"`
	Body           string            `json:"body"`
	State          string            `json:"state"`
	Author         ghAuthor          `json:"author"`
	Labels         []ghLabel         `json:"labels"`
	Assignees      []ghAuthor        `json:"assignees"`
	Milestone      *ghMilestone      `json:"milestone"`
	ProjectItems   []ghProjectItem   `json:"projectItems"`
	ReactionGroups []ghReactionGroup `json:"reactionGroups"`
	Comments       int               `json:"comments"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
	ClosedAt       *string           `json:"closedAt"`
	URL            string            `json:"url"`
}

// ghMilestone represents a milestone in gh JSON output.
type ghMilestone struct {
	Title  string `json:"title"`
	Number int    `json:"number"`
	State  string `json:"state"`
}

// ghProjectItem represents a project item in gh JSON output.
type ghProjectItem struct {
	Title  string `json:"title"`
	Number int    `json:"number"`
}

// ghReactionGroup represents a reaction group in gh JSON output.
type ghReactionGroup struct {
	Content string       `json:"content"`
	Users   ghReactUsers `json:"users"`
}

// ghReactUsers represents the users who reacted.
type ghReactUsers struct {
	TotalCount int `json:"totalCount"`
}

// IssueViewParser parses the output of 'gh issue view'.
type IssueViewParser struct {
	schema domain.Schema
}

// NewIssueViewParser creates a new IssueViewParser with the gh-issue-view schema.
func NewIssueViewParser() *IssueViewParser {
	return &IssueViewParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/gh-issue-view.json",
			"GitHub Issue View Output",
			"object",
			map[string]domain.PropertySchema{
				"number": {Type: "integer", Description: "Issue number"},
				"title":  {Type: "string", Description: "Issue title"},
				"body":   {Type: "string", Description: "Issue description body"},
				"state":  {Type: "string", Description: "Issue state (OPEN, CLOSED)"},
			},
			[]string{"number", "title", "state"},
		),
	}
}

// Parse reads gh issue view JSON output and returns structured data.
func (p *IssueViewParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			"",
			0,
		), nil
	}

	raw := string(data)

	var item ghIssueView
	if err := json.Unmarshal(data, &item); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("parse JSON: %w", err),
			raw,
			0,
		), nil
	}

	// Convert to our domain types
	result := &IssueViewResult{
		Number:    item.Number,
		Title:     item.Title,
		Body:      item.Body,
		State:     item.State,
		Author:    Author{Login: item.Author.Login, Name: item.Author.Name},
		Labels:    convertLabels(item.Labels),
		Assignees: convertAssignees(item.Assignees),
		Milestone: convertMilestone(item.Milestone),
		Project:   convertProject(item.ProjectItems),
		Reactions: convertReactions(item.ReactionGroups),
		Comments:  item.Comments,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		ClosedAt:  item.ClosedAt,
		URL:       item.URL,
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for gh issue view output.
func (p *IssueViewParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *IssueViewParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "gh" {
		return false
	}
	if len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == subCmdIssue && subcommands[1] == subCmdView
}

// convertMilestone converts a gh milestone to our domain type.
func convertMilestone(m *ghMilestone) *Milestone {
	if m == nil {
		return nil
	}
	return &Milestone{
		Title:  m.Title,
		Number: m.Number,
		State:  m.State,
	}
}

// convertProject converts gh project items to our domain type.
// Returns the first project item if any exist, nil otherwise.
func convertProject(items []ghProjectItem) *Project {
	if len(items) == 0 {
		return nil
	}
	return &Project{
		Title:  items[0].Title,
		Number: items[0].Number,
	}
}

// convertReactions converts gh reaction groups to our domain type.
func convertReactions(groups []ghReactionGroup) Reactions {
	reactions := Reactions{}
	for _, g := range groups {
		switch g.Content {
		case "THUMBS_UP":
			reactions.ThumbsUp = g.Users.TotalCount
		case "THUMBS_DOWN":
			reactions.ThumbsDown = g.Users.TotalCount
		case "LAUGH":
			reactions.Laugh = g.Users.TotalCount
		case "HOORAY":
			reactions.Hooray = g.Users.TotalCount
		case "CONFUSED":
			reactions.Confused = g.Users.TotalCount
		case "HEART":
			reactions.Heart = g.Users.TotalCount
		case "ROCKET":
			reactions.Rocket = g.Users.TotalCount
		case "EYES":
			reactions.Eyes = g.Users.TotalCount
		}
	}
	return reactions
}
