package kubectl_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/kubectl"
)

const noResourcesFoundMsg = "No resources found in default namespace."

func TestGetServicesParser_ParseSingleService(t *testing.T) {
	input := `NAME         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   10.96.0.1        <none>        443/TCP   30d`

	parser := kubectl.NewGetServicesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	svcResult, ok := result.Data.(*kubectl.GetServicesResult)
	require.True(t, ok, "result.Data should be *kubectl.GetServicesResult")

	require.Len(t, svcResult.Services, 1)

	svc := svcResult.Services[0]
	assert.Equal(t, "kubernetes", svc.Name)
	assert.Equal(t, "ClusterIP", svc.Type)
	assert.Equal(t, "10.96.0.1", svc.ClusterIP)
	assert.Equal(t, "", svc.ExternalIP)
	assert.Equal(t, "30d", svc.Age)

	require.Len(t, svc.Ports, 1)
	assert.Equal(t, 443, svc.Ports[0].Port)
	assert.Equal(t, "TCP", svc.Ports[0].Protocol)
	assert.Equal(t, 0, svc.Ports[0].NodePort)
}

func TestGetServicesParser_ParseMultipleServices(t *testing.T) {
	input := `NAME           TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)        AGE
kubernetes     ClusterIP      10.96.0.1       <none>         443/TCP        30d
nginx-svc      LoadBalancer   10.96.100.50    203.0.113.10   80:30080/TCP   5d
redis-svc      NodePort       10.96.200.75    <none>         6379:31379/TCP 10d`

	parser := kubectl.NewGetServicesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	svcResult, ok := result.Data.(*kubectl.GetServicesResult)
	require.True(t, ok, "result.Data should be *kubectl.GetServicesResult")

	require.Len(t, svcResult.Services, 3)

	// First service - ClusterIP
	assert.Equal(t, "kubernetes", svcResult.Services[0].Name)
	assert.Equal(t, "ClusterIP", svcResult.Services[0].Type)

	// Second service - LoadBalancer with external IP
	assert.Equal(t, "nginx-svc", svcResult.Services[1].Name)
	assert.Equal(t, "LoadBalancer", svcResult.Services[1].Type)
	assert.Equal(t, "203.0.113.10", svcResult.Services[1].ExternalIP)
	require.Len(t, svcResult.Services[1].Ports, 1)
	assert.Equal(t, 80, svcResult.Services[1].Ports[0].Port)
	assert.Equal(t, 30080, svcResult.Services[1].Ports[0].NodePort)

	// Third service - NodePort
	assert.Equal(t, "redis-svc", svcResult.Services[2].Name)
	assert.Equal(t, "NodePort", svcResult.Services[2].Type)
	assert.Equal(t, 6379, svcResult.Services[2].Ports[0].Port)
	assert.Equal(t, 31379, svcResult.Services[2].Ports[0].NodePort)
}

func TestGetServicesParser_ParseMultiplePorts(t *testing.T) {
	input := `NAME         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)              AGE
my-service   ClusterIP   10.96.50.100   <none>        80/TCP,443/TCP       5d`

	parser := kubectl.NewGetServicesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	svcResult, ok := result.Data.(*kubectl.GetServicesResult)
	require.True(t, ok)

	require.Len(t, svcResult.Services, 1)
	require.Len(t, svcResult.Services[0].Ports, 2)

	assert.Equal(t, 80, svcResult.Services[0].Ports[0].Port)
	assert.Equal(t, "TCP", svcResult.Services[0].Ports[0].Protocol)
	assert.Equal(t, 443, svcResult.Services[0].Ports[1].Port)
	assert.Equal(t, "TCP", svcResult.Services[0].Ports[1].Protocol)
}

func TestGetServicesParser_ParseWithNamespace(t *testing.T) {
	input := `NAMESPACE     NAME           TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
default       kubernetes     ClusterIP   10.96.0.1      <none>        443/TCP   30d
kube-system   kube-dns       ClusterIP   10.96.0.10     <none>        53/UDP    30d`

	parser := kubectl.NewGetServicesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	svcResult, ok := result.Data.(*kubectl.GetServicesResult)
	require.True(t, ok)

	require.Len(t, svcResult.Services, 2)

	assert.Equal(t, "default", svcResult.Services[0].Namespace)
	assert.Equal(t, "kubernetes", svcResult.Services[0].Name)

	assert.Equal(t, "kube-system", svcResult.Services[1].Namespace)
	assert.Equal(t, "kube-dns", svcResult.Services[1].Name)
	assert.Equal(t, "UDP", svcResult.Services[1].Ports[0].Protocol)
}

