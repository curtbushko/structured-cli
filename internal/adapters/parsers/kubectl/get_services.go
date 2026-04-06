package kubectl

import (
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regular expressions for parsing port strings.
var (
	// portPattern matches port strings like "80/TCP", "80:30080/TCP"
	portPattern = regexp.MustCompile(`^(\d+)(?::(\d+))?/(\w+)$`)
)

// Service column names for kubectl get services.
var serviceColumnNames = []string{"NAMESPACE", "NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)", "AGE"}

// Required columns for service output validation.
var serviceRequiredColumns = []string{"NAME", "TYPE", "CLUSTER-IP"}

// GetServicesParser parses the output of 'kubectl get services'.
type GetServicesParser struct {
	schema domain.Schema
}

// NewGetServicesParser creates a new GetServicesParser with the kubectl-get-services schema.
func NewGetServicesParser() *GetServicesParser {
	return &GetServicesParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/kubectl-get-services.json",
			"Kubectl Get Services Output",
			"object",
			map[string]domain.PropertySchema{
				"services": {Type: "array", Description: "List of services"},
			},
			[]string{"services"},
		),
	}
}

// Parse reads kubectl get services output and returns structured data.
func (p *GetServicesParser) Parse(r io.Reader) (domain.ParseResult, error) {
	prep := readAndPrepare(r, serviceColumnNames, serviceRequiredColumns)

	if prep.Error != nil {
		return emptyResultWithError(prep.Error, prep.ErrorMsg), nil
	}

	emptyResult := &GetServicesResult{Services: []Service{}}
	if prep.IsEmpty {
		return emptyResultOK(emptyResult, prep.Input.Raw), nil
	}

	services := parseLines(prep.Input.Scanner, prep.Input.Columns, parseServiceLine)

	return domain.NewParseResult(&GetServicesResult{Services: services}, prep.Input.Raw, 0), nil
}

// Schema returns the JSON Schema for kubectl get services output.
func (p *GetServicesParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *GetServicesParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdKubectl || len(subcommands) < 2 || subcommands[0] != cmdGet {
		return false
	}
	// Match "services", "service", or "svc" (the common aliases)
	resource := subcommands[1]
	return resource == "services" || resource == "service" || resource == "svc"
}

// parseServiceLine parses a single line of kubectl get services output.
func parseServiceLine(line string, columns []columnInfo) Service {
	svc := Service{Ports: []Port{}}

	for _, col := range columns {
		value := extractColumnValue(line, col)
		setServiceField(&svc, col.name, value)
	}

	return svc
}

// setServiceField sets a field on the service based on column name.
func setServiceField(svc *Service, colNameVal, value string) {
	switch colNameVal {
	case colNamespace:
		svc.Namespace = value
	case colName:
		svc.Name = value
	case "TYPE":
		svc.Type = value
	case "CLUSTER-IP":
		svc.ClusterIP = normalizeKubectlValue(value)
	case "EXTERNAL-IP":
		svc.ExternalIP = normalizeKubectlValue(value)
	case "PORT(S)":
		svc.Ports = parsePorts(value)
	case colAge:
		svc.Age = value
	}
}

// parsePorts parses port strings like "80/TCP", "80:30080/TCP", or "80/TCP,443/TCP".
func parsePorts(s string) []Port {
	if s == noneValue || s == "" {
		return []Port{}
	}

	var ports []Port
	portStrs := strings.Split(s, ",")

	for _, ps := range portStrs {
		if port, ok := parsePort(strings.TrimSpace(ps)); ok {
			ports = append(ports, port)
		}
	}

	return ports
}

// parsePort parses a single port string like "80/TCP" or "80:30080/TCP".
func parsePort(s string) (Port, bool) {
	matches := portPattern.FindStringSubmatch(s)
	if len(matches) < 4 {
		return Port{}, false
	}

	port := Port{Protocol: matches[3]}
	port.Port, _ = strconv.Atoi(matches[1])
	if matches[2] != "" {
		port.NodePort, _ = strconv.Atoi(matches[2])
	}
	return port, true
}
