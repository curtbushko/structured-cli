package kubectl_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/kubectl"
)

func TestGetPodsParser_ParseSinglePod(t *testing.T) {
	input := `NAME                     READY   STATUS    RESTARTS   AGE
nginx-deployment-abc123   1/1     Running   0          5d`

	parser := kubectl.NewGetPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.GetPodsResult)
	require.True(t, ok, "result.Data should be *kubectl.GetPodsResult")

	require.Len(t, podsResult.Pods, 1)

	pod := podsResult.Pods[0]
	assert.Equal(t, "nginx-deployment-abc123", pod.Name)
	assert.Equal(t, "1/1", pod.Ready)
	assert.Equal(t, "Running", pod.Status)
	assert.Equal(t, 0, pod.Restarts)
	assert.Equal(t, "5d", pod.Age)
}

func TestGetPodsParser_ParseMultiplePods(t *testing.T) {
	input := `NAME                      READY   STATUS             RESTARTS      AGE
nginx-deployment-abc123    1/1     Running            0             5d
redis-master-xyz789        1/1     Running            3             10d
broken-pod-def456          0/1     CrashLoopBackOff   10            1h`

	parser := kubectl.NewGetPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.GetPodsResult)
	require.True(t, ok, "result.Data should be *kubectl.GetPodsResult")

	require.Len(t, podsResult.Pods, 3)

	// First pod
	assert.Equal(t, "nginx-deployment-abc123", podsResult.Pods[0].Name)
	assert.Equal(t, "Running", podsResult.Pods[0].Status)
	assert.Equal(t, 0, podsResult.Pods[0].Restarts)

	// Second pod
	assert.Equal(t, "redis-master-xyz789", podsResult.Pods[1].Name)
	assert.Equal(t, 3, podsResult.Pods[1].Restarts)

	// Third pod with error status
	assert.Equal(t, "broken-pod-def456", podsResult.Pods[2].Name)
	assert.Equal(t, "CrashLoopBackOff", podsResult.Pods[2].Status)
	assert.Equal(t, 10, podsResult.Pods[2].Restarts)
}

func TestGetPodsParser_ParseWideOutput(t *testing.T) {
	input := `NAME                     READY   STATUS    RESTARTS   AGE   IP           NODE
nginx-deployment-abc123   1/1     Running   0          5d    10.244.0.5   worker-1`

	parser := kubectl.NewGetPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.GetPodsResult)
	require.True(t, ok, "result.Data should be *kubectl.GetPodsResult")

	require.Len(t, podsResult.Pods, 1)

	pod := podsResult.Pods[0]
	assert.Equal(t, "nginx-deployment-abc123", pod.Name)
	assert.Equal(t, "10.244.0.5", pod.IP)
	assert.Equal(t, "worker-1", pod.Node)
}

func TestGetPodsParser_ParseWithNamespace(t *testing.T) {
	input := `NAMESPACE     NAME                     READY   STATUS    RESTARTS   AGE
default       nginx-deployment-abc123   1/1     Running   0          5d
kube-system   coredns-xyz789            1/1     Running   2          30d`

	parser := kubectl.NewGetPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.GetPodsResult)
	require.True(t, ok, "result.Data should be *kubectl.GetPodsResult")

	require.Len(t, podsResult.Pods, 2)

	assert.Equal(t, "default", podsResult.Pods[0].Namespace)
	assert.Equal(t, "nginx-deployment-abc123", podsResult.Pods[0].Name)

	assert.Equal(t, "kube-system", podsResult.Pods[1].Namespace)
	assert.Equal(t, "coredns-xyz789", podsResult.Pods[1].Name)
}

func TestGetPodsParser_ParseVariousStatuses(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Pending",
			input: `NAME     READY   STATUS    RESTARTS   AGE
pod-1    0/1     Pending   0          1m`,
			expected: "Pending",
		},
		{
			name: "Terminating",
			input: `NAME     READY   STATUS       RESTARTS   AGE
pod-1    1/1     Terminating  0          5d`,
			expected: "Terminating",
		},
		{
			name: "ImagePullBackOff",
			input: `NAME     READY   STATUS             RESTARTS   AGE
pod-1    0/1     ImagePullBackOff   0          10m`,
			expected: "ImagePullBackOff",
		},
		{
			name: "Error",
			input: `NAME     READY   STATUS   RESTARTS   AGE
pod-1    0/1     Error    5          2h`,
			expected: "Error",
		},
		{
			name: "Completed",
			input: `NAME     READY   STATUS      RESTARTS   AGE
pod-1    0/1     Completed   0          1d`,
			expected: "Completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := kubectl.NewGetPodsParser()
			result, err := parser.Parse(strings.NewReader(tt.input))

			require.NoError(t, err)
			require.Nil(t, result.Error)

			podsResult, ok := result.Data.(*kubectl.GetPodsResult)
			require.True(t, ok)

			require.Len(t, podsResult.Pods, 1)
			assert.Equal(t, tt.expected, podsResult.Pods[0].Status)
		})
	}
}

func TestGetPodsParser_ParseRestartsWithAge(t *testing.T) {
	// Newer kubectl versions show restart count with time since last restart
	input := `NAME     READY   STATUS    RESTARTS        AGE
pod-1    1/1     Running   5 (2h ago)      1d`

	parser := kubectl.NewGetPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.GetPodsResult)
	require.True(t, ok)

	require.Len(t, podsResult.Pods, 1)
	assert.Equal(t, 5, podsResult.Pods[0].Restarts)
}

func TestGetPodsParser_Matches(t *testing.T) {
	parser := kubectl.NewGetPodsParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"kubectl get pods", "kubectl", []string{"get", "pods"}, true},
		{"kubectl get pod", "kubectl", []string{"get", "pod"}, true},
		{"kubectl get po", "kubectl", []string{"get", "po"}, true},
		{"kubectl get pods with flags", "kubectl", []string{"get", "pods", "-n", "default"}, true},
		{"kubectl get pods -o wide", "kubectl", []string{"get", "pods", "-o", "wide"}, true},
		{"kubectl get services", "kubectl", []string{"get", "services"}, false},
		{"kubectl get deployments", "kubectl", []string{"get", "deployments"}, false},
		{"kubectl only", "kubectl", []string{}, false},
		{"kubectl get only", "kubectl", []string{"get"}, false},
		{"docker get pods", "docker", []string{"get", "pods"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPodsParser_Schema(t *testing.T) {
	parser := kubectl.NewGetPodsParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "pods")
}

func TestGetPodsParser_EmptyOutput(t *testing.T) {
	input := `NAME   READY   STATUS   RESTARTS   AGE`

	parser := kubectl.NewGetPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.GetPodsResult)
	require.True(t, ok)

	assert.Empty(t, podsResult.Pods)
}

func TestGetPodsParser_NoResourcesFound(t *testing.T) {
	input := `No resources found in default namespace.`

	parser := kubectl.NewGetPodsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podsResult, ok := result.Data.(*kubectl.GetPodsResult)
	require.True(t, ok)

	assert.Empty(t, podsResult.Pods)
}
