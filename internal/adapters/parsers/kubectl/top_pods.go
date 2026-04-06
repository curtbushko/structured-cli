package kubectl

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ContainerMetrics represents metrics for a single container.
type ContainerMetrics struct {
	// Name is the container name.
	Name string `json:"name"`

	// CPU is the raw CPU usage string (e.g., "100m", "2").
	CPU string `json:"cpu"`

	// CPUCores is the CPU usage in cores (e.g., 0.1, 2.0).
	CPUCores float64 `json:"cpu_cores"`

	// Memory is the raw memory usage string (e.g., "256Mi", "1Gi").
	Memory string `json:"memory"`

	// MemoryBytes is the memory usage in bytes.
	MemoryBytes int64 `json:"memory_bytes"`
}

// PodMetrics represents metrics for a single pod.
type PodMetrics struct {
	// Name is the pod name.
	Name string `json:"name"`

	// Namespace is the pod's namespace (when -A or --all-namespaces is used).
	Namespace string `json:"namespace,omitempty"`

	// CPU is the raw CPU usage string (e.g., "100m", "2").
	CPU string `json:"cpu"`

	// CPUCores is the CPU usage in cores (e.g., 0.1, 2.0).
	CPUCores float64 `json:"cpu_cores"`

	// Memory is the raw memory usage string (e.g., "256Mi", "1Gi").
	Memory string `json:"memory"`

	// MemoryBytes is the memory usage in bytes.
	MemoryBytes int64 `json:"memory_bytes"`

	// Containers contains per-container metrics (when --containers is used).
	Containers []ContainerMetrics `json:"containers,omitempty"`
}

// TopPodsResult represents the structured output of 'kubectl top pods'.
type TopPodsResult struct {
	// Pods is the list of pod metrics.
	Pods []PodMetrics `json:"pods"`
}

// Regular expressions for parsing resource values.
var (
	millicorePattern = regexp.MustCompile(`^(\d+)m$`)
	corePattern      = regexp.MustCompile(`^(\d+)$`)
	memoryPattern    = regexp.MustCompile(`^(\d+)(Ki|Mi|Gi|Ti)?$`)
)

// Top pods column names.
var topPodsColumnNames = []string{"NAMESPACE", "POD", "NAME", "CPU(cores)", "MEMORY(bytes)"}

// Required columns for top pods output validation.
var topPodsRequiredColumns = []string{"CPU(cores)", "MEMORY(bytes)"}

// TopPodsParser parses the output of 'kubectl top pods'.
type TopPodsParser struct {
	schema domain.Schema
}

// NewTopPodsParser creates a new TopPodsParser with the kubectl-top-pods schema.
func NewTopPodsParser() *TopPodsParser {
	return &TopPodsParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/kubectl-top-pods.json",
			"Kubectl Top Pods Output",
			"object",
			map[string]domain.PropertySchema{
				"pods": {Type: "array", Description: "List of pod metrics"},
			},
			[]string{"pods"},
		),
	}
}

// Parse reads kubectl top pods output and returns structured data.
func (p *TopPodsParser) Parse(r io.Reader) (domain.ParseResult, error) {
	prep := readAndPrepare(r, topPodsColumnNames, topPodsRequiredColumns)

	if prep.Error != nil {
		return emptyResultWithError(prep.Error, prep.ErrorMsg), nil
	}

	emptyResult := &TopPodsResult{Pods: []PodMetrics{}}
	if prep.IsEmpty {
		return emptyResultOK(emptyResult, prep.Input.Raw), nil
	}

	// Detect container mode (POD column present means --containers flag used)
	hasContainerMode := hasColumn(prep.Input.Columns, "POD")

	if hasContainerMode {
		pods := parseContainerLines(prep.Input.Scanner, prep.Input.Columns)
		return domain.NewParseResult(&TopPodsResult{Pods: pods}, prep.Input.Raw, 0), nil
	}

	pods := parseLines(prep.Input.Scanner, prep.Input.Columns, parseTopPodLine)
	return domain.NewParseResult(&TopPodsResult{Pods: pods}, prep.Input.Raw, 0), nil
}

