package kubectl_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/kubectl"
)

func TestTopNodesParser_ParseSingleNode(t *testing.T) {
	input := `NAME       CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
worker-1   500m         25%    4096Mi          50%`

	parser := kubectl.NewTopNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.TopNodesResult)
	require.True(t, ok, "result.Data should be *kubectl.TopNodesResult")

	require.Len(t, nodesResult.Nodes, 1)

	node := nodesResult.Nodes[0]
	assert.Equal(t, "worker-1", node.Name)
	assert.Equal(t, "500m", node.CPU)
	assert.Equal(t, 0.5, node.CPUCores)
	assert.Equal(t, 25, node.CPUPercent)
	assert.Equal(t, "4096Mi", node.Memory)
	assert.Equal(t, int64(4294967296), node.MemoryBytes)
	assert.Equal(t, 50, node.MemoryPercent)
}

func TestTopNodesParser_ParseMultipleNodes(t *testing.T) {
	input := `NAME          CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
control-1     1000m        50%    8192Mi          75%
worker-1      500m         25%    4096Mi          50%
worker-2      2            100%   16Gi            90%`

	parser := kubectl.NewTopNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.TopNodesResult)
	require.True(t, ok, "result.Data should be *kubectl.TopNodesResult")

	require.Len(t, nodesResult.Nodes, 3)

	// Control plane node
	assert.Equal(t, "control-1", nodesResult.Nodes[0].Name)
	assert.Equal(t, "1000m", nodesResult.Nodes[0].CPU)
	assert.Equal(t, 1.0, nodesResult.Nodes[0].CPUCores)
	assert.Equal(t, 50, nodesResult.Nodes[0].CPUPercent)
	assert.Equal(t, "8192Mi", nodesResult.Nodes[0].Memory)
	assert.Equal(t, int64(8589934592), nodesResult.Nodes[0].MemoryBytes)
	assert.Equal(t, 75, nodesResult.Nodes[0].MemoryPercent)

	// Worker node 1
	assert.Equal(t, "worker-1", nodesResult.Nodes[1].Name)
	assert.Equal(t, 0.5, nodesResult.Nodes[1].CPUCores)
	assert.Equal(t, 25, nodesResult.Nodes[1].CPUPercent)

	// Worker node 2 - whole cores
	assert.Equal(t, "worker-2", nodesResult.Nodes[2].Name)
	assert.Equal(t, "2", nodesResult.Nodes[2].CPU)
	assert.Equal(t, 2.0, nodesResult.Nodes[2].CPUCores)
	assert.Equal(t, 100, nodesResult.Nodes[2].CPUPercent)
	assert.Equal(t, "16Gi", nodesResult.Nodes[2].Memory)
	assert.Equal(t, int64(17179869184), nodesResult.Nodes[2].MemoryBytes)
	assert.Equal(t, 90, nodesResult.Nodes[2].MemoryPercent)
}

func TestTopNodesParser_HandlePercentageValues(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedCPU    int
		expectedMemory int
	}{
		{
			name: "low percentages",
			input: `NAME   CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1   100m         5%     1024Mi          10%`,
			expectedCPU:    5,
			expectedMemory: 10,
		},
		{
			name: "high percentages",
			input: `NAME   CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1   4            99%    32Gi            95%`,
			expectedCPU:    99,
			expectedMemory: 95,
		},
		{
			name: "zero percentages",
			input: `NAME   CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1   0m           0%     0Mi             0%`,
			expectedCPU:    0,
			expectedMemory: 0,
		},
		{
			name: "100 percent",
			input: `NAME   CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1   8            100%   64Gi            100%`,
			expectedCPU:    100,
			expectedMemory: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := kubectl.NewTopNodesParser()
			result, err := parser.Parse(strings.NewReader(tt.input))

			require.NoError(t, err)
			require.Nil(t, result.Error)

			nodesResult, ok := result.Data.(*kubectl.TopNodesResult)
			require.True(t, ok)

			require.Len(t, nodesResult.Nodes, 1)
			assert.Equal(t, tt.expectedCPU, nodesResult.Nodes[0].CPUPercent)
			assert.Equal(t, tt.expectedMemory, nodesResult.Nodes[0].MemoryPercent)
		})
	}
}

func TestTopNodesParser_HandleVariousResourceFormats(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedCPU float64
		expectedMem int64
	}{
		{
			name: "millicores and MiB",
			input: `NAME   CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1   500m         25%    4096Mi          50%`,
			expectedCPU: 0.5,
			expectedMem: 4294967296,
		},
		{
			name: "whole cores and GiB",
			input: `NAME   CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1   4            50%    16Gi            75%`,
			expectedCPU: 4.0,
			expectedMem: 17179869184,
		},
		{
			name: "decimal cores",
			input: `NAME   CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
node-1   1500m        75%    8192Mi          50%`,
			expectedCPU: 1.5,
			expectedMem: 8589934592,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := kubectl.NewTopNodesParser()
			result, err := parser.Parse(strings.NewReader(tt.input))

			require.NoError(t, err)
			require.Nil(t, result.Error)

			nodesResult, ok := result.Data.(*kubectl.TopNodesResult)
			require.True(t, ok)

			require.Len(t, nodesResult.Nodes, 1)
			assert.Equal(t, tt.expectedCPU, nodesResult.Nodes[0].CPUCores)
			assert.Equal(t, tt.expectedMem, nodesResult.Nodes[0].MemoryBytes)
		})
	}
}

func TestTopNodesParser_Matches(t *testing.T) {
	parser := kubectl.NewTopNodesParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"kubectl top nodes", "kubectl", []string{"top", "nodes"}, true},
		{"kubectl top node", "kubectl", []string{"top", "node"}, true},
		{"kubectl top nodes with flags", "kubectl", []string{"top", "nodes", "--no-headers"}, true},
		{"kubectl top pods", "kubectl", []string{"top", "pods"}, false},
		{"kubectl top pod", "kubectl", []string{"top", "pod"}, false},
		{"kubectl get nodes", "kubectl", []string{"get", "nodes"}, false},
		{"kubectl only", "kubectl", []string{}, false},
		{"kubectl top only", "kubectl", []string{"top"}, false},
		{"docker top nodes", "docker", []string{"top", "nodes"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTopNodesParser_Schema(t *testing.T) {
	parser := kubectl.NewTopNodesParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "nodes")
}

func TestTopNodesParser_EmptyOutput(t *testing.T) {
	input := `NAME   CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%`

	parser := kubectl.NewTopNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.TopNodesResult)
	require.True(t, ok)

	assert.Empty(t, nodesResult.Nodes)
}

func TestTopNodesParser_MetricsAPINotAvailable(t *testing.T) {
	input := `error: Metrics API not available`

	parser := kubectl.NewTopNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.TopNodesResult)
	require.True(t, ok)

	assert.Empty(t, nodesResult.Nodes)
}
