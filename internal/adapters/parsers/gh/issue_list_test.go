package gh_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/gh"
)

func TestIssueListParser_ParseJSON(t *testing.T) {
	input := `[
  {
    "number": 42,
    "title": "Bug: Application crashes on startup",
    "state": "OPEN",
    "author": {"login": "reporter", "name": "Bug Reporter"},
    "labels": [{"name": "bug", "color": "d73a4a"}, {"name": "priority:high", "color": "ff0000"}],
    "assignees": [{"login": "developer", "name": "Dev Developer"}],
    "createdAt": "2024-01-10T08:00:00Z",
    "updatedAt": "2024-01-15T12:30:00Z",
    "url": "https://github.com/owner/repo/issues/42",
    "comments": 5
  },
  {
    "number": 43,
    "title": "Feature request: Dark mode",
    "state": "OPEN",
    "author": {"login": "user123"},
    "labels": [{"name": "enhancement", "color": "a2eeef"}],
    "assignees": [],
    "createdAt": "2024-01-12T14:00:00Z",
    "updatedAt": "2024-01-12T14:00:00Z",
    "url": "https://github.com/owner/repo/issues/43",
    "comments": 0
  }
]`

	parser := gh.NewIssueListParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	issueList, ok := result.Data.(*gh.IssueListResult)
	require.True(t, ok, "result.Data should be *gh.IssueListResult")

	require.Len(t, issueList.Issues, 2)

	// Verify first issue
	issue1 := issueList.Issues[0]
	assert.Equal(t, 42, issue1.Number)
	assert.Equal(t, "Bug: Application crashes on startup", issue1.Title)
	assert.Equal(t, "OPEN", issue1.State)
	assert.Equal(t, "reporter", issue1.Author.Login)
	assert.Equal(t, "Bug Reporter", issue1.Author.Name)
	assert.Len(t, issue1.Labels, 2)
	assert.Equal(t, "bug", issue1.Labels[0].Name)
	assert.Len(t, issue1.Assignees, 1)
	assert.Equal(t, "developer", issue1.Assignees[0].Login)
	assert.Equal(t, 5, issue1.Comments)
	assert.Equal(t, "https://github.com/owner/repo/issues/42", issue1.URL)

	// Verify second issue with no assignees
	issue2 := issueList.Issues[1]
	assert.Equal(t, 43, issue2.Number)
	assert.Empty(t, issue2.Assignees)
	assert.Equal(t, 0, issue2.Comments)
}

func TestIssueListParser_ParseEmptyList(t *testing.T) {
	input := `[]`

	parser := gh.NewIssueListParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	issueList, ok := result.Data.(*gh.IssueListResult)
	require.True(t, ok, "result.Data should be *gh.IssueListResult")
	assert.Empty(t, issueList.Issues)
}

func TestIssueListParser_ParseClosedIssues(t *testing.T) {
	input := `[
  {
    "number": 10,
    "title": "Fixed bug",
    "state": "CLOSED",
    "author": {"login": "fixer"},
    "labels": [],
    "assignees": [],
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-05T00:00:00Z",
    "url": "https://github.com/owner/repo/issues/10",
    "comments": 2
  }
]`

	parser := gh.NewIssueListParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	issueList, ok := result.Data.(*gh.IssueListResult)
	require.True(t, ok)

	require.Len(t, issueList.Issues, 1)
	assert.Equal(t, "CLOSED", issueList.Issues[0].State)
}

func TestIssueListParser_Schema(t *testing.T) {
	parser := gh.NewIssueListParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "issues")
}

func TestIssueListParser_Matches(t *testing.T) {
	parser := gh.NewIssueListParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"gh issue list", "gh", []string{"issue", "list"}, true},
		{"gh issue list with flags", "gh", []string{"issue", "list", "--state=all"}, true},
		{"gh issue view", "gh", []string{"issue", "view"}, false},
		{"gh pr list", "gh", []string{"pr", "list"}, false},
		{"gh only", "gh", []string{}, false},
		{"gh issue only", "gh", []string{"issue"}, false},
		{"git issue list", "git", []string{"issue", "list"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIssueListParser_InvalidJSON(t *testing.T) {
	parser := gh.NewIssueListParser()
	result, err := parser.Parse(strings.NewReader(invalidJSON))

	require.NoError(t, err)
	require.NotNil(t, result.Error)
}