// Schema returns the JSON Schema for kubectl top pods output.
func (p *TopPodsParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *TopPodsParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdKubectl || len(subcommands) < 2 || subcommands[0] != "top" {
		return false
	}
	// Match "pods" or "pod"
	resource := subcommands[1]
	return resource == "pods" || resource == "pod"
}

// hasColumn checks if a column exists in the columns list.
func hasColumn(columns []columnInfo, name string) bool {
	for _, col := range columns {
		if col.name == name {
			return true
		}
	}
	return false
}

// parseTopPodLine parses a single line of kubectl top pods output.
func parseTopPodLine(line string, columns []columnInfo) PodMetrics {
	pod := PodMetrics{}

	for _, col := range columns {
		value := extractColumnValue(line, col)
		setTopPodField(&pod, col.name, value)
	}

	return pod
}

// setTopPodField sets a field on the pod based on column name.
func setTopPodField(pod *PodMetrics, colNameVal, value string) {
	switch colNameVal {
	case colNamespace:
		pod.Namespace = value
	case colName:
		pod.Name = value
	case colCPU:
		pod.CPU = value
		pod.CPUCores = parseCPU(value)
	case colMemory:
		pod.Memory = value
		pod.MemoryBytes = parseMemory(value)
	}
}

// parseContainerLines parses lines for container-mode output (--containers flag).
func parseContainerLines(scanner *bufio.Scanner, columns []columnInfo) []PodMetrics {
	podMap := make(map[string]*PodMetrics)
	var podOrder []string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Extract values from the line
		var podName, containerName, cpu, memory, namespace string
		for _, col := range columns {
			value := extractColumnValue(line, col)
			switch col.name {
			case "POD":
				podName = value
			case colName:
				containerName = value
			case colNamespace:
				namespace = value
			case "CPU(cores)":
				cpu = value
			case "MEMORY(bytes)":
				memory = value
			}
		}

		// Add to existing pod or create new one
		if _, exists := podMap[podName]; !exists {
			podMap[podName] = &PodMetrics{
				Name:       podName,
				Namespace:  namespace,
				Containers: []ContainerMetrics{},
			}
			podOrder = append(podOrder, podName)
		}

		container := ContainerMetrics{
			Name:        containerName,
			CPU:         cpu,
			CPUCores:    parseCPU(cpu),
			Memory:      memory,
			MemoryBytes: parseMemory(memory),
		}
		podMap[podName].Containers = append(podMap[podName].Containers, container)
	}

	// Build result slice in order
	result := make([]PodMetrics, 0, len(podOrder))
	for _, name := range podOrder {
		result = append(result, *podMap[name])
	}

	return result
}

// parseCPU parses a CPU value like "100m" or "2" into cores.
func parseCPU(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Check for millicores (e.g., "100m")
	if matches := millicorePattern.FindStringSubmatch(s); len(matches) > 1 {
		millicores, _ := strconv.ParseFloat(matches[1], 64)
		return millicores / 1000.0
	}

	// Check for whole cores (e.g., "2")
	if matches := corePattern.FindStringSubmatch(s); len(matches) > 1 {
		cores, _ := strconv.ParseFloat(matches[1], 64)
		return cores
	}

	return 0
}

// parseMemory parses a memory value like "256Mi" or "1Gi" into bytes.
func parseMemory(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	matches := memoryPattern.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0
	}

	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0
	}

	var multiplier int64 = 1
	if len(matches) > 2 && matches[2] != "" {
		switch matches[2] {
		case "Ki":
			multiplier = 1024
		case "Mi":
			multiplier = 1024 * 1024
		case "Gi":
			multiplier = 1024 * 1024 * 1024
		case "Ti":
			multiplier = 1024 * 1024 * 1024 * 1024
		}
	}

	return value * multiplier
}
