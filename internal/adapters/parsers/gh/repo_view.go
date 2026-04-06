package gh

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// RepoOwner represents the owner of a GitHub repository.
type RepoOwner struct {
	// Login is the owner username.
	Login string `json:"login"`

	// Type is the owner type (User, Organization).
	Type string `json:"type"`
}

// License represents repository license information.
type License struct {
	// Key is the license key (e.g., mit, apache-2.0).
	Key string `json:"key"`

	// Name is the license name.
	Name string `json:"name"`

	// SPDXID is the SPDX identifier.
	SPDXID string `json:"spdx_id"`
}

// RepoViewResult represents the structured output of 'gh repo view'.
type RepoViewResult struct {
	// Name is the repository name.
	Name string `json:"name"`

	// FullName is the full repository name (owner/repo).
	FullName string `json:"full_name"`

	// Description is the repository description.
	Description string `json:"description"`

	// URL is the repository URL.
	URL string `json:"url"`

	// HomepageURL is the homepage URL if set.
	HomepageURL *string `json:"homepage_url,omitempty"`

	// DefaultBranch is the default branch name.
	DefaultBranch string `json:"default_branch"`

	// Visibility is the repository visibility (PUBLIC, PRIVATE, INTERNAL).
	Visibility string `json:"visibility"`

	// Fork indicates whether the repo is a fork.
	Fork bool `json:"fork"`

	// Archived indicates whether the repo is archived.
	Archived bool `json:"archived"`

	// Disabled indicates whether the repo is disabled.
	Disabled bool `json:"disabled"`

	// Template indicates whether the repo is a template.
	Template bool `json:"template"`

	// Owner is the repository owner.
	Owner RepoOwner `json:"owner"`

	// Language is the primary language if detected.
	Language *string `json:"language,omitempty"`

	// License is the license information if set.
	License *License `json:"license,omitempty"`

	// Topics are the repository topics.
	Topics []string `json:"topics"`

	// Stars is the star count.
	Stars int `json:"stars"`

	// Watchers is the watcher count.
	Watchers int `json:"watchers"`

	// Forks is the fork count.
	Forks int `json:"forks"`

	// OpenIssues is the open issue count.
	OpenIssues int `json:"open_issues"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at"`

	// PushedAt is the last push timestamp.
	PushedAt string `json:"pushed_at"`
}

// ghRepoViewItem represents the JSON output format from gh repo view --json.
type ghRepoViewItem struct {
	Name             string              `json:"name"`
	NameWithOwner    string              `json:"nameWithOwner"`
	Description      string              `json:"description"`
	URL              string              `json:"url"`
	HomepageURL      string              `json:"homepageUrl"`
	DefaultBranchRef *ghDefaultBranchRef `json:"defaultBranchRef"`
	Visibility       string              `json:"visibility"`
	IsFork           bool                `json:"isFork"`
	IsArchived       bool                `json:"isArchived"`
	IsDisabled       bool                `json:"isDisabled"`
	IsTemplate       bool                `json:"isTemplate"`
	Owner            ghRepoOwner         `json:"owner"`
	PrimaryLanguage  *ghPrimaryLanguage  `json:"primaryLanguage"`
	LicenseInfo      *ghLicenseInfo      `json:"licenseInfo"`
	RepositoryTopics ghRepositoryTopics  `json:"repositoryTopics"`
	StargazerCount   int                 `json:"stargazerCount"`
	Watchers         ghTotalCountWrapper `json:"watchers"`
	ForkCount        int                 `json:"forkCount"`
	Issues           ghTotalCountWrapper `json:"issues"`
	CreatedAt        string              `json:"createdAt"`
	UpdatedAt        string              `json:"updatedAt"`
	PushedAt         string              `json:"pushedAt"`
}

// ghDefaultBranchRef represents the default branch in gh JSON output.
type ghDefaultBranchRef struct {
	Name string `json:"name"`
}

// ghRepoOwner represents the owner in gh JSON output.
type ghRepoOwner struct {
	Login    string `json:"login"`
	Typename string `json:"__typename"`
}

// ghPrimaryLanguage represents the primary language in gh JSON output.
type ghPrimaryLanguage struct {
	Name string `json:"name"`
}

// ghLicenseInfo represents license info in gh JSON output.
type ghLicenseInfo struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SPDXID string `json:"spdxId"`
}

// ghRepositoryTopics represents repository topics in gh JSON output.
type ghRepositoryTopics struct {
	Nodes []ghTopicNode `json:"nodes"`
}

