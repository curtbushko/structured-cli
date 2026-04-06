package kubectl_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/kubectl"
)

func TestTopPodsParser_ParseSinglePod(t *testing.T) {
	input := `NAME                     CPU(cores)   MEMORY(bytes)
nginx-deployment-abc123   100m         256Mi`

	parser := kubectl.NewTopPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.TopPodsResult)
	require.True(t, ok, "result.Data should be *kubectl.TopPodsResult")

	require.Len(t, podsResult.Pods, 1)

	pod := podsResult.Pods[0]
	assert.Equal(t, "nginx-deployment-abc123", pod.Name)
	assert.Equal(t, "100m", pod.CPU)
	assert.Equal(t, 0.1, pod.CPUCores)
	assert.Equal(t, "256Mi", pod.Memory)
	assert.Equal(t, int64(268435456), pod.MemoryBytes)
}

func TestTopPodsParser_ParseMultiplePods(t *testing.T) {
	input := `NAME                     CPU(cores)   MEMORY(bytes)
nginx-deployment-abc123   100m         256Mi
redis-master-xyz789       500m         1Gi
worker-def456             2            512Mi`

	parser := kubectl.NewTopPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.TopPodsResult)
	require.True(t, ok, "result.Data should be *kubectl.TopPodsResult")

	require.Len(t, podsResult.Pods, 3)

	// First pod - millicores
	assert.Equal(t, "nginx-deployment-abc123", podsResult.Pods[0].Name)
	assert.Equal(t, "100m", podsResult.Pods[0].CPU)
	assert.Equal(t, 0.1, podsResult.Pods[0].CPUCores)
	assert.Equal(t, "256Mi", podsResult.Pods[0].Memory)
	assert.Equal(t, int64(268435456), podsResult.Pods[0].MemoryBytes)

	// Second pod - larger values
	assert.Equal(t, "redis-master-xyz789", podsResult.Pods[1].Name)
	assert.Equal(t, "500m", podsResult.Pods[1].CPU)
	assert.Equal(t, 0.5, podsResult.Pods[1].CPUCores)
	assert.Equal(t, "1Gi", podsResult.Pods[1].Memory)
	assert.Equal(t, int64(1073741824), podsResult.Pods[1].MemoryBytes)

	// Third pod - whole cores
	assert.Equal(t, "worker-def456", podsResult.Pods[2].Name)
	assert.Equal(t, "2", podsResult.Pods[2].CPU)
	assert.Equal(t, 2.0, podsResult.Pods[2].CPUCores)
}

func TestTopPodsParser_ParseWithNamespace(t *testing.T) {
	input := `NAMESPACE     NAME                     CPU(cores)   MEMORY(bytes)
default       nginx-deployment-abc123   100m         256Mi
kube-system   coredns-xyz789            50m          64Mi`

	parser := kubectl.NewTopPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.TopPodsResult)
	require.True(t, ok, "result.Data should be *kubectl.TopPodsResult")

	require.Len(t, podsResult.Pods, 2)

	assert.Equal(t, "default", podsResult.Pods[0].Namespace)
	assert.Equal(t, "nginx-deployment-abc123", podsResult.Pods[0].Name)

	assert.Equal(t, "kube-system", podsResult.Pods[1].Namespace)
	assert.Equal(t, "coredns-xyz789", podsResult.Pods[1].Name)
}

