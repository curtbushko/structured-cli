package helm

import (
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Column names for helm history output.
const (
	colDescription = "DESCRIPTION"
)

// History column names for helm history output.
var historyColumnNames = []string{colRevision, colUpdated, colStatus, colChart, colAppVersion, colDescription}

// Required columns for history output validation.
var historyRequiredColumns = []string{colRevision, colStatus, colChart}

// HistoryParser parses the output of 'helm history'.
type HistoryParser struct {
	schema domain.Schema
}

// NewHistoryParser creates a new HistoryParser with the helm-history schema.
func NewHistoryParser() *HistoryParser {
	return &HistoryParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/helm-history.json",
			"Helm History Output",
			"object",
			map[string]domain.PropertySchema{
				"revisions": {Type: "array", Description: "List of release revisions"},
			},
			[]string{"revisions"},
		),
	}
}

// Parse reads helm history output and returns structured data.
func (p *HistoryParser) Parse(r io.Reader) (domain.ParseResult, error) {
	prep := readAndPrepare(r, historyColumnNames, historyRequiredColumns)

	if prep.Error != nil {
		return emptyResultWithError(prep.Error, prep.ErrorMsg), nil
	}

	emptyResult := &HistoryResult{Revisions: []Revision{}}
	if prep.IsEmpty {
		return emptyResultOK(emptyResult, prep.Input.Raw), nil
	}

	revisions := parseLines(prep.Input.Scanner, prep.Input.Columns, parseRevisionLine)

	return domain.NewParseResult(&HistoryResult{Revisions: revisions}, prep.Input.Raw, 0), nil
}

// Schema returns the JSON Schema for helm history output.
func (p *HistoryParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *HistoryParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdHelm || len(subcommands) < 1 {
		return false
	}
	return subcommands[0] == cmdHistory
}

// parseRevisionLine parses a single line of helm history output.
func parseRevisionLine(line string, columns []columnInfo) Revision {
	revision := Revision{}

	for _, col := range columns {
		value := extractColumnValue(line, col)
		setRevisionField(&revision, col.name, value)
	}

	return revision
}

// setRevisionField sets a field on the revision based on column name.
func setRevisionField(revision *Revision, colNameVal, value string) {
	switch colNameVal {
	case colRevision:
		revision.Revision = parseInt(value)
	case colUpdated:
		revision.Updated = value
	case colStatus:
		revision.Status = value
	case colChart:
		revision.Chart = value
	case colAppVersion:
		revision.AppVersion = value
	case colDescription:
		revision.Description = value
	}
}
