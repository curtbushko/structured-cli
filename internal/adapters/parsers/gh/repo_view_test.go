package gh_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/gh"
)

func TestRepoViewParser_ParseFullRepo(t *testing.T) {
	input := `{
  "name": "structured-cli",
  "nameWithOwner": "curtbushko/structured-cli",
  "description": "A universal CLI wrapper that transforms raw CLI output into structured JSON",
  "url": "https://github.com/curtbushko/structured-cli",
  "homepageUrl": "https://structured-cli.dev",
  "defaultBranchRef": {"name": "main"},
  "visibility": "PUBLIC",
  "isFork": false,
  "isArchived": false,
  "isDisabled": false,
  "isTemplate": false,
  "owner": {"login": "curtbushko", "__typename": "User"},
  "primaryLanguage": {"name": "Go"},
  "licenseInfo": {"key": "mit", "name": "MIT License", "spdxId": "MIT"},
  "repositoryTopics": {"nodes": [{"topic": {"name": "cli"}}, {"topic": {"name": "golang"}}]},
  "stargazerCount": 150,
  "watchers": {"totalCount": 10},
  "forkCount": 25,
  "issues": {"totalCount": 5},
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-06-15T12:30:00Z",
  "pushedAt": "2024-06-15T12:00:00Z"
}`

	parser := gh.NewRepoViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	repo, ok := result.Data.(*gh.RepoViewResult)
	require.True(t, ok, "result.Data should be *gh.RepoViewResult")

	assert.Equal(t, "structured-cli", repo.Name)
	assert.Equal(t, "curtbushko/structured-cli", repo.FullName)
	assert.Equal(t, "A universal CLI wrapper that transforms raw CLI output into structured JSON", repo.Description)
	assert.Equal(t, "https://github.com/curtbushko/structured-cli", repo.URL)
	require.NotNil(t, repo.HomepageURL)
	assert.Equal(t, "https://structured-cli.dev", *repo.HomepageURL)
	assert.Equal(t, "main", repo.DefaultBranch)
	assert.Equal(t, "PUBLIC", repo.Visibility)
	assert.False(t, repo.Fork)
	assert.False(t, repo.Archived)
	assert.False(t, repo.Disabled)
	assert.False(t, repo.Template)
	assert.Equal(t, "curtbushko", repo.Owner.Login)
	assert.Equal(t, "User", repo.Owner.Type)
	require.NotNil(t, repo.Language)
	assert.Equal(t, "Go", *repo.Language)
	require.NotNil(t, repo.License)
	assert.Equal(t, "mit", repo.License.Key)
	assert.Equal(t, "MIT License", repo.License.Name)
	assert.Equal(t, "MIT", repo.License.SPDXID)
	assert.Equal(t, []string{"cli", "golang"}, repo.Topics)
	assert.Equal(t, 150, repo.Stars)
	assert.Equal(t, 10, repo.Watchers)
	assert.Equal(t, 25, repo.Forks)
	assert.Equal(t, 5, repo.OpenIssues)
	assert.Equal(t, "2024-01-01T00:00:00Z", repo.CreatedAt)
	assert.Equal(t, "2024-06-15T12:30:00Z", repo.UpdatedAt)
	assert.Equal(t, "2024-06-15T12:00:00Z", repo.PushedAt)
}

func TestRepoViewParser_ParseMinimalRepo(t *testing.T) {
	input := `{
  "name": "minimal-repo",
  "nameWithOwner": "user/minimal-repo",
  "description": "",
  "url": "https://github.com/user/minimal-repo",
  "homepageUrl": "",
  "defaultBranchRef": {"name": "main"},
  "visibility": "PRIVATE",
  "isFork": false,
  "isArchived": false,
  "isDisabled": false,
  "isTemplate": false,
  "owner": {"login": "user", "__typename": "User"},
  "primaryLanguage": null,
  "licenseInfo": null,
  "repositoryTopics": {"nodes": []},
  "stargazerCount": 0,
  "watchers": {"totalCount": 1},
  "forkCount": 0,
  "issues": {"totalCount": 0},
  "createdAt": "2024-01-01T00:00:00Z",
  "updatedAt": "2024-01-01T00:00:00Z",
  "pushedAt": "2024-01-01T00:00:00Z"
}`

	parser := gh.NewRepoViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	repo, ok := result.Data.(*gh.RepoViewResult)
	require.True(t, ok, "result.Data should be *gh.RepoViewResult")

	assert.Equal(t, "minimal-repo", repo.Name)
	assert.Equal(t, "user/minimal-repo", repo.FullName)
	assert.Empty(t, repo.Description)
	assert.Nil(t, repo.HomepageURL)
	assert.Equal(t, "PRIVATE", repo.Visibility)
	assert.Nil(t, repo.Language)
	assert.Nil(t, repo.License)
	assert.Empty(t, repo.Topics)
	assert.Equal(t, 0, repo.Stars)
}

