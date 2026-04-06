package kubectl

import (
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Node column names for kubectl get nodes.
var nodeColumnNames = []string{
	"NAME", "STATUS", "ROLES", "AGE", "VERSION",
	"INTERNAL-IP", "EXTERNAL-IP", "OS-IMAGE", "KERNEL-VERSION", "CONTAINER-RUNTIME",
}

// Required columns for node output validation.
var nodeRequiredColumns = []string{"NAME", "STATUS", "ROLES", "AGE", "VERSION"}

// GetNodesParser parses the output of 'kubectl get nodes'.
type GetNodesParser struct {
	schema domain.Schema
}

// NewGetNodesParser creates a new GetNodesParser with the kubectl-get-nodes schema.
func NewGetNodesParser() *GetNodesParser {
	return &GetNodesParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/kubectl-get-nodes.json",
			"Kubectl Get Nodes Output",
			"object",
			map[string]domain.PropertySchema{
				"nodes": {Type: "array", Description: "List of nodes"},
			},
			[]string{"nodes"},
		),
	}
}

// Parse reads kubectl get nodes output and returns structured data.
func (p *GetNodesParser) Parse(r io.Reader) (domain.ParseResult, error) {
	prep := readAndPrepare(r, nodeColumnNames, nodeRequiredColumns)

	if prep.Error != nil {
		return emptyResultWithError(prep.Error, prep.ErrorMsg), nil
	}

	emptyResult := &GetNodesResult{Nodes: []Node{}}
	if prep.IsEmpty {
		return emptyResultOK(emptyResult, prep.Input.Raw), nil
	}

	nodes := parseLines(prep.Input.Scanner, prep.Input.Columns, parseNodeLine)

	return domain.NewParseResult(&GetNodesResult{Nodes: nodes}, prep.Input.Raw, 0), nil
}

// Schema returns the JSON Schema for kubectl get nodes output.
func (p *GetNodesParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *GetNodesParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdKubectl || len(subcommands) < 2 || subcommands[0] != cmdGet {
		return false
	}
	// Match "nodes", "node", or "no" (the common aliases)
	resource := subcommands[1]
	return resource == "nodes" || resource == "node" || resource == "no"
}

// parseNodeLine parses a single line of kubectl get nodes output.
func parseNodeLine(line string, columns []columnInfo) Node {
	node := Node{}

	for _, col := range columns {
		value := extractColumnValue(line, col)
		setNodeField(&node, col.name, value)
	}

	return node
}

// setNodeField sets a field on the node based on column name.
func setNodeField(node *Node, colNameVal, value string) {
	switch colNameVal {
	case colName:
		node.Name = value
	case "STATUS":
		node.Status = value
	case "ROLES":
		node.Roles = parseRoles(value)
	case colAge:
		node.Age = value
	case "VERSION":
		node.Version = value
	case "INTERNAL-IP":
		node.InternalIP = normalizeKubectlValue(value)
	case "EXTERNAL-IP":
		node.ExternalIP = normalizeKubectlValue(value)
	case "OS-IMAGE":
		node.OSImage = value
	case "KERNEL-VERSION":
		node.KernelVersion = value
	case "CONTAINER-RUNTIME":
		node.ContainerRuntime = value
	}
}

// parseRoles parses a roles string like "control-plane,master" into a slice.
// Returns nil for "<none>".
func parseRoles(s string) []string {
	if s == "" || s == noneValue {
		return nil
	}
	return strings.Split(s, ",")
}
