package kubectl

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Section names for parsing describe pod output.
const (
	sectionMain       = "main"
	sectionContainers = "containers"
	sectionConditions = "conditions"
	sectionEvents     = "events"
	sectionOther      = "other"
	sectionEventsHdr  = "events-header"
)

// Regular expressions for parsing kubectl describe pod output.
var (
	// Key-value line pattern: "Key:  Value" or "Key:         Value"
	keyValuePattern = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9 ]*?):\s+(.*)$`)

	// Container name pattern: "  containername:"
	containerNamePattern = regexp.MustCompile(`^  ([a-zA-Z0-9][-a-zA-Z0-9_.]*):\s*$`)

	// Indented key-value pattern: "    Key:  Value"
	indentedKeyValuePattern = regexp.MustCompile(`^\s{4}([A-Za-z][A-Za-z0-9 ]*?):\s+(.*)$`)

	// Resource limit/request pattern: "      cpu:     500m"
	resourcePattern = regexp.MustCompile(`^\s{6}([a-z]+):\s+(.+)$`)

	// Condition line pattern: "  Type   Status"
	conditionPattern = regexp.MustCompile(`^\s{2}(\S+)\s+(\S+)(?:\s|$)`)

	// Event line pattern: "  Type    Reason    Age   From   Message"
	eventPattern = regexp.MustCompile(`^\s{2}(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(.*)$`)

	// Node pattern to extract node name from "node-name/ip-address"
	nodePattern = regexp.MustCompile(`^([^/]+)(?:/.*)?$`)

	// Label continuation pattern: "              key=value"
	labelContinuationPattern = regexp.MustCompile(`^\s{16,}([a-zA-Z0-9._/-]+)=(.*)$`)
)

// DescribePodResult represents the structured output of 'kubectl describe pod'.
type DescribePodResult struct {
	// Name is the pod name.
	Name string `json:"name"`

	// Namespace is the pod's namespace.
	Namespace string `json:"namespace"`

	// Node is the node the pod is running on.
	Node string `json:"node"`

	// StartTime is when the pod started.
	StartTime string `json:"start_time"`

	// Labels are the pod's labels.
	Labels map[string]string `json:"labels"`

	// Status is the pod status (e.g., Running, Pending).
	Status string `json:"status"`

	// IP is the pod IP address.
	IP string `json:"ip"`

	// Containers is the list of containers in the pod.
	Containers []Container `json:"containers"`

	// Conditions is the list of pod conditions.
	Conditions []Condition `json:"conditions"`

	// Events is the list of events related to the pod.
	Events []Event `json:"events"`
}

// Container represents a container in a pod.
type Container struct {
	// Name is the container name.
	Name string `json:"name"`

	// Image is the container image.
	Image string `json:"image"`

	// State is the container state (e.g., Running, Waiting, Terminated).
	State string `json:"state"`

	// Ready indicates if the container is ready.
	Ready bool `json:"ready"`

	// RestartCount is the number of times the container has been restarted.
	RestartCount int `json:"restart_count"`

	// Limits are the resource limits for the container.
	Limits map[string]string `json:"limits,omitempty"`

	// Requests are the resource requests for the container.
	Requests map[string]string `json:"requests,omitempty"`
}

// Condition represents a pod condition.
type Condition struct {
	// Type is the condition type (e.g., Ready, Initialized).
	Type string `json:"type"`

	// Status is the condition status (True, False, Unknown).
	Status string `json:"status"`
}

// Event represents an event related to the pod.
type Event struct {
	// Type is the event type (Normal, Warning).
	Type string `json:"type"`

	// Reason is the event reason.
	Reason string `json:"reason"`

	// Age is the event age.
	Age string `json:"age"`

	// From is the source of the event.
	From string `json:"from"`

	// Message is the event message.
	Message string `json:"message"`
}

// DescribePodParser parses the output of 'kubectl describe pod'.
type DescribePodParser struct {
	schema domain.Schema
}

// NewDescribePodParser creates a new DescribePodParser.
func NewDescribePodParser() *DescribePodParser {
	return &DescribePodParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/kubectl-describe-pod.json",
			"Kubectl Describe Pod Output",
			"object",
			map[string]domain.PropertySchema{
				"name":       {Type: "string", Description: "Pod name"},
				"namespace":  {Type: "string", Description: "Pod namespace"},
				"node":       {Type: "string", Description: "Node the pod is running on"},
				"start_time": {Type: "string", Description: "Pod start time"},
				"labels":     {Type: "object", Description: "Pod labels"},
				"status":     {Type: "string", Description: "Pod status"},
				"ip":         {Type: "string", Description: "Pod IP address"},
				"containers": {Type: "array", Description: "List of containers"},
				"conditions": {Type: "array", Description: "List of conditions"},
				"events":     {Type: "array", Description: "List of events"},
			},
			[]string{"name", "namespace", "status"},
		),
	}
}

// Parse reads kubectl describe pod output and returns structured data.
func (p *DescribePodParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)
	result := p.parseDescribeOutput(raw)

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for kubectl describe pod output.
func (p *DescribePodParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *DescribePodParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdKubectl || len(subcommands) < 2 || subcommands[0] != "describe" {
		return false
	}
	// Match "pod", "pods", or "po" (the common aliases)
	resource := subcommands[1]
	return resource == "pod" || resource == "pods" || resource == "po"
}

// parseDescribeOutput parses the describe pod output into structured data.
func (p *DescribePodParser) parseDescribeOutput(raw string) *DescribePodResult {
	result := &DescribePodResult{
		Labels:     make(map[string]string),
		Containers: []Container{},
		Conditions: []Condition{},
		Events:     []Event{},
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	var section string
	var currentContainer *Container
	var inLimits, inRequests bool

	for scanner.Scan() {
		line := scanner.Text()

		// Detect section changes
		prevSection := section
		section = p.detectSection(line, section)

		// If we're leaving the containers section, save the last container
		if prevSection == sectionContainers && section != sectionContainers && currentContainer != nil {
			result.Containers = append(result.Containers, *currentContainer)
			currentContainer = nil
		}

		switch section {
		case sectionMain:
			p.parseMainSection(line, result)
		case sectionContainers:
			currentContainer, inLimits, inRequests = p.parseContainersSection(
				line, result, currentContainer, inLimits, inRequests,
			)
		case sectionConditions:
			p.parseConditionsSection(line, result)
		case sectionEvents:
			p.parseEventsSection(line, result)
		}
	}

	// Append last container if we ended in containers section
	if currentContainer != nil {
		result.Containers = append(result.Containers, *currentContainer)
	}

	return result
}

// detectSection determines which section of the describe output we're in.
func (p *DescribePodParser) detectSection(line, currentSection string) string {
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(line, "Containers:") {
		return sectionContainers
	}
	if strings.HasPrefix(line, "Conditions:") {
		return sectionConditions
	}
	if strings.HasPrefix(line, "Events:") {
		return sectionEvents
	}

	// Check for Init Containers or other container sections that end container parsing
	if strings.HasPrefix(line, "Init Containers:") ||
		strings.HasPrefix(line, "Volumes:") ||
		strings.HasPrefix(line, "QoS Class:") {
		return sectionOther
	}

	// If we haven't seen a section header yet, we're in the main section
	if currentSection == "" {
		return sectionMain
	}

	// Skip headers in events section
	if currentSection == sectionEvents && strings.HasPrefix(trimmed, "Type") && strings.Contains(line, "Reason") {
		return sectionEventsHdr
	}
	if currentSection == sectionEventsHdr && strings.HasPrefix(trimmed, "----") {
		return sectionEvents
	}

	return currentSection
}

// parseMainSection parses the main key-value section at the top.
func (p *DescribePodParser) parseMainSection(line string, result *DescribePodResult) {
	matches := keyValuePattern.FindStringSubmatch(line)
	if matches == nil {
		// Check for label continuation
		if labelMatches := labelContinuationPattern.FindStringSubmatch(line); labelMatches != nil {
			result.Labels[labelMatches[1]] = labelMatches[2]
		}
		return
	}

	key := matches[1]
	value := strings.TrimSpace(matches[2])

	switch key {
	case "Name":
		result.Name = value
	case "Namespace":
		result.Namespace = value
	case "Node":
		// Extract node name from "node-name/ip-address"
		if nodeMatches := nodePattern.FindStringSubmatch(value); nodeMatches != nil {
			result.Node = nodeMatches[1]
		}
	case "Start Time":
		result.StartTime = value
	case "Labels":
		p.parseLabels(value, result)
	case "Status":
		result.Status = value
	case "IP":
		if value != noneValue {
			result.IP = value
		}
	}
}

// parseLabels parses the labels from the first label line.
func (p *DescribePodParser) parseLabels(value string, result *DescribePodResult) {
	if value == noneValue {
		return
	}

	// Parse "key=value" format
	parts := strings.SplitN(value, "=", 2)
	if len(parts) == 2 {
		result.Labels[parts[0]] = parts[1]
	}
}

// parseContainersSection parses the Containers section.
func (p *DescribePodParser) parseContainersSection(
	line string,
	result *DescribePodResult,
	currentContainer *Container,
	inLimits, inRequests bool,
) (*Container, bool, bool) {
	// Check for new container name
	if containerMatches := containerNamePattern.FindStringSubmatch(line); containerMatches != nil {
		return p.startNewContainer(containerMatches[1], result, currentContainer)
	}

	if currentContainer == nil {
		return currentContainer, inLimits, inRequests
	}

	// Check for Limits/Requests section
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "Limits:") {
		return currentContainer, true, false
	}
	if strings.HasPrefix(trimmed, "Requests:") {
		return currentContainer, false, true
	}

	// Parse resource limits/requests
	if inLimits || inRequests {
		return p.parseResourceLine(line, trimmed, currentContainer, inLimits, inRequests)
	}

	// Parse container key-value pairs
	p.parseContainerKeyValue(line, currentContainer)

	return currentContainer, inLimits, inRequests
}

// startNewContainer saves the previous container and starts a new one.
func (p *DescribePodParser) startNewContainer(
	name string,
	result *DescribePodResult,
	currentContainer *Container,
) (*Container, bool, bool) {
	if currentContainer != nil {
		result.Containers = append(result.Containers, *currentContainer)
	}
	newContainer := &Container{
		Name:     name,
		Limits:   make(map[string]string),
		Requests: make(map[string]string),
	}
	return newContainer, false, false
}

// parseResourceLine parses a resource limit/request line.
func (p *DescribePodParser) parseResourceLine(
	line, trimmed string,
	container *Container,
	inLimits, inRequests bool,
) (*Container, bool, bool) {
	if resourceMatches := resourcePattern.FindStringSubmatch(line); resourceMatches != nil {
		if inLimits {
			container.Limits[resourceMatches[1]] = resourceMatches[2]
		} else {
			container.Requests[resourceMatches[1]] = resourceMatches[2]
		}
		return container, inLimits, inRequests
	}

	// Check if we've left the resource section
	leftSection := (len(line) > 0 && line[0] != ' ') ||
		(len(trimmed) > 0 && !strings.HasPrefix(line, "      "))
	if leftSection {
		return container, false, false
	}

	return container, inLimits, inRequests
}

// parseContainerKeyValue parses container key-value pairs.
func (p *DescribePodParser) parseContainerKeyValue(line string, container *Container) {
	matches := indentedKeyValuePattern.FindStringSubmatch(line)
	if matches == nil {
		return
	}

	key := matches[1]
	value := strings.TrimSpace(matches[2])

	switch key {
	case "Image":
		container.Image = value
	case "State":
		container.State = value
	case "Ready":
		container.Ready = strings.ToLower(value) == "true"
	case "Restart Count":
		count, _ := strconv.Atoi(value)
		container.RestartCount = count
	}
}

// parseConditionsSection parses the Conditions section.
func (p *DescribePodParser) parseConditionsSection(line string, result *DescribePodResult) {
	trimmed := strings.TrimSpace(line)

	// Skip headers
	if trimmed == "" || strings.HasPrefix(trimmed, "Type") || strings.HasPrefix(trimmed, "----") {
		return
	}

	if matches := conditionPattern.FindStringSubmatch(line); matches != nil {
		// Skip if this looks like a header
		if matches[1] == "Type" {
			return
		}
		result.Conditions = append(result.Conditions, Condition{
			Type:   matches[1],
			Status: matches[2],
		})
	}
}

// parseEventsSection parses the Events section.
func (p *DescribePodParser) parseEventsSection(line string, result *DescribePodResult) {
	trimmed := strings.TrimSpace(line)

	// Handle <none> case
	if trimmed == noneValue || strings.HasSuffix(trimmed, noneValue) {
		return
	}

	// Skip headers and separators
	if trimmed == "" ||
		strings.HasPrefix(trimmed, "Type") ||
		strings.HasPrefix(trimmed, "----") {
		return
	}

	if matches := eventPattern.FindStringSubmatch(line); matches != nil {
		// Skip if this looks like a header
		if matches[1] == "Type" {
			return
		}
		result.Events = append(result.Events, Event{
			Type:    matches[1],
			Reason:  matches[2],
			Age:     matches[3],
			From:    matches[4],
			Message: strings.TrimSpace(matches[5]),
		})
	}
}
