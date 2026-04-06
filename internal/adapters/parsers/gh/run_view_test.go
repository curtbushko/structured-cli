package gh_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/gh"
)

func TestRunViewParser_ParseJSON(t *testing.T) {
	input := `{
  "databaseId": 12345678901,
  "displayTitle": "Build and Test",
  "status": "completed",
  "conclusion": "success",
  "workflowName": "CI",
  "headBranch": "main",
  "headSha": "abc123def456789",
  "event": "push",
  "actor": {"login": "octocat", "name": "The Octocat"},
  "jobs": [
    {
      "databaseId": 98765432101,
      "name": "build",
      "status": "completed",
      "conclusion": "success",
      "startedAt": "2024-01-15T10:30:00Z",
      "completedAt": "2024-01-15T10:32:00Z",
      "steps": [
        {
          "name": "Checkout",
          "status": "completed",
          "conclusion": "success",
          "number": 1
        },
        {
          "name": "Build",
          "status": "completed",
          "conclusion": "success",
          "number": 2
        }
      ]
    },
    {
      "databaseId": 98765432102,
      "name": "test",
      "status": "completed",
      "conclusion": "success",
      "startedAt": "2024-01-15T10:32:00Z",
      "completedAt": "2024-01-15T10:35:00Z",
      "steps": [
        {
          "name": "Run tests",
          "status": "completed",
          "conclusion": "success",
          "number": 1
        }
      ]
    }
  ],
  "url": "https://github.com/owner/repo/actions/runs/12345678901",
  "createdAt": "2024-01-15T10:30:00Z",
  "updatedAt": "2024-01-15T10:35:00Z",
  "runStartedAt": "2024-01-15T10:30:00Z"
}`

	parser := gh.NewRunViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	runView, ok := result.Data.(*gh.RunViewResult)
	require.True(t, ok, "result.Data should be *gh.RunViewResult")

	assert.Equal(t, int64(12345678901), runView.DatabaseID)
	assert.Equal(t, "Build and Test", runView.DisplayTitle)
	assert.Equal(t, "completed", runView.Status)
	require.NotNil(t, runView.Conclusion)
	assert.Equal(t, "success", *runView.Conclusion)
	assert.Equal(t, "CI", runView.WorkflowName)
	assert.Equal(t, "main", runView.HeadBranch)
	assert.Equal(t, "abc123def456789", runView.HeadSha)
	assert.Equal(t, "push", runView.Event)
	assert.Equal(t, "octocat", runView.Actor.Login)
	assert.Equal(t, "2024-01-15T10:30:00Z", runView.RunStartedAt)

	// Verify jobs
	require.Len(t, runView.Jobs, 2)

	job1 := runView.Jobs[0]
	assert.Equal(t, int64(98765432101), job1.DatabaseID)
	assert.Equal(t, "build", job1.Name)
	assert.Equal(t, "completed", job1.Status)
	require.NotNil(t, job1.Conclusion)
	assert.Equal(t, "success", *job1.Conclusion)
	require.NotNil(t, job1.StartedAt)
	assert.Equal(t, "2024-01-15T10:30:00Z", *job1.StartedAt)
	require.NotNil(t, job1.CompletedAt)
	assert.Equal(t, "2024-01-15T10:32:00Z", *job1.CompletedAt)

	// Verify steps
	require.Len(t, job1.Steps, 2)
	step1 := job1.Steps[0]
	assert.Equal(t, "Checkout", step1.Name)
	assert.Equal(t, "completed", step1.Status)
	require.NotNil(t, step1.Conclusion)
	assert.Equal(t, "success", *step1.Conclusion)
	assert.Equal(t, 1, step1.Number)
}

func TestRunViewParser_ParseInProgressRun(t *testing.T) {
	input := `{
  "databaseId": 12345678902,
  "displayTitle": "Deploy to Production",
  "status": "in_progress",
  "conclusion": null,
  "workflowName": "CD",
  "headBranch": "main",
  "headSha": "def456abc789",
  "event": "workflow_dispatch",
  "actor": {"login": "developer"},
  "jobs": [
    {
      "databaseId": 98765432201,
      "name": "deploy",
      "status": "in_progress",
      "conclusion": null,
      "startedAt": "2024-01-16T09:00:00Z",
      "completedAt": null,
      "steps": [
        {
          "name": "Deploying",
          "status": "in_progress",
          "conclusion": null,
          "number": 1
        }
      ]
    }
  ],
  "url": "https://github.com/owner/repo/actions/runs/12345678902",
  "createdAt": "2024-01-16T09:00:00Z",
  "updatedAt": "2024-01-16T09:05:00Z",
  "runStartedAt": "2024-01-16T09:00:00Z"
}`

	parser := gh.NewRunViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	runView, ok := result.Data.(*gh.RunViewResult)
	require.True(t, ok, "result.Data should be *gh.RunViewResult")

	assert.Equal(t, "in_progress", runView.Status)
	assert.Nil(t, runView.Conclusion)

	require.Len(t, runView.Jobs, 1)
	job := runView.Jobs[0]
	assert.Equal(t, "in_progress", job.Status)
	assert.Nil(t, job.Conclusion)
	assert.Nil(t, job.CompletedAt)

	require.Len(t, job.Steps, 1)
	step := job.Steps[0]
	assert.Nil(t, step.Conclusion)
}