// ghTopicNode represents a topic node in gh JSON output.
type ghTopicNode struct {
	Topic ghTopic `json:"topic"`
}

// ghTopic represents a topic in gh JSON output.
type ghTopic struct {
	Name string `json:"name"`
}

// ghTotalCountWrapper wraps a totalCount field from gh JSON output.
type ghTotalCountWrapper struct {
	TotalCount int `json:"totalCount"`
}

// RepoViewParser parses the output of 'gh repo view'.
type RepoViewParser struct {
	schema domain.Schema
}

// NewRepoViewParser creates a new RepoViewParser with the gh-repo-view schema.
func NewRepoViewParser() *RepoViewParser {
	return &RepoViewParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/gh-repo-view.json",
			"GitHub Repo View Output",
			"object",
			map[string]domain.PropertySchema{
				"name":           {Type: "string", Description: "Repository name"},
				"full_name":      {Type: "string", Description: "Full repository name (owner/repo)"},
				"description":    {Type: "string", Description: "Repository description"},
				"url":            {Type: "string", Description: "Repository URL"},
				"homepage_url":   {Type: "string", Description: "Homepage URL if set"},
				"default_branch": {Type: "string", Description: "Default branch name"},
				"visibility":     {Type: "string", Description: "Visibility (PUBLIC, PRIVATE, INTERNAL)"},
				"fork":           {Type: "boolean", Description: "Whether repo is a fork"},
				"archived":       {Type: "boolean", Description: "Whether repo is archived"},
				"disabled":       {Type: "boolean", Description: "Whether repo is disabled"},
				"template":       {Type: "boolean", Description: "Whether repo is a template"},
				"owner":          {Type: "object", Description: "Repository owner"},
				"language":       {Type: "string", Description: "Primary language"},
				"license":        {Type: "object", Description: "License information"},
				"topics":         {Type: "array", Description: "Repository topics"},
				"stars":          {Type: "integer", Description: "Star count"},
				"watchers":       {Type: "integer", Description: "Watcher count"},
				"forks":          {Type: "integer", Description: "Fork count"},
				"open_issues":    {Type: "integer", Description: "Open issue count"},
				"created_at":     {Type: "string", Description: "Creation timestamp"},
				"updated_at":     {Type: "string", Description: "Last update timestamp"},
				"pushed_at":      {Type: "string", Description: "Last push timestamp"},
			},
			[]string{"name", "full_name", "owner", "visibility"},
		),
	}
}

// Parse reads gh repo view JSON output and returns structured data.
func (p *RepoViewParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			"",
			0,
		), nil
	}

	raw := string(data)

	var item ghRepoViewItem
	if err := json.Unmarshal(data, &item); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("parse JSON: %w", err),
			raw,
			0,
		), nil
	}

	// Convert to our domain types
	result := &RepoViewResult{
		Name:        item.Name,
		FullName:    item.NameWithOwner,
		Description: item.Description,
		URL:         item.URL,
		Visibility:  item.Visibility,
		Fork:        item.IsFork,
		Archived:    item.IsArchived,
		Disabled:    item.IsDisabled,
		Template:    item.IsTemplate,
		Owner:       RepoOwner{Login: item.Owner.Login, Type: item.Owner.Typename},
		Topics:      convertTopics(item.RepositoryTopics),
		Stars:       item.StargazerCount,
		Watchers:    item.Watchers.TotalCount,
		Forks:       item.ForkCount,
		OpenIssues:  item.Issues.TotalCount,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		PushedAt:    item.PushedAt,
	}

	// Handle default branch
	if item.DefaultBranchRef != nil {
		result.DefaultBranch = item.DefaultBranchRef.Name
	}

	// Handle optional homepage URL
	if item.HomepageURL != "" {
		result.HomepageURL = &item.HomepageURL
	}

	// Handle optional primary language
	if item.PrimaryLanguage != nil {
		result.Language = &item.PrimaryLanguage.Name
	}

	// Handle optional license
	if item.LicenseInfo != nil {
		result.License = &License{
			Key:    item.LicenseInfo.Key,
			Name:   item.LicenseInfo.Name,
			SPDXID: item.LicenseInfo.SPDXID,
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for gh repo view output.
func (p *RepoViewParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RepoViewParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "gh" {
		return false
	}
	if len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == "repo" && subcommands[1] == "view"
}

// convertTopics converts gh repository topics to a string slice.
func convertTopics(topics ghRepositoryTopics) []string {
	result := make([]string, 0, len(topics.Nodes))
	for _, node := range topics.Nodes {
		result = append(result, node.Topic.Name)
	}
	return result
}
