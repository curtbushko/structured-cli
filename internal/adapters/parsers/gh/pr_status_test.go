package gh_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/gh"
)

func TestPRStatusParser_ParseJSON(t *testing.T) {
	input := `{
  "currentBranch": {
    "number": 123,
    "title": "Add new feature",
    "headRefName": "feature-branch",
    "url": "https://github.com/owner/repo/pull/123",
    "state": "OPEN",
    "reviewDecision": "APPROVED",
    "statusCheckRollup": [{"status": "SUCCESS"}]
  },
  "createdBy": [
    {"number": 124, "title": "Another PR", "headRefName": "another-branch", "url": "https://github.com/owner/repo/pull/124"},
    {"number": 125, "title": "Third PR", "headRefName": "third-branch", "url": "https://github.com/owner/repo/pull/125"}
  ],
  "needsReview": [
    {"number": 200, "title": "Review needed", "headRefName": "review-branch", "url": "https://github.com/owner/repo/pull/200"}
  ]
}`

	parser := gh.NewPRStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	prStatus, ok := result.Data.(*gh.PRStatusResult)
	require.True(t, ok, "result.Data should be *gh.PRStatusResult")

	// Current branch PR
	require.NotNil(t, prStatus.CurrentBranch)
	assert.Equal(t, 123, prStatus.CurrentBranch.Number)
	assert.Equal(t, "Add new feature", prStatus.CurrentBranch.Title)
	assert.Equal(t, "feature-branch", prStatus.CurrentBranch.HeadBranch)
	assert.Equal(t, "OPEN", prStatus.CurrentBranch.State)

	// Created by you
	require.Len(t, prStatus.CreatedByYou, 2)
	assert.Equal(t, 124, prStatus.CreatedByYou[0].Number)
	assert.Equal(t, 125, prStatus.CreatedByYou[1].Number)

	// Requesting review
	require.Len(t, prStatus.RequestingReview, 1)
	assert.Equal(t, 200, prStatus.RequestingReview[0].Number)
	assert.Equal(t, "Review needed", prStatus.RequestingReview[0].Title)
}

func TestPRStatusParser_ParseNoCurrentBranch(t *testing.T) {
	input := `{
  "currentBranch": null,
  "createdBy": [],
  "needsReview": []
}`

	parser := gh.NewPRStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	prStatus, ok := result.Data.(*gh.PRStatusResult)
	require.True(t, ok, "result.Data should be *gh.PRStatusResult")

	assert.Nil(t, prStatus.CurrentBranch)
	assert.Empty(t, prStatus.CreatedByYou)
	assert.Empty(t, prStatus.RequestingReview)
}

func TestPRStatusParser_ParseEmptyLists(t *testing.T) {
	input := `{
  "currentBranch": {
    "number": 1,
    "title": "My PR",
    "headRefName": "my-branch",
    "url": "https://github.com/owner/repo/pull/1",
    "state": "OPEN"
  },
  "createdBy": [],
  "needsReview": []
}`

	parser := gh.NewPRStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	prStatus, ok := result.Data.(*gh.PRStatusResult)
	require.True(t, ok, "result.Data should be *gh.PRStatusResult")

	require.NotNil(t, prStatus.CurrentBranch)
	assert.Equal(t, 1, prStatus.CurrentBranch.Number)
	assert.Empty(t, prStatus.CreatedByYou)
	assert.Empty(t, prStatus.RequestingReview)
}

func TestPRStatusParser_Schema(t *testing.T) {
	parser := gh.NewPRStatusParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "current_branch")
	assert.Contains(t, schema.Properties, "created_by_you")
	assert.Contains(t, schema.Properties, "requesting_review")
}

func TestPRStatusParser_Matches(t *testing.T) {
	parser := gh.NewPRStatusParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"gh pr status", "gh", []string{"pr", "status"}, true},
		{"gh pr status with flags", "gh", []string{"pr", "status", "--json"}, true},
		{"gh pr list", "gh", []string{"pr", "list"}, false},
		{"gh pr view", "gh", []string{"pr", "view"}, false},
		{"gh issue status", "gh", []string{"issue", "status"}, false},
		{"gh only", "gh", []string{}, false},
		{"gh pr only", "gh", []string{"pr"}, false},
		{"git pr status", "git", []string{"pr", "status"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPRStatusParser_InvalidJSON(t *testing.T) {
	parser := gh.NewPRStatusParser()
	result, err := parser.Parse(strings.NewReader(invalidJSON))

	require.NoError(t, err)
	require.NotNil(t, result.Error)
}
