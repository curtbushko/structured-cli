package kubectl_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/kubectl"
)

func TestGetNodesParser_ParseSingleNode(t *testing.T) {
	input := `NAME       STATUS   ROLES           AGE   VERSION
worker-1   Ready    <none>          10d   v1.28.0`

	parser := kubectl.NewGetNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.GetNodesResult)
	require.True(t, ok, "result.Data should be *kubectl.GetNodesResult")

	require.Len(t, nodesResult.Nodes, 1)

	node := nodesResult.Nodes[0]
	assert.Equal(t, "worker-1", node.Name)
	assert.Equal(t, "Ready", node.Status)
	assert.Empty(t, node.Roles) // <none> should result in empty roles
	assert.Equal(t, "10d", node.Age)
	assert.Equal(t, "v1.28.0", node.Version)
}

func TestGetNodesParser_ParseMultipleNodes(t *testing.T) {
	input := `NAME          STATUS   ROLES           AGE   VERSION
control-1     Ready    control-plane   30d   v1.28.0
worker-1      Ready    <none>          10d   v1.28.0
worker-2      Ready    <none>          10d   v1.28.0`

	parser := kubectl.NewGetNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.GetNodesResult)
	require.True(t, ok, "result.Data should be *kubectl.GetNodesResult")

	require.Len(t, nodesResult.Nodes, 3)

	// Control plane node
	assert.Equal(t, "control-1", nodesResult.Nodes[0].Name)
	assert.Equal(t, []string{"control-plane"}, nodesResult.Nodes[0].Roles)

	// Worker nodes
	assert.Equal(t, "worker-1", nodesResult.Nodes[1].Name)
	assert.Empty(t, nodesResult.Nodes[1].Roles)

	assert.Equal(t, "worker-2", nodesResult.Nodes[2].Name)
	assert.Empty(t, nodesResult.Nodes[2].Roles)
}

func TestGetNodesParser_ParseWideOutput(t *testing.T) {
	input := `NAME       STATUS   ROLES           AGE   VERSION   INTERNAL-IP    EXTERNAL-IP   OS-IMAGE             KERNEL-VERSION      CONTAINER-RUNTIME
worker-1   Ready    <none>          10d   v1.28.0   192.168.1.10   <none>        Ubuntu 22.04.3 LTS   5.15.0-91-generic   containerd://1.7.2`

	parser := kubectl.NewGetNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.GetNodesResult)
	require.True(t, ok, "result.Data should be *kubectl.GetNodesResult")

	require.Len(t, nodesResult.Nodes, 1)

	node := nodesResult.Nodes[0]
	assert.Equal(t, "worker-1", node.Name)
	assert.Equal(t, "Ready", node.Status)
	assert.Equal(t, "v1.28.0", node.Version)
	assert.Equal(t, "192.168.1.10", node.InternalIP)
	assert.Empty(t, node.ExternalIP) // <none> should be empty
	assert.Equal(t, "Ubuntu 22.04.3 LTS", node.OSImage)
	assert.Equal(t, "5.15.0-91-generic", node.KernelVersion)
	assert.Equal(t, "containerd://1.7.2", node.ContainerRuntime)
}

func TestGetNodesParser_ParseMultipleRoles(t *testing.T) {
	input := `NAME          STATUS   ROLES                  AGE   VERSION
control-1     Ready    control-plane,master   30d   v1.28.0`

	parser := kubectl.NewGetNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.GetNodesResult)
	require.True(t, ok, "result.Data should be *kubectl.GetNodesResult")

	require.Len(t, nodesResult.Nodes, 1)

	node := nodesResult.Nodes[0]
	assert.Equal(t, "control-1", node.Name)
	assert.Equal(t, []string{"control-plane", "master"}, node.Roles)
}

func TestGetNodesParser_ParseNoRoles(t *testing.T) {
	input := `NAME       STATUS   ROLES    AGE   VERSION
worker-1   Ready    <none>   10d   v1.28.0`

	parser := kubectl.NewGetNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.GetNodesResult)
	require.True(t, ok, "result.Data should be *kubectl.GetNodesResult")

	require.Len(t, nodesResult.Nodes, 1)
	assert.Empty(t, nodesResult.Nodes[0].Roles)
}

func TestGetNodesParser_ParseVariousStatuses(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Ready",
			input: `NAME     STATUS   ROLES    AGE   VERSION
node-1   Ready    <none>   1d    v1.28.0`,
			expected: "Ready",
		},
		{
			name: "NotReady",
			input: `NAME     STATUS     ROLES    AGE   VERSION
node-1   NotReady   <none>   1d    v1.28.0`,
			expected: "NotReady",
		},
		{
			name: "SchedulingDisabled",
			input: `NAME     STATUS                     ROLES    AGE   VERSION
node-1   Ready,SchedulingDisabled   <none>   1d    v1.28.0`,
			expected: "Ready,SchedulingDisabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := kubectl.NewGetNodesParser()
			result, err := parser.Parse(strings.NewReader(tt.input))

			require.NoError(t, err)
			require.Nil(t, result.Error)

			nodesResult, ok := result.Data.(*kubectl.GetNodesResult)
			require.True(t, ok)

			require.Len(t, nodesResult.Nodes, 1)
			assert.Equal(t, tt.expected, nodesResult.Nodes[0].Status)
		})
	}
}

func TestGetNodesParser_Matches(t *testing.T) {
	parser := kubectl.NewGetNodesParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"kubectl get nodes", "kubectl", []string{"get", "nodes"}, true},
		{"kubectl get node", "kubectl", []string{"get", "node"}, true},
		{"kubectl get no", "kubectl", []string{"get", "no"}, true},
		{"kubectl get nodes with flags", "kubectl", []string{"get", "nodes", "-o", "wide"}, true},
		{"kubectl get pods", "kubectl", []string{"get", "pods"}, false},
		{"kubectl get services", "kubectl", []string{"get", "services"}, false},
		{"kubectl get deployments", "kubectl", []string{"get", "deployments"}, false},
		{"kubectl only", "kubectl", []string{}, false},
		{"kubectl get only", "kubectl", []string{"get"}, false},
		{"docker get nodes", "docker", []string{"get", "nodes"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetNodesParser_Schema(t *testing.T) {
	parser := kubectl.NewGetNodesParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "nodes")
}

func TestGetNodesParser_EmptyOutput(t *testing.T) {
	input := `NAME   STATUS   ROLES   AGE   VERSION`

	parser := kubectl.NewGetNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.GetNodesResult)
	require.True(t, ok)

	assert.Empty(t, nodesResult.Nodes)
}

func TestGetNodesParser_NoResourcesFound(t *testing.T) {
	input := `No resources found.`

	parser := kubectl.NewGetNodesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	nodesResult, ok := result.Data.(*kubectl.GetNodesResult)
	require.True(t, ok)

	assert.Empty(t, nodesResult.Nodes)
}