func TestRunViewParser_ParseCompletedRunWithFailure(t *testing.T) {
	input := `{
  "databaseId": 12345678903,
  "displayTitle": "Failed Build",
  "status": "completed",
  "conclusion": "failure",
  "workflowName": "CI",
  "headBranch": "feature/broken",
  "headSha": "broken123",
  "event": "pull_request",
  "actor": {"login": "contributor"},
  "jobs": [
    {
      "databaseId": 98765432301,
      "name": "build",
      "status": "completed",
      "conclusion": "failure",
      "startedAt": "2024-01-17T08:00:00Z",
      "completedAt": "2024-01-17T08:02:00Z",
      "steps": [
        {
          "name": "Checkout",
          "status": "completed",
          "conclusion": "success",
          "number": 1
        },
        {
          "name": "Build",
          "status": "completed",
          "conclusion": "failure",
          "number": 2
        }
      ]
    }
  ],
  "url": "https://github.com/owner/repo/actions/runs/12345678903",
  "createdAt": "2024-01-17T08:00:00Z",
  "updatedAt": "2024-01-17T08:02:00Z",
  "runStartedAt": "2024-01-17T08:00:00Z"
}`

	parser := gh.NewRunViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	runView, ok := result.Data.(*gh.RunViewResult)
	require.True(t, ok, "result.Data should be *gh.RunViewResult")

	require.NotNil(t, runView.Conclusion)
	assert.Equal(t, "failure", *runView.Conclusion)

	require.Len(t, runView.Jobs, 1)
	job := runView.Jobs[0]
	require.NotNil(t, job.Conclusion)
	assert.Equal(t, "failure", *job.Conclusion)

	// Second step failed
	require.Len(t, job.Steps, 2)
	require.NotNil(t, job.Steps[1].Conclusion)
	assert.Equal(t, "failure", *job.Steps[1].Conclusion)
}

func TestRunViewParser_ParseEmptyJobs(t *testing.T) {
	input := `{
  "databaseId": 12345678904,
  "displayTitle": "Queued Run",
  "status": "queued",
  "conclusion": null,
  "workflowName": "CI",
  "headBranch": "main",
  "headSha": "queued123",
  "event": "push",
  "actor": {"login": "octocat"},
  "jobs": [],
  "url": "https://github.com/owner/repo/actions/runs/12345678904",
  "createdAt": "2024-01-18T10:00:00Z",
  "updatedAt": "2024-01-18T10:00:00Z",
  "runStartedAt": "2024-01-18T10:00:00Z"
}`

	parser := gh.NewRunViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	runView, ok := result.Data.(*gh.RunViewResult)
	require.True(t, ok, "result.Data should be *gh.RunViewResult")

	assert.Empty(t, runView.Jobs)
}

func TestRunViewParser_Schema(t *testing.T) {
	parser := gh.NewRunViewParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "database_id")
	assert.Contains(t, schema.Properties, "jobs")
}

func TestRunViewParser_Matches(t *testing.T) {
	parser := gh.NewRunViewParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"gh run view", "gh", []string{"run", "view"}, true},
		{"gh run view with ID", "gh", []string{"run", "view", "12345"}, true},
		{"gh run view with flags", "gh", []string{"run", "view", "--json", "status"}, true},
		{"gh run list", "gh", []string{"run", "list"}, false},
		{"gh pr view", "gh", []string{"pr", "view"}, false},
		{"gh only", "gh", []string{}, false},
		{"gh run only", "gh", []string{"run"}, false},
		{"git run view", "git", []string{"run", "view"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRunViewParser_InvalidJSON(t *testing.T) {
	parser := gh.NewRunViewParser()
	result, err := parser.Parse(strings.NewReader(invalidJSON))

	require.NoError(t, err)
	require.NotNil(t, result.Error)
}
