# Phase 19: End-to-End BDD Tests

Real-world e2e tests using godog with actual CLI tools and file systems.

## Infrastructure

- [x] godog BDD framework setup
- [x] Test binary build automation
- [x] Temporary directory/repo management
- [x] Step definitions for common operations
- [x] Add required tools to flake.nix for CI

## Git Commands E2E

- [x] `git status` - clean repo, untracked files, modified files
- [x] `git log` - basic output, commit fields, empty repo
- [x] `git diff` - unstaged changes, staged changes, hunks
- [x] `git branch` - list branches, current branch detection
- [x] `git show` - commit details
- [x] `git blame` - file attribution
- [x] `git reflog` - reference log

## Output Modes E2E

- [x] Passthrough mode (default)
- [x] JSON output with `--json` flag
- [x] JSON output with environment variable
- [x] `--json` flag position (before/after command)

## Unsupported Commands E2E

- [x] Fallback JSON for unsupported subcommands
- [x] Passthrough for unsupported commands

## Error Handling E2E

- [x] Command failure in JSON mode
- [x] Command failure in passthrough mode
- [x] Parser failure with raw output fallback

## File Operations E2E

- [x] `ls` - list directory, specific path
- [x] `cat` - read file contents
- [x] `head`/`tail` - read first/last lines
- [x] `wc` - word count
- [x] `find` - search by name, type
- [x] `grep` - search in files
- [x] `du` - disk usage
- [x] `df` - disk free

## Go Commands E2E (if go available)

- [x] `go build` - successful build, build errors
- [x] `go test` - run tests, with coverage
- [x] `go vet` - static analysis
- [x] `go fmt` - format check

## NPM Commands E2E (if npm available)

- [x] `npm list` - dependency tree
- [x] `npm outdated` - outdated packages

## Docker Commands E2E (if docker available)

- [x] `docker ps` - list containers
- [x] `docker images` - list images

## Make/Just Commands E2E

- [x] `make` - successful build, target listing
- [x] `just` - recipe execution, listing
