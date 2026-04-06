// Package gh provides parsers for GitHub CLI (gh) command output.
// This package is in the adapters layer and implements parsers for
// converting raw gh command output into structured domain types.
package gh

// Subcommand constants for gh commands.
const (
	subCmdList  = "list"
	subCmdView  = "view"
	subCmdIssue = "issue"
	subCmdPR    = "pr"
	subCmdRepo  = "repo"
	subCmdRun   = "run"
)

// Author represents a GitHub user.
type Author struct {
	// Login is the GitHub username.
	Login string `json:"login"`

	// Name is the display name of the user.
	Name string `json:"name,omitempty"`
}

// Label represents a GitHub label.
type Label struct {
	// Name is the label name.
	Name string `json:"name"`

	// Color is the hex color code (without #).
	Color string `json:"color,omitempty"`

	// Description is the label description.
	Description string `json:"description,omitempty"`
}

// Review represents a pull request review.
type Review struct {
	// Author is the reviewer.
	Author Author `json:"author"`

	// State is the review state (PENDING, APPROVED, CHANGES_REQUESTED, COMMENTED, DISMISSED).
	State string `json:"state"`

	// Body is the review comment body.
	Body string `json:"body,omitempty"`

	// SubmittedAt is the timestamp when the review was submitted.
	SubmittedAt string `json:"submitted_at,omitempty"`
}

// Check represents a CI/CD check status.
type Check struct {
	// Name is the name of the check.
	Name string `json:"name"`

	// Status is the check status (pending, passing, failing).
	Status string `json:"status"`

	// Conclusion is the check conclusion (success, failure, neutral, etc).
	Conclusion string `json:"conclusion,omitempty"`

	// URL is the link to the check details.
	URL string `json:"url,omitempty"`
}

// PullRequest represents a GitHub pull request.
type PullRequest struct {
	// Number is the PR number.
	Number int `json:"number"`

	// Title is the PR title.
	Title string `json:"title"`

	// State is the PR state (OPEN, CLOSED, MERGED).
	State string `json:"state"`

	// Author is the PR author.
	Author Author `json:"author"`

	// Labels are the PR labels.
	Labels []Label `json:"labels"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at"`

	// URL is the PR URL.
	URL string `json:"url"`

	// HeadBranch is the head branch name.
	HeadBranch string `json:"head_branch"`

	// BaseBranch is the base branch name.
	BaseBranch string `json:"base_branch"`

	// Draft indicates whether the PR is a draft.
	Draft bool `json:"draft"`
}

// PRListResult represents the structured output of 'gh pr list'.
type PRListResult struct {
	// PullRequests is the list of pull requests.
	PullRequests []PullRequest `json:"pull_requests"`
}

// PRViewResult represents the structured output of 'gh pr view'.
type PRViewResult struct {
	// Number is the PR number.
	Number int `json:"number"`

	// Title is the PR title.
	Title string `json:"title"`

	// Body is the PR description body.
	Body string `json:"body"`

	// State is the PR state.
	State string `json:"state"`

	// Author is the PR author.
	Author Author `json:"author"`

	// Labels are the PR labels.
	Labels []Label `json:"labels"`

	// Assignees are the PR assignees.
	Assignees []Author `json:"assignees"`

	// Reviewers are the PR reviewers.
	Reviewers []Author `json:"reviewers"`

	// Reviews are the PR reviews.
	Reviews []Review `json:"reviews"`

	// Checks are the CI/CD checks.
	Checks []Check `json:"checks"`

	// Comments is the number of comments.
	Comments int `json:"comments"`

	// Additions is the number of lines added.
	Additions int `json:"additions"`

	// Deletions is the number of lines deleted.
	Deletions int `json:"deletions"`

	// ChangedFiles is the number of changed files.
	ChangedFiles int `json:"changed_files"`

	// Mergeable is the mergeable state (MERGEABLE, CONFLICTING, UNKNOWN).
	Mergeable string `json:"mergeable"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at"`

	// URL is the PR URL.
	URL string `json:"url"`
}

