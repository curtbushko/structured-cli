package helm

import (
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Column names for helm search repo output.
const (
	colChartVersion = "CHART VERSION"
)

// searchRepoColumnNames defines column names in order for helm search repo output.
var searchRepoColumnNames = []string{colName, colChartVersion, colAppVersion, colDescription}

// searchRepoRequiredColumns defines required columns for validation.
var searchRepoRequiredColumns = []string{colName, colChartVersion}

// SearchRepoParser parses the output of 'helm search repo'.
type SearchRepoParser struct {
	schema domain.Schema
}

// NewSearchRepoParser creates a new SearchRepoParser with the helm-search-repo schema.
func NewSearchRepoParser() *SearchRepoParser {
	return &SearchRepoParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/helm-search-repo.json",
			"Helm Search Repo Output",
			"object",
			map[string]domain.PropertySchema{
				"charts": {Type: "array", Description: "List of charts matching the search"},
			},
			[]string{"charts"},
		),
	}
}

// Parse reads helm search repo output and returns structured data.
func (p *SearchRepoParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return emptyResultWithError(err, ""), nil
	}

	raw := string(data)
	trimmed := strings.TrimSpace(raw)

	// Handle empty output or "No results found"
	if trimmed == "" || strings.HasPrefix(trimmed, "No results found") {
		return emptyResultOK(&SearchResult{Charts: []ChartInfo{}}, raw), nil
	}

	prep := readAndPrepare(strings.NewReader(raw), searchRepoColumnNames, searchRepoRequiredColumns)

	if prep.Error != nil {
		return emptyResultWithError(prep.Error, prep.ErrorMsg), nil
	}

	emptyResult := &SearchResult{Charts: []ChartInfo{}}
	if prep.IsEmpty {
		return emptyResultOK(emptyResult, prep.Input.Raw), nil
	}

	charts := parseLines(prep.Input.Scanner, prep.Input.Columns, parseChartLine)

	return domain.NewParseResult(&SearchResult{Charts: charts}, prep.Input.Raw, 0), nil
}

// Schema returns the JSON Schema for helm search repo output.
func (p *SearchRepoParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *SearchRepoParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdHelm || len(subcommands) < 2 {
		return false
	}
	// Match "search repo"
	return subcommands[0] == cmdSearch && subcommands[1] == "repo"
}

// parseChartLine parses a single line of helm search repo output.
func parseChartLine(line string, columns []columnInfo) ChartInfo {
	chart := ChartInfo{}

	for _, col := range columns {
		value := extractColumnValue(line, col)
		setChartField(&chart, col.name, value)
	}

	return chart
}

// setChartField sets a field on the chart based on column name.
func setChartField(chart *ChartInfo, colNameVal, value string) {
	switch colNameVal {
	case colName:
		chart.Name = value
	case colChartVersion:
		chart.ChartVersion = value
	case colAppVersion:
		chart.AppVersion = value
	case colDescription:
		chart.Description = value
	}
}
