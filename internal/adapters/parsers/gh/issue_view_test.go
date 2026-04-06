package gh_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/gh"
)

func TestIssueViewParser_ParseJSON(t *testing.T) {
	input := `{
  "number": 42,
  "title": "Bug: Application crashes on startup",
  "body": "## Description\n\nThe application crashes immediately when launched.\n\n## Steps to reproduce\n\n1. Launch app\n2. Observe crash",
  "state": "OPEN",
  "author": {"login": "reporter", "name": "Bug Reporter"},
  "labels": [{"name": "bug", "color": "d73a4a"}, {"name": "priority:high", "color": "ff0000"}],
  "assignees": [{"login": "developer", "name": "Dev Developer"}],
  "milestone": {"title": "v1.0", "number": 1, "state": "open"},
  "projectItems": [{"title": "Q1 Sprint", "number": 5}],
  "reactionGroups": [
    {"content": "THUMBS_UP", "users": {"totalCount": 10}},
    {"content": "THUMBS_DOWN", "users": {"totalCount": 2}},
    {"content": "LAUGH", "users": {"totalCount": 1}},
    {"content": "HOORAY", "users": {"totalCount": 3}},
    {"content": "CONFUSED", "users": {"totalCount": 0}},
    {"content": "HEART", "users": {"totalCount": 5}},
    {"content": "ROCKET", "users": {"totalCount": 2}},
    {"content": "EYES", "users": {"totalCount": 1}}
  ],
  "comments": 5,
  "createdAt": "2024-01-10T08:00:00Z",
  "updatedAt": "2024-01-15T12:30:00Z",
  "closedAt": null,
  "url": "https://github.com/owner/repo/issues/42"
}`

	parser := gh.NewIssueViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	issue, ok := result.Data.(*gh.IssueViewResult)
	require.True(t, ok, "result.Data should be *gh.IssueViewResult")

	assert.Equal(t, 42, issue.Number)
	assert.Equal(t, "Bug: Application crashes on startup", issue.Title)
	assert.Contains(t, issue.Body, "## Description")
	assert.Equal(t, "OPEN", issue.State)
	assert.Equal(t, "reporter", issue.Author.Login)
	assert.Equal(t, "Bug Reporter", issue.Author.Name)

	// Labels
	assert.Len(t, issue.Labels, 2)
	assert.Equal(t, "bug", issue.Labels[0].Name)
	assert.Equal(t, "d73a4a", issue.Labels[0].Color)

	// Assignees
	assert.Len(t, issue.Assignees, 1)
	assert.Equal(t, "developer", issue.Assignees[0].Login)

	// Milestone
	require.NotNil(t, issue.Milestone)
	assert.Equal(t, "v1.0", issue.Milestone.Title)
	assert.Equal(t, 1, issue.Milestone.Number)
	assert.Equal(t, "open", issue.Milestone.State)

	// Project
	require.NotNil(t, issue.Project)
	assert.Equal(t, "Q1 Sprint", issue.Project.Title)

	// Reactions
	assert.Equal(t, 10, issue.Reactions.ThumbsUp)
	assert.Equal(t, 2, issue.Reactions.ThumbsDown)
	assert.Equal(t, 1, issue.Reactions.Laugh)
	assert.Equal(t, 3, issue.Reactions.Hooray)
	assert.Equal(t, 0, issue.Reactions.Confused)
	assert.Equal(t, 5, issue.Reactions.Heart)
	assert.Equal(t, 2, issue.Reactions.Rocket)
	assert.Equal(t, 1, issue.Reactions.Eyes)

	// Timestamps
	assert.Equal(t, 5, issue.Comments)
	assert.Equal(t, "2024-01-10T08:00:00Z", issue.CreatedAt)
	assert.Equal(t, "2024-01-15T12:30:00Z", issue.UpdatedAt)
	assert.Nil(t, issue.ClosedAt)
	assert.Equal(t, "https://github.com/owner/repo/issues/42", issue.URL)
}

func TestIssueViewParser_ParseClosedIssue(t *testing.T) {
	input := `{
  "number": 10,
  "title": "Fixed bug",
  "body": "This bug has been fixed.",
  "state": "CLOSED",
  "author": {"login": "fixer"},
  "labels": [],
  "assignees": [],
  "milestone": null,
  "projectItems": [],
  "reactionGroups": [],
  "comments": 2,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-05T00:00:00Z",
  "closedAt": "2024-01-05T00:00:00Z",
  "url": "https://github.com/owner/repo/issues/10"
}`

	parser := gh.NewIssueViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	issue, ok := result.Data.(*gh.IssueViewResult)
	require.True(t, ok)

	assert.Equal(t, "CLOSED", issue.State)
	require.NotNil(t, issue.ClosedAt)
	assert.Equal(t, "2024-01-05T00:00:00Z", *issue.ClosedAt)
	assert.Nil(t, issue.Milestone)
	assert.Nil(t, issue.Project)
}

func TestIssueViewParser_ParseIssueWithNoOptionalFields(t *testing.T) {
	input := `{
  "number": 1,
  "title": "Simple issue",
  "body": "",
  "state": "OPEN",
  "author": {"login": "user"},
  "labels": [],
  "assignees": [],
  "milestone": null,
  "projectItems": [],
  "reactionGroups": [],
  "comments": 0,
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "closedAt": null,
  "url": "https://github.com/owner/repo/issues/1"
}`

	parser := gh.NewIssueViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	issue, ok := result.Data.(*gh.IssueViewResult)
	require.True(t, ok)

	assert.Equal(t, 1, issue.Number)
	assert.Empty(t, issue.Body)
	assert.Empty(t, issue.Labels)
	assert.Empty(t, issue.Assignees)
	assert.Nil(t, issue.Milestone)
	assert.Nil(t, issue.Project)
	assert.Equal(t, 0, issue.Reactions.ThumbsUp)
}

func TestIssueViewParser_Schema(t *testing.T) {
	parser := gh.NewIssueViewParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "number")
	assert.Contains(t, schema.Properties, "title")
	assert.Contains(t, schema.Properties, "body")
	assert.Contains(t, schema.Properties, "state")
}

func TestIssueViewParser_Matches(t *testing.T) {
	parser := gh.NewIssueViewParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"gh issue view", "gh", []string{"issue", "view"}, true},
		{"gh issue view with number", "gh", []string{"issue", "view", "42"}, true},
		{"gh issue list", "gh", []string{"issue", "list"}, false},
		{"gh pr view", "gh", []string{"pr", "view"}, false},
		{"gh only", "gh", []string{}, false},
		{"gh issue only", "gh", []string{"issue"}, false},
		{"git issue view", "git", []string{"issue", "view"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIssueViewParser_InvalidJSON(t *testing.T) {
	parser := gh.NewIssueViewParser()
	result, err := parser.Parse(strings.NewReader(invalidJSON))

	require.NoError(t, err)
	require.NotNil(t, result.Error)
}