// PRSummary represents a brief summary of a pull request.
type PRSummary struct {
	// Number is the PR number.
	Number int `json:"number"`

	// Title is the PR title.
	Title string `json:"title"`

	// HeadBranch is the head branch name.
	HeadBranch string `json:"head_branch"`

	// URL is the PR URL.
	URL string `json:"url"`
}

// CurrentBranchPR represents the PR for the current branch.
type CurrentBranchPR struct {
	// Number is the PR number.
	Number int `json:"number"`

	// Title is the PR title.
	Title string `json:"title"`

	// HeadBranch is the head branch name.
	HeadBranch string `json:"head_branch"`

	// URL is the PR URL.
	URL string `json:"url"`

	// State is the PR state.
	State string `json:"state"`

	// ReviewStatus is the status of reviews (e.g., "Approved", "Changes requested").
	ReviewStatus string `json:"review_status,omitempty"`

	// CheckStatus is the status of checks (e.g., "All checks passing").
	CheckStatus string `json:"check_status,omitempty"`
}

// PRStatusResult represents the structured output of 'gh pr status'.
type PRStatusResult struct {
	// CurrentBranch is the PR for the current branch, or nil if none.
	CurrentBranch *CurrentBranchPR `json:"current_branch"`

	// CreatedByYou are PRs created by the authenticated user.
	CreatedByYou []PRSummary `json:"created_by_you"`

	// RequestingReview are PRs requesting your review.
	RequestingReview []PRSummary `json:"requesting_review"`
}

// Issue represents a GitHub issue.
type Issue struct {
	// Number is the issue number.
	Number int `json:"number"`

	// Title is the issue title.
	Title string `json:"title"`

	// State is the issue state (OPEN, CLOSED).
	State string `json:"state"`

	// Author is the issue author.
	Author Author `json:"author"`

	// Labels are the issue labels.
	Labels []Label `json:"labels"`

	// Assignees are the issue assignees.
	Assignees []Author `json:"assignees"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at"`

	// URL is the issue URL.
	URL string `json:"url"`

	// Comments is the number of comments.
	Comments int `json:"comments"`
}

// IssueListResult represents the structured output of 'gh issue list'.
type IssueListResult struct {
	// Issues is the list of issues.
	Issues []Issue `json:"issues"`
}

// Milestone represents a GitHub milestone.
type Milestone struct {
	// Title is the milestone title.
	Title string `json:"title"`

	// Number is the milestone number.
	Number int `json:"number"`

	// State is the milestone state (open, closed).
	State string `json:"state"`
}

// Project represents a GitHub project.
type Project struct {
	// Title is the project title.
	Title string `json:"title"`

	// Number is the project number.
	Number int `json:"number"`
}

// Reactions represents reaction counts on an issue or PR.
type Reactions struct {
	// ThumbsUp is the count of thumbs up reactions.
	ThumbsUp int `json:"thumbs_up"`

	// ThumbsDown is the count of thumbs down reactions.
	ThumbsDown int `json:"thumbs_down"`

	// Laugh is the count of laugh reactions.
	Laugh int `json:"laugh"`

	// Hooray is the count of hooray reactions.
	Hooray int `json:"hooray"`

	// Confused is the count of confused reactions.
	Confused int `json:"confused"`

	// Heart is the count of heart reactions.
	Heart int `json:"heart"`

	// Rocket is the count of rocket reactions.
	Rocket int `json:"rocket"`

	// Eyes is the count of eyes reactions.
	Eyes int `json:"eyes"`
}

// IssueViewResult represents the structured output of 'gh issue view'.
type IssueViewResult struct {
	// Number is the issue number.
	Number int `json:"number"`

	// Title is the issue title.
	Title string `json:"title"`

	// Body is the issue description body.
	Body string `json:"body"`

	// State is the issue state.
	State string `json:"state"`

	// Author is the issue author.
	Author Author `json:"author"`

	// Labels are the issue labels.
	Labels []Label `json:"labels"`

	// Assignees are the issue assignees.
	Assignees []Author `json:"assignees"`

	// Milestone is the issue milestone, or nil if not set.
	Milestone *Milestone `json:"milestone"`

	// Project is the issue project, or nil if not set.
	Project *Project `json:"project"`

	// Reactions contains the reaction counts.
	Reactions Reactions `json:"reactions"`

	// Comments is the number of comments.
	Comments int `json:"comments"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at"`

	// ClosedAt is the close timestamp if closed, or nil if open.
	ClosedAt *string `json:"closed_at"`

	// URL is the issue URL.
	URL string `json:"url"`
}