func TestGetServicesParser_ParseDifferentServiceTypes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
	}{
		{
			name: "ClusterIP",
			input: `NAME     TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)   AGE
svc-1    ClusterIP   10.96.0.100   <none>        80/TCP    1d`,
			expectedType: "ClusterIP",
		},
		{
			name: "NodePort",
			input: `NAME     TYPE       CLUSTER-IP    EXTERNAL-IP   PORT(S)          AGE
svc-1    NodePort   10.96.0.100   <none>        80:30080/TCP     1d`,
			expectedType: "NodePort",
		},
		{
			name: "LoadBalancer",
			input: `NAME     TYPE           CLUSTER-IP    EXTERNAL-IP      PORT(S)        AGE
svc-1    LoadBalancer   10.96.0.100   203.0.113.100    80:30080/TCP   1d`,
			expectedType: "LoadBalancer",
		},
		{
			name: "ExternalName",
			input: `NAME     TYPE           CLUSTER-IP   EXTERNAL-IP         PORT(S)   AGE
svc-1    ExternalName   <none>       my.database.com     <none>    1d`,
			expectedType: "ExternalName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := kubectl.NewGetServicesParser()
			result, err := parser.Parse(strings.NewReader(tt.input))

			require.NoError(t, err)
			require.Nil(t, result.Error)

			svcResult, ok := result.Data.(*kubectl.GetServicesResult)
			require.True(t, ok)

			require.Len(t, svcResult.Services, 1)
			assert.Equal(t, tt.expectedType, svcResult.Services[0].Type)
		})
	}
}

func TestGetServicesParser_Matches(t *testing.T) {
	parser := kubectl.NewGetServicesParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"kubectl get services", "kubectl", []string{"get", "services"}, true},
		{"kubectl get service", "kubectl", []string{"get", "service"}, true},
		{"kubectl get svc", "kubectl", []string{"get", "svc"}, true},
		{"kubectl get services with flags", "kubectl", []string{"get", "services", "-n", "default"}, true},
		{"kubectl get pods", "kubectl", []string{"get", "pods"}, false},
		{"kubectl get deployments", "kubectl", []string{"get", "deployments"}, false},
		{"kubectl only", "kubectl", []string{}, false},
		{"kubectl get only", "kubectl", []string{"get"}, false},
		{"docker get services", "docker", []string{"get", "services"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetServicesParser_Schema(t *testing.T) {
	parser := kubectl.NewGetServicesParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "services")
}

func TestGetServicesParser_EmptyOutput(t *testing.T) {
	input := `NAME   TYPE   CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE`

	parser := kubectl.NewGetServicesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	svcResult, ok := result.Data.(*kubectl.GetServicesResult)
	require.True(t, ok)

	assert.Empty(t, svcResult.Services)
}

func TestGetServicesParser_NoResourcesFound(t *testing.T) {
	input := noResourcesFoundMsg

	parser := kubectl.NewGetServicesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	svcResult, ok := result.Data.(*kubectl.GetServicesResult)
	require.True(t, ok)

	assert.Empty(t, svcResult.Services)
}

func TestGetServicesParser_PendingExternalIP(t *testing.T) {
	input := `NAME     TYPE           CLUSTER-IP    EXTERNAL-IP   PORT(S)        AGE
nginx    LoadBalancer   10.96.0.100   <pending>     80:30080/TCP   1m`

	parser := kubectl.NewGetServicesParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	svcResult, ok := result.Data.(*kubectl.GetServicesResult)
	require.True(t, ok)

	require.Len(t, svcResult.Services, 1)
	// <pending> should be treated as empty/no external IP
	assert.Equal(t, "", svcResult.Services[0].ExternalIP)
}