func TestRepoViewParser_ParseForkedRepo(t *testing.T) {
	input := `{
  "name": "forked-repo",
  "nameWithOwner": "user/forked-repo",
  "description": "A forked repository",
  "url": "https://github.com/user/forked-repo",
  "homepageUrl": "",
  "defaultBranchRef": {"name": "main"},
  "visibility": "PUBLIC",
  "isFork": true,
  "isArchived": false,
  "isDisabled": false,
  "isTemplate": false,
  "owner": {"login": "user", "__typename": "User"},
  "primaryLanguage": {"name": "Python"},
  "licenseInfo": {"key": "apache-2.0", "name": "Apache License 2.0", "spdxId": "Apache-2.0"},
  "repositoryTopics": {"nodes": []},
  "stargazerCount": 5,
  "watchers": {"totalCount": 2},
  "forkCount": 0,
  "issues": {"totalCount": 1},
  "createdAt": "2024-03-01T00:00:00Z",
  "updatedAt": "2024-03-15T00:00:00Z",
  "pushedAt": "2024-03-14T00:00:00Z"
}`

	parser := gh.NewRepoViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	repo, ok := result.Data.(*gh.RepoViewResult)
	require.True(t, ok, "result.Data should be *gh.RepoViewResult")

	assert.True(t, repo.Fork)
	assert.Equal(t, "PUBLIC", repo.Visibility)
	require.NotNil(t, repo.Language)
	assert.Equal(t, "Python", *repo.Language)
}

func TestRepoViewParser_ParseOrganizationOwner(t *testing.T) {
	input := `{
  "name": "org-repo",
  "nameWithOwner": "myorg/org-repo",
  "description": "Organization repository",
  "url": "https://github.com/myorg/org-repo",
  "homepageUrl": "",
  "defaultBranchRef": {"name": "main"},
  "visibility": "INTERNAL",
  "isFork": false,
  "isArchived": false,
  "isDisabled": false,
  "isTemplate": false,
  "owner": {"login": "myorg", "__typename": "Organization"},
  "primaryLanguage": {"name": "TypeScript"},
  "licenseInfo": null,
  "repositoryTopics": {"nodes": []},
  "stargazerCount": 10,
  "watchers": {"totalCount": 5},
  "forkCount": 2,
  "issues": {"totalCount": 3},
  "createdAt": "2023-01-01T00:00:00Z",
  "updatedAt": "2024-01-15T00:00:00Z",
  "pushedAt": "2024-01-14T00:00:00Z"
}`

	parser := gh.NewRepoViewParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	repo, ok := result.Data.(*gh.RepoViewResult)
	require.True(t, ok, "result.Data should be *gh.RepoViewResult")

	assert.Equal(t, "myorg", repo.Owner.Login)
	assert.Equal(t, "Organization", repo.Owner.Type)
	assert.Equal(t, "INTERNAL", repo.Visibility)
}

func TestRepoViewParser_Schema(t *testing.T) {
	parser := gh.NewRepoViewParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "full_name")
	assert.Contains(t, schema.Properties, "owner")
	assert.Contains(t, schema.Properties, "visibility")
	assert.Contains(t, schema.Properties, "stars")
	assert.Contains(t, schema.Properties, "forks")
}

func TestRepoViewParser_Matches(t *testing.T) {
	parser := gh.NewRepoViewParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"gh repo view", "gh", []string{"repo", "view"}, true},
		{"gh repo view with repo name", "gh", []string{"repo", "view", "owner/repo"}, true},
		{"gh repo view with flags", "gh", []string{"repo", "view", "--json"}, true},
		{"gh repo list", "gh", []string{"repo", "list"}, false},
		{"gh repo clone", "gh", []string{"repo", "clone"}, false},
		{"gh pr view", "gh", []string{"pr", "view"}, false},
		{"gh only", "gh", []string{}, false},
		{"gh repo only", "gh", []string{"repo"}, false},
		{"git repo view", "git", []string{"repo", "view"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRepoViewParser_InvalidJSON(t *testing.T) {
	parser := gh.NewRepoViewParser()
	result, err := parser.Parse(strings.NewReader(invalidJSON))

	require.NoError(t, err)
	require.NotNil(t, result.Error)
}
