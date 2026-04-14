# Phase 31: Bug Fixes & Feature Improvements

## Tasks

- [x] Fix 'go test' output - correctly handle cached test results and include 'ok' status in count
- [x] Fix 'go test' savings calculation - output multiple lines of text instead of single JSON line
- [x] Fix 'docker images' parser - missing images in the list
- [x] Improve 'go test' error output - only output errors in JSON block when tests fail (token savings)
- [x] Filter token savings - ignore all token savings if they are +/- 100

## Notes

- Focus on parsers in `internal/adapters/parsers/golang/` and `internal/adapters/parsers/docker/`
- Ensure stats calculation in tracking system correctly handles multi-line vs JSON output
- Follow TDD: write/update tests first, then fix implementation
- Verify architecture compliance with `go-arch-lint check`
