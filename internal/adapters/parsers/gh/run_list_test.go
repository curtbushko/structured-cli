package gh_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/gh"
)

func TestRunListParser_ParseJSON(t *testing.T) {
	input := `[
  {
    "databaseId": 12345678901,
    "displayTitle": "Build and Test",
    "status": "completed",
    "conclusion": "success",
    "workflowName": "CI",
    "headBranch": "main",
    "headSha": "abc123def456789",
    "event": "push",
    "actor": {"login": "octocat", "name": "The Octocat"},
    "url": "https://github.com/owner/repo/actions/runs/12345678901",
    "createdAt": "2024-01-15T10:30:00Z",
    "updatedAt": "2024-01-15T10:35:00Z"
  },
  {
    "databaseId": 12345678902,
    "displayTitle": "Deploy to Production",
    "status": "in_progress",
    "conclusion": null,
    "workflowName": "CD",
    "headBranch": "release/v1.0",
    "headSha": "def456abc123789",
    "event": "workflow_dispatch",
    "actor": {"login": "developer"},
    "url": "https://github.com/owner/repo/actions/runs/12345678902",
    "createdAt": "2024-01-16T09:00:00Z",
    "updatedAt": "2024-01-16T09:05:00Z"
  }
]`

	parser := gh.NewRunListParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	runList, ok := result.Data.(*gh.RunListResult)
	require.True(t, ok, "result.Data should be *gh.RunListResult")

	require.Len(t, runList.WorkflowRuns, 2)

	// Verify first run (completed)
	run1 := runList.WorkflowRuns[0]
	assert.Equal(t, int64(12345678901), run1.DatabaseID)
	assert.Equal(t, "Build and Test", run1.DisplayTitle)
	assert.Equal(t, "completed", run1.Status)
	require.NotNil(t, run1.Conclusion)
	assert.Equal(t, "success", *run1.Conclusion)
	assert.Equal(t, "CI", run1.WorkflowName)
	assert.Equal(t, "main", run1.HeadBranch)
	assert.Equal(t, "abc123def456789", run1.HeadSha)
	assert.Equal(t, "push", run1.Event)
	assert.Equal(t, "octocat", run1.Actor.Login)
	assert.Equal(t, "The Octocat", run1.Actor.Name)
	assert.Equal(t, "https://github.com/owner/repo/actions/runs/12345678901", run1.URL)

	// Verify second run (in progress with null conclusion)
	run2 := runList.WorkflowRuns[1]
	assert.Equal(t, int64(12345678902), run2.DatabaseID)
	assert.Equal(t, "in_progress", run2.Status)
	assert.Nil(t, run2.Conclusion)
	assert.Equal(t, "workflow_dispatch", run2.Event)
}

func TestRunListParser_ParseEmptyList(t *testing.T) {
	input := `[]`

	parser := gh.NewRunListParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	runList, ok := result.Data.(*gh.RunListResult)
	require.True(t, ok, "result.Data should be *gh.RunListResult")
	assert.Empty(t, runList.WorkflowRuns)
}

func TestRunListParser_ParseInProgressRun(t *testing.T) {
	input := `[
  {
    "databaseId": 12345678903,
    "displayTitle": "Running Tests",
    "status": "queued",
    "conclusion": null,
    "workflowName": "Test Suite",
    "headBranch": "feature/new-tests",
    "headSha": "789abc123def456",
    "event": "pull_request",
    "actor": {"login": "contributor"},
    "url": "https://github.com/owner/repo/actions/runs/12345678903",
    "createdAt": "2024-01-17T08:00:00Z",
    "updatedAt": "2024-01-17T08:00:00Z"
  }
]`

	parser := gh.NewRunListParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	runList, ok := result.Data.(*gh.RunListResult)
	require.True(t, ok, "result.Data should be *gh.RunListResult")

	require.Len(t, runList.WorkflowRuns, 1)
	run := runList.WorkflowRuns[0]

	assert.Equal(t, "queued", run.Status)
	assert.Nil(t, run.Conclusion)
	assert.Equal(t, "pull_request", run.Event)
}

func TestRunListParser_Schema(t *testing.T) {
	parser := gh.NewRunListParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "workflow_runs")
}

func TestRunListParser_Matches(t *testing.T) {
	parser := gh.NewRunListParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"gh run list", "gh", []string{"run", "list"}, true},
		{"gh run list with flags", "gh", []string{"run", "list", "--limit=10"}, true},
		{"gh run view", "gh", []string{"run", "view"}, false},
		{"gh pr list", "gh", []string{"pr", "list"}, false},
		{"gh only", "gh", []string{}, false},
		{"gh run only", "gh", []string{"run"}, false},
		{"git run list", "git", []string{"run", "list"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRunListParser_InvalidJSON(t *testing.T) {
	parser := gh.NewRunListParser()
	result, err := parser.Parse(strings.NewReader(invalidJSON))

	require.NoError(t, err)
	require.NotNil(t, result.Error)
}
