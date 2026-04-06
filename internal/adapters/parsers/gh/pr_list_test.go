package gh_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/gh"
)

func TestPRListParser_ParseJSON(t *testing.T) {
	input := `[
  {
    "number": 123,
    "title": "Add new feature",
    "state": "OPEN",
    "author": {"login": "octocat", "name": "The Octocat"},
    "labels": [{"name": "enhancement", "color": "84b6eb"}],
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-16T14:45:00Z",
    "url": "https://github.com/owner/repo/pull/123",
    "headRefName": "feature-branch",
    "baseRefName": "main",
    "isDraft": false
  },
  {
    "number": 124,
    "title": "Fix bug",
    "state": "OPEN",
    "author": {"login": "developer"},
    "labels": [{"name": "bug", "color": "d73a4a"}],
    "createdAt": "2024-01-16T09:00:00Z",
    "updatedAt": "2024-01-16T09:00:00Z",
    "url": "https://github.com/owner/repo/pull/124",
    "headRefName": "bugfix-branch",
    "baseRefName": "main",
    "isDraft": true
  }
]`

	parser := gh.NewPRListParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	prList, ok := result.Data.(*gh.PRListResult)
	require.True(t, ok, "result.Data should be *gh.PRListResult")

	require.Len(t, prList.PullRequests, 2)

	// Verify first PR
	pr1 := prList.PullRequests[0]
	assert.Equal(t, 123, pr1.Number)
	assert.Equal(t, "Add new feature", pr1.Title)
	assert.Equal(t, "OPEN", pr1.State)
	assert.Equal(t, "octocat", pr1.Author.Login)
	assert.Equal(t, "The Octocat", pr1.Author.Name)
	assert.Len(t, pr1.Labels, 1)
	assert.Equal(t, "enhancement", pr1.Labels[0].Name)
	assert.Equal(t, "feature-branch", pr1.HeadBranch)
	assert.Equal(t, "main", pr1.BaseBranch)
	assert.False(t, pr1.Draft)

	// Verify second PR is draft
	pr2 := prList.PullRequests[1]
	assert.Equal(t, 124, pr2.Number)
	assert.True(t, pr2.Draft)
}

func TestPRListParser_ParseEmptyList(t *testing.T) {
	input := `[]`

	parser := gh.NewPRListParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	prList, ok := result.Data.(*gh.PRListResult)
	require.True(t, ok, "result.Data should be *gh.PRListResult")
	assert.Empty(t, prList.PullRequests)
}

func TestPRListParser_Schema(t *testing.T) {
	parser := gh.NewPRListParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "pull_requests")
}

func TestPRListParser_Matches(t *testing.T) {
	parser := gh.NewPRListParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"gh pr list", "gh", []string{"pr", "list"}, true},
		{"gh pr list with flags", "gh", []string{"pr", "list", "--state=all"}, true},
		{"gh pr view", "gh", []string{"pr", "view"}, false},
		{"gh pr status", "gh", []string{"pr", "status"}, false},
		{"gh issue list", "gh", []string{"issue", "list"}, false},
		{"gh only", "gh", []string{}, false},
		{"gh pr only", "gh", []string{"pr"}, false},
		{"git pr list", "git", []string{"pr", "list"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPRListParser_InvalidJSON(t *testing.T) {
	parser := gh.NewPRListParser()
	result, err := parser.Parse(strings.NewReader(invalidJSON))

	require.NoError(t, err)
	require.NotNil(t, result.Error)
}