// Step represents a step in a GitHub Actions job.
type Step struct {
	// Name is the step name.
	Name string `json:"name"`

	// Status is the step status (queued, in_progress, completed).
	Status string `json:"status"`

	// Conclusion is the step conclusion (success, failure, etc), null if in progress.
	Conclusion *string `json:"conclusion"`

	// Number is the step number.
	Number int `json:"number"`
}

// Job represents a job in a GitHub Actions workflow run.
type Job struct {
	// DatabaseID is the job ID.
	DatabaseID int64 `json:"database_id"`

	// Name is the job name.
	Name string `json:"name"`

	// Status is the job status (queued, in_progress, completed).
	Status string `json:"status"`

	// Conclusion is the job conclusion (success, failure, etc), null if in progress.
	Conclusion *string `json:"conclusion"`

	// StartedAt is the job start timestamp, null if not started.
	StartedAt *string `json:"started_at"`

	// CompletedAt is the job completion timestamp, null if not completed.
	CompletedAt *string `json:"completed_at"`

	// Steps are the steps in the job.
	Steps []Step `json:"steps"`
}

// WorkflowRun represents a GitHub Actions workflow run.
type WorkflowRun struct {
	// DatabaseID is the run ID.
	DatabaseID int64 `json:"database_id"`

	// DisplayTitle is the run display title.
	DisplayTitle string `json:"display_title"`

	// Status is the run status (queued, in_progress, completed).
	Status string `json:"status"`

	// Conclusion is the run conclusion (success, failure, etc), null if in progress.
	Conclusion *string `json:"conclusion"`

	// WorkflowName is the workflow name.
	WorkflowName string `json:"workflow_name"`

	// HeadBranch is the branch that triggered the run.
	HeadBranch string `json:"head_branch"`

	// HeadSha is the commit SHA.
	HeadSha string `json:"head_sha"`

	// Event is the trigger event (push, pull_request, etc).
	Event string `json:"event"`

	// Actor is the user who triggered the run.
	Actor Author `json:"actor"`

	// URL is the run URL.
	URL string `json:"url"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at"`
}

// RunListResult represents the structured output of 'gh run list'.
type RunListResult struct {
	// WorkflowRuns is the list of workflow runs.
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

// RunViewResult represents the structured output of 'gh run view'.
type RunViewResult struct {
	// DatabaseID is the run ID.
	DatabaseID int64 `json:"database_id"`

	// DisplayTitle is the run display title.
	DisplayTitle string `json:"display_title"`

	// Status is the run status (queued, in_progress, completed).
	Status string `json:"status"`

	// Conclusion is the run conclusion (success, failure, etc), null if in progress.
	Conclusion *string `json:"conclusion"`

	// WorkflowName is the workflow name.
	WorkflowName string `json:"workflow_name"`

	// HeadBranch is the branch that triggered the run.
	HeadBranch string `json:"head_branch"`

	// HeadSha is the commit SHA.
	HeadSha string `json:"head_sha"`

	// Event is the trigger event (push, pull_request, etc).
	Event string `json:"event"`

	// Actor is the user who triggered the run.
	Actor Author `json:"actor"`

	// Jobs are the jobs in the workflow run.
	Jobs []Job `json:"jobs"`

	// URL is the run URL.
	URL string `json:"url"`

	// CreatedAt is the creation timestamp.
	CreatedAt string `json:"created_at"`

	// UpdatedAt is the last update timestamp.
	UpdatedAt string `json:"updated_at"`

	// RunStartedAt is the run start timestamp.
	RunStartedAt string `json:"run_started_at"`
}
