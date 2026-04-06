package helm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusParser_Parse(t *testing.T) {
	t.Run("parses complete status output", func(t *testing.T) {
		input := `NAME: nginx-ingress
LAST DEPLOYED: Mon Jan 15 10:30:45 2024
NAMESPACE: default
STATUS: deployed
REVISION: 3
NOTES:
The nginx-ingress has been installed.
Get the application URL by running these commands:
  export POD_NAME=$(kubectl get pods ...)
`
		parser := NewStatusParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		statusResult, ok := result.Data.(*StatusResult)
		require.True(t, ok, "expected *StatusResult")

		assert.Equal(t, "nginx-ingress", statusResult.Name)
		assert.Equal(t, "Mon Jan 15 10:30:45 2024", statusResult.LastDeployed)
		assert.Equal(t, "default", statusResult.Namespace)
		assert.Equal(t, "deployed", statusResult.Status)
		assert.Equal(t, 3, statusResult.Revision)
		assert.Contains(t, statusResult.Notes, "The nginx-ingress has been installed")
		assert.Contains(t, statusResult.Notes, "export POD_NAME")
	})

	t.Run("parses status with resources section", func(t *testing.T) {
		input := `NAME: myapp
LAST DEPLOYED: Tue Jan 16 14:22:33 2024
NAMESPACE: production
STATUS: deployed
REVISION: 5
RESOURCES:
==> v1/Service
NAME           TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)   AGE
myapp-service  ClusterIP   10.96.45.123   <none>        80/TCP    5m

==> v1/Deployment
NAME           READY   UP-TO-DATE   AVAILABLE   AGE
myapp-deploy   3/3     3            3           5m

NOTES:
Application deployed successfully.
`
		parser := NewStatusParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		statusResult, ok := result.Data.(*StatusResult)
		require.True(t, ok, "expected *StatusResult")

		assert.Equal(t, "myapp", statusResult.Name)
		assert.Equal(t, "production", statusResult.Namespace)
		assert.Equal(t, "deployed", statusResult.Status)
		assert.Equal(t, 5, statusResult.Revision)

		// Should have parsed resources
		require.Len(t, statusResult.Resources, 2)
		assert.Equal(t, "v1/Service", statusResult.Resources[0].Kind)
		assert.Equal(t, "myapp-service", statusResult.Resources[0].Name)
		assert.Equal(t, "v1/Deployment", statusResult.Resources[1].Kind)
		assert.Equal(t, "myapp-deploy", statusResult.Resources[1].Name)

		assert.Contains(t, statusResult.Notes, "Application deployed successfully")
	})

	t.Run("parses status with notes section only", func(t *testing.T) {
		input := `NAME: redis
LAST DEPLOYED: Wed Jan 17 09:15:30 2024
NAMESPACE: cache
STATUS: deployed
REVISION: 1
NOTES:
Redis has been installed.

To connect to your Redis server:
  kubectl port-forward svc/redis-master 6379:6379
  redis-cli -h 127.0.0.1 -p 6379
`
		parser := NewStatusParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		statusResult, ok := result.Data.(*StatusResult)
		require.True(t, ok, "expected *StatusResult")

		assert.Equal(t, "redis", statusResult.Name)
		assert.Equal(t, "cache", statusResult.Namespace)
		assert.Equal(t, "deployed", statusResult.Status)
		assert.Equal(t, 1, statusResult.Revision)
		assert.Contains(t, statusResult.Notes, "Redis has been installed")
		assert.Contains(t, statusResult.Notes, "redis-cli")
	})

	t.Run("handles minimal output", func(t *testing.T) {
		input := `NAME: minimal
LAST DEPLOYED: Thu Jan 18 12:00:00 2024
NAMESPACE: default
STATUS: deployed
REVISION: 1
`
		parser := NewStatusParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		statusResult, ok := result.Data.(*StatusResult)
		require.True(t, ok, "expected *StatusResult")

		assert.Equal(t, "minimal", statusResult.Name)
		assert.Equal(t, "default", statusResult.Namespace)
		assert.Equal(t, "deployed", statusResult.Status)
		assert.Equal(t, 1, statusResult.Revision)
		assert.Empty(t, statusResult.Notes)
		assert.Empty(t, statusResult.Resources)
	})

	t.Run("handles failed status", func(t *testing.T) {
		input := `NAME: failed-app
LAST DEPLOYED: Fri Jan 19 08:45:00 2024
NAMESPACE: test
STATUS: failed
REVISION: 2
DESCRIPTION: Upgrade "failed-app" failed: timed out waiting for the condition
`
		parser := NewStatusParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		statusResult, ok := result.Data.(*StatusResult)
		require.True(t, ok, "expected *StatusResult")

		assert.Equal(t, "failed-app", statusResult.Name)
		assert.Equal(t, "failed", statusResult.Status)
		assert.Equal(t, 2, statusResult.Revision)
		assert.Contains(t, statusResult.Description, "timed out waiting")
	})

	t.Run("handles empty input", func(t *testing.T) {
		input := ``
		parser := NewStatusParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		statusResult, ok := result.Data.(*StatusResult)
		require.True(t, ok, "expected *StatusResult")
		assert.Empty(t, statusResult.Name)
	})
}

func TestStatusParser_Matches(t *testing.T) {
	parser := NewStatusParser()

	t.Run("matches helm status", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"status", "myrelease"}))
	})

	t.Run("matches helm status with release name only", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"status"}))
	})

	t.Run("matches helm status with flags", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"status", "myrelease", "-n", "namespace"}))
		assert.True(t, parser.Matches("helm", []string{"status", "myrelease", "--show-resources"}))
	})

	t.Run("does not match other commands", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{"list"}))
		assert.False(t, parser.Matches("helm", []string{"install"}))
		assert.False(t, parser.Matches("helm", []string{"upgrade"}))
		assert.False(t, parser.Matches("kubectl", []string{"status"}))
		assert.False(t, parser.Matches("helm", []string{}))
	})
}

func TestStatusParser_Schema(t *testing.T) {
	parser := NewStatusParser()
	schema := parser.Schema()

	assert.Equal(t, "https://structured-cli.dev/schemas/helm-status.json", schema.ID)
	assert.Equal(t, "Helm Status Output", schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "namespace")
	assert.Contains(t, schema.Properties, "status")
	assert.Contains(t, schema.Properties, "revision")
	assert.Contains(t, schema.Properties, "last_deployed")
	assert.Contains(t, schema.Properties, "notes")
	assert.Contains(t, schema.Properties, "resources")
	assert.Contains(t, schema.Required, "name")
	assert.Contains(t, schema.Required, "status")
}
