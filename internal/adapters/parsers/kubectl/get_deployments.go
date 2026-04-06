package kubectl

import (
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Deployment column names for kubectl get deployments.
var deploymentColumnNames = []string{"NAMESPACE", "NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"}

// Required columns for deployment output validation.
var deploymentRequiredColumns = []string{"NAME", "READY", "UP-TO-DATE", "AVAILABLE"}

// GetDeploymentsParser parses the output of 'kubectl get deployments'.
type GetDeploymentsParser struct {
	schema domain.Schema
}

// NewGetDeploymentsParser creates a new GetDeploymentsParser with the kubectl-get-deployments schema.
func NewGetDeploymentsParser() *GetDeploymentsParser {
	return &GetDeploymentsParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/kubectl-get-deployments.json",
			"Kubectl Get Deployments Output",
			"object",
			map[string]domain.PropertySchema{
				"deployments": {Type: "array", Description: "List of deployments"},
			},
			[]string{"deployments"},
		),
	}
}

// Parse reads kubectl get deployments output and returns structured data.
func (p *GetDeploymentsParser) Parse(r io.Reader) (domain.ParseResult, error) {
	prep := readAndPrepare(r, deploymentColumnNames, deploymentRequiredColumns)

	if prep.Error != nil {
		return emptyResultWithError(prep.Error, prep.ErrorMsg), nil
	}

	emptyResult := &GetDeploymentsResult{Deployments: []Deployment{}}
	if prep.IsEmpty {
		return emptyResultOK(emptyResult, prep.Input.Raw), nil
	}

	deployments := parseLines(prep.Input.Scanner, prep.Input.Columns, parseDeploymentLine)

	return domain.NewParseResult(&GetDeploymentsResult{Deployments: deployments}, prep.Input.Raw, 0), nil
}

// Schema returns the JSON Schema for kubectl get deployments output.
func (p *GetDeploymentsParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *GetDeploymentsParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdKubectl || len(subcommands) < 2 || subcommands[0] != cmdGet {
		return false
	}
	// Match "deployments", "deployment", or "deploy" (the common aliases)
	resource := subcommands[1]
	return resource == "deployments" || resource == "deployment" || resource == "deploy"
}

// parseDeploymentLine parses a single line of kubectl get deployments output.
func parseDeploymentLine(line string, columns []columnInfo) Deployment {
	deployment := Deployment{}

	for _, col := range columns {
		value := extractColumnValue(line, col)
		setDeploymentField(&deployment, col.name, value)
	}

	return deployment
}

// setDeploymentField sets a field on the deployment based on column name.
func setDeploymentField(deployment *Deployment, colNameVal, value string) {
	switch colNameVal {
	case colNamespace:
		deployment.Namespace = value
	case colName:
		deployment.Name = value
	case "READY":
		deployment.Ready = value
		deployment.ReadyCount, deployment.DesiredCount = parseReadyCounts(value)
	case "UP-TO-DATE":
		deployment.UpToDate = parseIntOrZero(value)
	case "AVAILABLE":
		deployment.Available = parseIntOrZero(value)
	case colAge:
		deployment.Age = value
	}
}

// parseReadyCounts parses a ready string like "3/3" into ready and desired counts.
func parseReadyCounts(s string) (int, int) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return 0, 0
	}
	ready := parseIntOrZero(parts[0])
	desired := parseIntOrZero(parts[1])
	return ready, desired
}

// parseIntOrZero parses an integer or returns 0 on error.
func parseIntOrZero(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
