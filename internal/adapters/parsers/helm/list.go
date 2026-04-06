package helm

import (
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// List column names for helm list output.
var listColumnNames = []string{colName, colNamespace, colRevision, colUpdated, colStatus, colChart, colAppVersion}

// Required columns for list output validation.
var listRequiredColumns = []string{colName, colStatus, colChart}

// ListParser parses the output of 'helm list'.
type ListParser struct {
	schema domain.Schema
}

// NewListParser creates a new ListParser with the helm-list schema.
func NewListParser() *ListParser {
	return &ListParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/helm-list.json",
			"Helm List Output",
			"object",
			map[string]domain.PropertySchema{
				"releases": {Type: "array", Description: "List of Helm releases"},
			},
			[]string{"releases"},
		),
	}
}

// Parse reads helm list output and returns structured data.
func (p *ListParser) Parse(r io.Reader) (domain.ParseResult, error) {
	prep := readAndPrepare(r, listColumnNames, listRequiredColumns)

	if prep.Error != nil {
		return emptyResultWithError(prep.Error, prep.ErrorMsg), nil
	}

	emptyResult := &ListResult{Releases: []Release{}}
	if prep.IsEmpty {
		return emptyResultOK(emptyResult, prep.Input.Raw), nil
	}

	releases := parseLines(prep.Input.Scanner, prep.Input.Columns, parseReleaseLine)

	return domain.NewParseResult(&ListResult{Releases: releases}, prep.Input.Raw, 0), nil
}

// Schema returns the JSON Schema for helm list output.
func (p *ListParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ListParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdHelm || len(subcommands) < 1 {
		return false
	}
	// Match "list" or "ls" (the common alias)
	return subcommands[0] == cmdList || subcommands[0] == cmdLs
}

// parseReleaseLine parses a single line of helm list output.
func parseReleaseLine(line string, columns []columnInfo) Release {
	release := Release{}

	for _, col := range columns {
		value := extractColumnValue(line, col)
		setReleaseField(&release, col.name, value)
	}

	return release
}

// setReleaseField sets a field on the release based on column name.
func setReleaseField(release *Release, colNameVal, value string) {
	switch colNameVal {
	case colName:
		release.Name = value
	case colNamespace:
		release.Namespace = value
	case colRevision:
		release.Revision = parseInt(value)
	case colUpdated:
		release.Updated = value
	case colStatus:
		release.Status = value
	case colChart:
		release.Chart = value
	case colAppVersion:
		release.AppVersion = value
	}
}
