package kubectl

import (
	"io"
	"regexp"
	"strconv"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regular expressions for parsing kubectl get pods output.
var (
	// restartsWithAge matches restart count with optional age like "5 (2h ago)"
	restartsWithAge = regexp.MustCompile(`^(\d+)`)
)

// Pod column names for kubectl get pods.
var podColumnNames = []string{"NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "AGE", "IP", "NODE"}

// Required columns for pod output validation.
var podRequiredColumns = []string{"NAME", "READY", "STATUS"}

// GetPodsParser parses the output of 'kubectl get pods'.
type GetPodsParser struct {
	schema domain.Schema
}

// NewGetPodsParser creates a new GetPodsParser with the kubectl-get-pods schema.
func NewGetPodsParser() *GetPodsParser {
	return &GetPodsParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/kubectl-get-pods.json",
			"Kubectl Get Pods Output",
			"object",
			map[string]domain.PropertySchema{
				"pods": {Type: "array", Description: "List of pods"},
			},
			[]string{"pods"},
		),
	}
}

// Parse reads kubectl get pods output and returns structured data.
func (p *GetPodsParser) Parse(r io.Reader) (domain.ParseResult, error) {
	prep := readAndPrepare(r, podColumnNames, podRequiredColumns)

	if prep.Error != nil {
		return emptyResultWithError(prep.Error, prep.ErrorMsg), nil
	}

	emptyResult := &GetPodsResult{Pods: []Pod{}}
	if prep.IsEmpty {
		return emptyResultOK(emptyResult, prep.Input.Raw), nil
	}

	pods := parseLines(prep.Input.Scanner, prep.Input.Columns, parsePodLine)

	return domain.NewParseResult(&GetPodsResult{Pods: pods}, prep.Input.Raw, 0), nil
}

// Schema returns the JSON Schema for kubectl get pods output.
func (p *GetPodsParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *GetPodsParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdKubectl || len(subcommands) < 2 || subcommands[0] != cmdGet {
		return false
	}
	// Match "pods", "pod", or "po" (the common aliases)
	resource := subcommands[1]
	return resource == "pods" || resource == "pod" || resource == "po"
}

// parsePodLine parses a single line of kubectl get pods output.
func parsePodLine(line string, columns []columnInfo) Pod {
	pod := Pod{}

	for _, col := range columns {
		value := extractColumnValue(line, col)
		setPodField(&pod, col.name, value)
	}

	return pod
}

// setPodField sets a field on the pod based on column name.
func setPodField(pod *Pod, colNameVal, value string) {
	switch colNameVal {
	case colNamespace:
		pod.Namespace = value
	case colName:
		pod.Name = value
	case "READY":
		pod.Ready = value
	case "STATUS":
		pod.Status = value
	case "RESTARTS":
		pod.Restarts = parseRestarts(value)
	case colAge:
		pod.Age = value
	case "IP":
		pod.IP = value
	case "NODE":
		pod.Node = value
	}
}

// parseRestarts extracts the restart count from strings like "5" or "5 (2h ago)".
func parseRestarts(s string) int {
	matches := restartsWithAge.FindStringSubmatch(s)
	if len(matches) > 1 {
		count, _ := strconv.Atoi(matches[1])
		return count
	}
	return 0
}