func TestTopPodsParser_ParseContainerMetrics(t *testing.T) {
	input := `POD                       NAME        CPU(cores)   MEMORY(bytes)
nginx-deployment-abc123   nginx       100m         256Mi
nginx-deployment-abc123   sidecar     50m          64Mi
redis-master-xyz789       redis       500m         1Gi`

	parser := kubectl.NewTopPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.TopPodsResult)
	require.True(t, ok, "result.Data should be *kubectl.TopPodsResult")

	require.Len(t, podsResult.Pods, 2)

	// First pod with 2 containers
	nginx := podsResult.Pods[0]
	assert.Equal(t, "nginx-deployment-abc123", nginx.Name)
	require.Len(t, nginx.Containers, 2)
	assert.Equal(t, "nginx", nginx.Containers[0].Name)
	assert.Equal(t, "100m", nginx.Containers[0].CPU)
	assert.Equal(t, 0.1, nginx.Containers[0].CPUCores)
	assert.Equal(t, "sidecar", nginx.Containers[1].Name)
	assert.Equal(t, "50m", nginx.Containers[1].CPU)

	// Second pod with 1 container
	redis := podsResult.Pods[1]
	assert.Equal(t, "redis-master-xyz789", redis.Name)
	require.Len(t, redis.Containers, 1)
	assert.Equal(t, "redis", redis.Containers[0].Name)
}

func TestTopPodsParser_HandleDifferentResourceFormats(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedCPU float64
		expectedMem int64
	}{
		{
			name: "millicores and MiB",
			input: `NAME   CPU(cores)   MEMORY(bytes)
pod-1  100m         256Mi`,
			expectedCPU: 0.1,
			expectedMem: 268435456,
		},
		{
			name: "whole cores and GiB",
			input: `NAME   CPU(cores)   MEMORY(bytes)
pod-1  2            1Gi`,
			expectedCPU: 2.0,
			expectedMem: 1073741824,
		},
		{
			name: "nanocores display and KiB",
			input: `NAME   CPU(cores)   MEMORY(bytes)
pod-1  1500m        512Ki`,
			expectedCPU: 1.5,
			expectedMem: 524288,
		},
		{
			name: "zero values",
			input: `NAME   CPU(cores)   MEMORY(bytes)
pod-1  0m           0Mi`,
			expectedCPU: 0.0,
			expectedMem: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := kubectl.NewTopPodsParser()
			result, err := parser.Parse(strings.NewReader(tt.input))

			require.NoError(t, err)
			require.Nil(t, result.Error)

			podsResult, ok := result.Data.(*kubectl.TopPodsResult)
			require.True(t, ok)

			require.Len(t, podsResult.Pods, 1)
			assert.Equal(t, tt.expectedCPU, podsResult.Pods[0].CPUCores)
			assert.Equal(t, tt.expectedMem, podsResult.Pods[0].MemoryBytes)
		})
	}
}

func TestTopPodsParser_Matches(t *testing.T) {
	parser := kubectl.NewTopPodsParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"kubectl top pods", "kubectl", []string{"top", "pods"}, true},
		{"kubectl top pod", "kubectl", []string{"top", "pod"}, true},
		{"kubectl top pods with flags", "kubectl", []string{"top", "pods", "-n", "default"}, true},
		{"kubectl top pods --containers", "kubectl", []string{"top", "pods", "--containers"}, true},
		{"kubectl top nodes", "kubectl", []string{"top", "nodes"}, false},
		{"kubectl top node", "kubectl", []string{"top", "node"}, false},
		{"kubectl get pods", "kubectl", []string{"get", "pods"}, false},
		{"kubectl only", "kubectl", []string{}, false},
		{"kubectl top only", "kubectl", []string{"top"}, false},
		{"docker top pods", "docker", []string{"top", "pods"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTopPodsParser_Schema(t *testing.T) {
	parser := kubectl.NewTopPodsParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "pods")
}

func TestTopPodsParser_EmptyOutput(t *testing.T) {
	input := `NAME   CPU(cores)   MEMORY(bytes)`

	parser := kubectl.NewTopPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.TopPodsResult)
	require.True(t, ok)

	assert.Empty(t, podsResult.Pods)
}

func TestTopPodsParser_NoResourcesFound(t *testing.T) {
	input := `error: Metrics API not available`

	parser := kubectl.NewTopPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.TopPodsResult)
	require.True(t, ok)

	assert.Empty(t, podsResult.Pods)
}
