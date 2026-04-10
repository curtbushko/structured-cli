# Phase 24: GitHub CLI (gh) Parsers

GitHub CLI integration for repository, PR, and issue management.

## Parsers

- [x] `gh pr list` - List pull requests with status, author, labels
- [x] `gh pr view` - PR details with reviews, checks, comments
- [x] `gh pr status` - PR status for current branch
- [x] `gh issue list` - List issues with labels, assignees
- [x] `gh issue view` - Issue details with comments
- [x] `gh repo view` - Repository metadata
- [x] `gh run list` - Workflow runs with status
- [x] `gh run view` - Workflow run details with jobs

## E2E Tests

- [x] `gh pr list` returns structured PR data
- [x] `gh issue list` returns structured issue data
- [x] `gh run list` returns workflow status
- [x] Graceful handling when gh not installed
