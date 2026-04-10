# Phase 26: Helm Parsers

Helm package manager integration.

## Parsers

- [x] `helm list` - Release listing with status, chart, app version
- [x] `helm status` - Release status with resources
- [x] `helm history` - Release history with revisions
- [x] `helm search repo` - Chart search results
- [x] `helm show values` - Chart default values

## E2E Tests

- [x] `helm list` returns structured release data
- [x] `helm status` returns release status
- [x] Graceful handling when helm not installed
