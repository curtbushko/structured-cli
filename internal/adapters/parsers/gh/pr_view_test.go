package gh_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/gh"
)

func TestPRViewParser_ParseJSON(t *testing.T) {
	input := `{
  "number": 123,
  "title": "Add new feature",
  "body": "This PR adds a great new feature.\n\nDetails here.",
  "state": "OPEN",
  "author": {"login": "octocat", "name": "The Octocat"},
  "labels": [{"name": "enhancement", "color": "84b6eb"}],
  "assignees": [{"login": "reviewer1"}],
  "reviewRequests": [{"login": "reviewer2"}],
  "reviews": [
    {"author": {"login": "reviewer1"}, "state": "APPROVED", "body": "LGTM!", "submittedAt": "2024-01-16T10:00:00Z"}
  ],
  "statusCheckRollup": [
    {"name": "CI", "status": "SUCCESS", "conclusion": "SUCCESS", "detailsUrl": "https://ci.example.com/123"}
  ],
  "comments": {"totalCount": 5},
  "additions": 100,
  "deletions": 50,
  "changedFiles": 10,
  "mergeable": "MERGEABLE",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-16T14:45:00Z",
  "url": "https://github.com/owner/repo/pull/123"
}`

	parser := gh.NewPRViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	prView, ok := result.Data.(*gh.PRViewResult)
	require.True(t, ok, "result.Data should be *gh.PRViewResult")

	assert.Equal(t, 123, prView.Number)
	assert.Equal(t, "Add new feature", prView.Title)
	assert.Contains(t, prView.Body, "great new feature")
	assert.Equal(t, "OPEN", prView.State)
	assert.Equal(t, "octocat", prView.Author.Login)

	// Labels
	require.Len(t, prView.Labels, 1)
	assert.Equal(t, "enhancement", prView.Labels[0].Name)

	// Assignees
	require.Len(t, prView.Assignees, 1)
	assert.Equal(t, "reviewer1", prView.Assignees[0].Login)

	// Reviewers
	require.Len(t, prView.Reviewers, 1)
	assert.Equal(t, "reviewer2", prView.Reviewers[0].Login)

	// Reviews
	require.Len(t, prView.Reviews, 1)
	assert.Equal(t, "reviewer1", prView.Reviews[0].Author.Login)
	assert.Equal(t, "APPROVED", prView.Reviews[0].State)

	// Checks
	require.Len(t, prView.Checks, 1)
	assert.Equal(t, "CI", prView.Checks[0].Name)
	assert.Equal(t, "SUCCESS", prView.Checks[0].Status)

	// Stats
	assert.Equal(t, 5, prView.Comments)
	assert.Equal(t, 100, prView.Additions)
	assert.Equal(t, 50, prView.Deletions)
	assert.Equal(t, 10, prView.ChangedFiles)
	assert.Equal(t, "MERGEABLE", prView.Mergeable)
}

func TestPRViewParser_ParseMinimalPR(t *testing.T) {
	input := `{
  "number": 1,
  "title": "Minimal PR",
  "body": "",
  "state": "OPEN",
  "author": {"login": "user"},
  "labels": [],
  "assignees": [],
  "reviewRequests": [],
  "reviews": [],
  "statusCheckRollup": [],
  "comments": {"totalCount": 0},
  "additions": 1,
  "deletions": 0,
  "changedFiles": 1,
  "mergeable": "UNKNOWN",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:30:00Z",
  "url": "https://github.com/owner/repo/pull/1"
}`

	parser := gh.NewPRViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	prView, ok := result.Data.(*gh.PRViewResult)
	require.True(t, ok, "result.Data should be *gh.PRViewResult")

	assert.Equal(t, 1, prView.Number)
	assert.Equal(t, "Minimal PR", prView.Title)
	assert.Empty(t, prView.Body)
	assert.Empty(t, prView.Labels)
	assert.Empty(t, prView.Reviews)
	assert.Empty(t, prView.Checks)
}

func TestPRViewParser_Schema(t *testing.T) {
	parser := gh.NewPRViewParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "number")
	assert.Contains(t, schema.Properties, "title")
	assert.Contains(t, schema.Properties, "reviews")
	assert.Contains(t, schema.Properties, "checks")
}

func TestPRViewParser_Matches(t *testing.T) {
	parser := gh.NewPRViewParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"gh pr view", "gh", []string{"pr", "view"}, true},
		{"gh pr view with number", "gh", []string{"pr", "view", "123"}, true},
		{"gh pr view with flags", "gh", []string{"pr", "view", "--json"}, true},
		{"gh pr list", "gh", []string{"pr", "list"}, false},
		{"gh pr status", "gh", []string{"pr", "status"}, false},
		{"gh issue view", "gh", []string{"issue", "view"}, false},
		{"gh only", "gh", []string{}, false},
		{"gh pr only", "gh", []string{"pr"}, false},
		{"git pr view", "git", []string{"pr", "view"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPRViewParser_InvalidJSON(t *testing.T) {
	parser := gh.NewPRViewParser()
	result, err := parser.Parse(strings.NewReader(invalidJSON))

	require.NoError(t, err)
	require.NotNil(t, result.Error)
}
