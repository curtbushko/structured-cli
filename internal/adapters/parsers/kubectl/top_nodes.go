package kubectl

import (
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// NodeMetrics represents metrics for a single node.
type NodeMetrics struct {
	// Name is the node name.
	Name string `json:"name"`

	// CPU is the raw CPU usage string (e.g., "500m", "2").
	CPU string `json:"cpu"`

	// CPUCores is the CPU usage in cores (e.g., 0.5, 2.0).
	CPUCores float64 `json:"cpu_cores"`

	// CPUPercent is the CPU usage percentage.
	CPUPercent int `json:"cpu_percent"`

	// Memory is the raw memory usage string (e.g., "4096Mi", "16Gi").
	Memory string `json:"memory"`

	// MemoryBytes is the memory usage in bytes.
	MemoryBytes int64 `json:"memory_bytes"`

	// MemoryPercent is the memory usage percentage.
	MemoryPercent int `json:"memory_percent"`
}

// TopNodesResult represents the structured output of 'kubectl top nodes'.
type TopNodesResult struct {
	// Nodes is the list of node metrics.
	Nodes []NodeMetrics `json:"nodes"`
}

// Top nodes column names.
var topNodesColumnNames = []string{"NAME", "CPU(cores)", "CPU%", "MEMORY(bytes)", "MEMORY%"}

// Required columns for top nodes output validation.
var topNodesRequiredColumns = []string{"NAME", "CPU(cores)", "MEMORY(bytes)"}

// TopNodesParser parses the output of 'kubectl top nodes'.
type TopNodesParser struct {
	schema domain.Schema
}

// NewTopNodesParser creates a new TopNodesParser with the kubectl-top-nodes schema.
func NewTopNodesParser() *TopNodesParser {
	return &TopNodesParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/kubectl-top-nodes.json",
			"Kubectl Top Nodes Output",
			"object",
			map[string]domain.PropertySchema{
				"nodes": {Type: "array", Description: "List of node metrics"},
			},
			[]string{"nodes"},
		),
	}
}

// Parse reads kubectl top nodes output and returns structured data.
func (p *TopNodesParser) Parse(r io.Reader) (domain.ParseResult, error) {
	prep := readAndPrepare(r, topNodesColumnNames, topNodesRequiredColumns)

	if prep.Error != nil {
		return emptyResultWithError(prep.Error, prep.ErrorMsg), nil
	}

	emptyResult := &TopNodesResult{Nodes: []NodeMetrics{}}
	if prep.IsEmpty {
		return emptyResultOK(emptyResult, prep.Input.Raw), nil
	}

	nodes := parseLines(prep.Input.Scanner, prep.Input.Columns, parseTopNodeLine)
	return domain.NewParseResult(&TopNodesResult{Nodes: nodes}, prep.Input.Raw, 0), nil
}

// Schema returns the JSON Schema for kubectl top nodes output.
func (p *TopNodesParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *TopNodesParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdKubectl || len(subcommands) < 2 || subcommands[0] != "top" {
		return false
	}
	// Match "nodes" or "node"
	resource := subcommands[1]
	return resource == "nodes" || resource == "node"
}

// parseTopNodeLine parses a single line of kubectl top nodes output.
func parseTopNodeLine(line string, columns []columnInfo) NodeMetrics {
	node := NodeMetrics{}

	for _, col := range columns {
		value := extractColumnValue(line, col)
		setTopNodeField(&node, col.name, value)
	}

	return node
}

// setTopNodeField sets a field on the node based on column name.
func setTopNodeField(node *NodeMetrics, colNameVal, value string) {
	switch colNameVal {
	case colName:
		node.Name = value
	case colCPU:
		node.CPU = value
		node.CPUCores = parseCPU(value)
	case colCPUPercent:
		node.CPUPercent = parsePercent(value)
	case colMemory:
		node.Memory = value
		node.MemoryBytes = parseMemory(value)
	case colMemoryPercent:
		node.MemoryPercent = parsePercent(value)
	}
}

// parsePercent parses a percentage string like "25%" into an integer.
func parsePercent(s string) int {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	percent, _ := strconv.Atoi(s)
	return percent
}
