package kubectl_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/kubectl"
)

func TestGetDeploymentsParser_ParseSingleDeployment(t *testing.T) {
	input := `NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           5d`

	parser := kubectl.NewGetDeploymentsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	deploymentsResult, ok := result.Data.(*kubectl.GetDeploymentsResult)
	require.True(t, ok, "result.Data should be *kubectl.GetDeploymentsResult")

	require.Len(t, deploymentsResult.Deployments, 1)

	deployment := deploymentsResult.Deployments[0]
	assert.Equal(t, "nginx-deployment", deployment.Name)
	assert.Equal(t, "3/3", deployment.Ready)
	assert.Equal(t, 3, deployment.ReadyCount)
	assert.Equal(t, 3, deployment.DesiredCount)
	assert.Equal(t, 3, deployment.UpToDate)
	assert.Equal(t, 3, deployment.Available)
	assert.Equal(t, "5d", deployment.Age)
}

func TestGetDeploymentsParser_ParseMultipleDeployments(t *testing.T) {
	input := `NAME               READY   UP-TO-DATE   AVAILABLE   AGE
nginx-deployment   3/3     3            3           5d
redis-master       1/1     1            1           10d
broken-deploy      0/2     2            0           1h`

	parser := kubectl.NewGetDeploymentsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	deploymentsResult, ok := result.Data.(*kubectl.GetDeploymentsResult)
	require.True(t, ok, "result.Data should be *kubectl.GetDeploymentsResult")

	require.Len(t, deploymentsResult.Deployments, 3)

	// First deployment
	assert.Equal(t, "nginx-deployment", deploymentsResult.Deployments[0].Name)
	assert.Equal(t, 3, deploymentsResult.Deployments[0].ReadyCount)
	assert.Equal(t, 3, deploymentsResult.Deployments[0].DesiredCount)

	// Second deployment
	assert.Equal(t, "redis-master", deploymentsResult.Deployments[1].Name)
	assert.Equal(t, 1, deploymentsResult.Deployments[1].ReadyCount)

	// Third deployment with issues
	assert.Equal(t, "broken-deploy", deploymentsResult.Deployments[2].Name)
	assert.Equal(t, 0, deploymentsResult.Deployments[2].ReadyCount)
	assert.Equal(t, 2, deploymentsResult.Deployments[2].DesiredCount)
	assert.Equal(t, 0, deploymentsResult.Deployments[2].Available)
}

func TestGetDeploymentsParser_ParseWithNamespace(t *testing.T) {
	input := `NAMESPACE     NAME               READY   UP-TO-DATE   AVAILABLE   AGE
default       nginx-deployment   3/3     3            3           5d
kube-system   coredns            2/2     2            2           30d`

	parser := kubectl.NewGetDeploymentsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	deploymentsResult, ok := result.Data.(*kubectl.GetDeploymentsResult)
	require.True(t, ok, "result.Data should be *kubectl.GetDeploymentsResult")

	require.Len(t, deploymentsResult.Deployments, 2)

	assert.Equal(t, "default", deploymentsResult.Deployments[0].Namespace)
	assert.Equal(t, "nginx-deployment", deploymentsResult.Deployments[0].Name)

	assert.Equal(t, "kube-system", deploymentsResult.Deployments[1].Namespace)
	assert.Equal(t, "coredns", deploymentsResult.Deployments[1].Name)
}

func TestGetDeploymentsParser_ParseDifferentReadyStates(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedReady   int
		expectedDesired int
	}{
		{
			name: "All replicas ready",
			input: `NAME     READY   UP-TO-DATE   AVAILABLE   AGE
deploy1   5/5     5            5           1d`,
			expectedReady:   5,
			expectedDesired: 5,
		},
		{
			name: "No replicas ready",
			input: `NAME     READY   UP-TO-DATE   AVAILABLE   AGE
deploy1   0/3     3            0           1h`,
			expectedReady:   0,
			expectedDesired: 3,
		},
		{
			name: "Partial replicas ready",
			input: `NAME     READY   UP-TO-DATE   AVAILABLE   AGE
deploy1   2/4     4            2           30m`,
			expectedReady:   2,
			expectedDesired: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := kubectl.NewGetDeploymentsParser()
			result, err := parser.Parse(strings.NewReader(tt.input))

			require.NoError(t, err)
			require.Nil(t, result.Error)

			deploymentsResult, ok := result.Data.(*kubectl.GetDeploymentsResult)
			require.True(t, ok)

			require.Len(t, deploymentsResult.Deployments, 1)
			assert.Equal(t, tt.expectedReady, deploymentsResult.Deployments[0].ReadyCount)
			assert.Equal(t, tt.expectedDesired, deploymentsResult.Deployments[0].DesiredCount)
		})
	}
}

func TestGetDeploymentsParser_Matches(t *testing.T) {
	parser := kubectl.NewGetDeploymentsParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"kubectl get deployments", "kubectl", []string{"get", "deployments"}, true},
		{"kubectl get deployment", "kubectl", []string{"get", "deployment"}, true},
		{"kubectl get deploy", "kubectl", []string{"get", "deploy"}, true},
		{"kubectl get deployments with flags", "kubectl", []string{"get", "deployments", "-n", "default"}, true},
		{"kubectl get deployments -o wide", "kubectl", []string{"get", "deployments", "-o", "wide"}, true},
		{"kubectl get pods", "kubectl", []string{"get", "pods"}, false},
		{"kubectl get services", "kubectl", []string{"get", "services"}, false},
		{"kubectl get nodes", "kubectl", []string{"get", "nodes"}, false},
		{"kubectl only", "kubectl", []string{}, false},
		{"kubectl get only", "kubectl", []string{"get"}, false},
		{"docker get deployments", "docker", []string{"get", "deployments"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDeploymentsParser_Schema(t *testing.T) {
	parser := kubectl.NewGetDeploymentsParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "deployments")
}

func TestGetDeploymentsParser_EmptyOutput(t *testing.T) {
	input := `NAME   READY   UP-TO-DATE   AVAILABLE   AGE`

	parser := kubectl.NewGetDeploymentsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	deploymentsResult, ok := result.Data.(*kubectl.GetDeploymentsResult)
	require.True(t, ok)

	assert.Empty(t, deploymentsResult.Deployments)
}

func TestGetDeploymentsParser_NoResourcesFound(t *testing.T) {
	input := `No resources found in default namespace.`

	parser := kubectl.NewGetDeploymentsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	deploymentsResult, ok := result.Data.(*kubectl.GetDeploymentsResult)
	require.True(t, ok)

	assert.Empty(t, deploymentsResult.Deployments)
}
