package helm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowValuesParser_Parse(t *testing.T) {
	t.Run("parses YAML values output", func(t *testing.T) {
		input := `## @section Global parameters
## Global Docker image parameters

global:
  imageRegistry: ""
  imagePullSecrets: []
  storageClass: ""

## @section Common parameters

replicaCount: 1
image:
  registry: docker.io
  repository: bitnami/nginx
  tag: 1.25.3-debian-11-r0
`
		parser := NewShowValuesParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		valuesResult, ok := result.Data.(*ShowValuesResult)
		require.True(t, ok, "expected *ShowValuesResult")

		// Should extract top-level keys
		require.NotEmpty(t, valuesResult.Values)

		// Find global key
		var foundGlobal, foundReplicaCount, foundImage bool
		for _, v := range valuesResult.Values {
			switch v.Key {
			case "global":
				foundGlobal = true
			case "replicaCount":
				foundReplicaCount = true
				assert.Equal(t, 1, v.Value)
			case "image":
				foundImage = true
			}
		}
		assert.True(t, foundGlobal, "should have global key")
		assert.True(t, foundReplicaCount, "should have replicaCount key")
		assert.True(t, foundImage, "should have image key")

		// Raw should contain the original output
		assert.Contains(t, valuesResult.Raw, "global:")
		assert.Contains(t, valuesResult.Raw, "replicaCount: 1")
	})

	t.Run("parses complex nested values", func(t *testing.T) {
		input := `service:
  type: ClusterIP
  port: 80
  annotations: {}

ingress:
  enabled: false
  className: ""
  hosts:
    - host: chart-example.local
      paths:
        - path: /
          pathType: ImplementationSpecific

resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 100m
    memory: 128Mi
`
		parser := NewShowValuesParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		valuesResult, ok := result.Data.(*ShowValuesResult)
		require.True(t, ok, "expected *ShowValuesResult")

		// Should extract top-level keys
		require.NotEmpty(t, valuesResult.Values)

		var foundService, foundIngress, foundResources bool
		for _, v := range valuesResult.Values {
			switch v.Key {
			case "service":
				foundService = true
			case "ingress":
				foundIngress = true
			case "resources":
				foundResources = true
			}
		}
		assert.True(t, foundService, "should have service key")
		assert.True(t, foundIngress, "should have ingress key")
		assert.True(t, foundResources, "should have resources key")

		// Raw should contain the original output
		assert.Contains(t, valuesResult.Raw, "service:")
		assert.Contains(t, valuesResult.Raw, "ingress:")
	})

	t.Run("extracts value descriptions from comments", func(t *testing.T) {
		input := `## @param replicaCount Number of replicas
replicaCount: 1

## @param image.registry Container image registry
## @param image.repository Container image repository
image:
  registry: docker.io
  repository: bitnami/nginx
`
		parser := NewShowValuesParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		valuesResult, ok := result.Data.(*ShowValuesResult)
		require.True(t, ok, "expected *ShowValuesResult")

		// Should have descriptions from comments
		var replicaValue *ChartValue
		for i := range valuesResult.Values {
			if valuesResult.Values[i].Key == "replicaCount" {
				replicaValue = &valuesResult.Values[i]
				break
			}
		}
		require.NotNil(t, replicaValue, "should have replicaCount")
		assert.Contains(t, replicaValue.Description, "Number of replicas")
	})

	t.Run("handles chart not found error output", func(t *testing.T) {
		input := `Error: chart "nonexistent" not found in https://charts.bitnami.com/bitnami repository`
		parser := NewShowValuesParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		valuesResult, ok := result.Data.(*ShowValuesResult)
		require.True(t, ok, "expected *ShowValuesResult")

		// Should have no values but raw error
		assert.Empty(t, valuesResult.Values)
		assert.Contains(t, valuesResult.Raw, "Error:")
	})

	t.Run("handles empty input", func(t *testing.T) {
		input := ``
		parser := NewShowValuesParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		valuesResult, ok := result.Data.(*ShowValuesResult)
		require.True(t, ok, "expected *ShowValuesResult")
		assert.Empty(t, valuesResult.Values)
		assert.Empty(t, valuesResult.Raw)
	})

	t.Run("handles simple key-value format", func(t *testing.T) {
		input := `enabled: true
debug: false
port: 8080
name: "myapp"
`
		parser := NewShowValuesParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		valuesResult, ok := result.Data.(*ShowValuesResult)
		require.True(t, ok, "expected *ShowValuesResult")

		require.Len(t, valuesResult.Values, 4)

		values := make(map[string]any)
		for _, v := range valuesResult.Values {
			values[v.Key] = v.Value
		}

		assert.Equal(t, true, values["enabled"])
		assert.Equal(t, false, values["debug"])
		assert.Equal(t, 8080, values["port"])
		assert.Equal(t, "myapp", values["name"])
	})
}

func TestShowValuesParser_Matches(t *testing.T) {
	parser := NewShowValuesParser()

	t.Run("matches helm show values", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"show", "values", "bitnami/nginx"}))
	})

	t.Run("matches helm show values with repo", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"show", "values", "bitnami/redis"}))
	})

	t.Run("matches helm show values with flags", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"show", "values", "bitnami/nginx", "--version", "1.0.0"}))
	})

	t.Run("does not match helm show chart", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{"show", "chart", "bitnami/nginx"}))
	})

	t.Run("does not match helm show readme", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{"show", "readme", "bitnami/nginx"}))
	})

	t.Run("does not match other commands", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{"list"}))
		assert.False(t, parser.Matches("helm", []string{"status"}))
		assert.False(t, parser.Matches("helm", []string{"show"}))
		assert.False(t, parser.Matches("kubectl", []string{"show", "values"}))
		assert.False(t, parser.Matches("helm", []string{}))
	})
}

func TestShowValuesParser_Schema(t *testing.T) {
	parser := NewShowValuesParser()
	schema := parser.Schema()

	assert.Equal(t, "https://structured-cli.dev/schemas/helm-show-values.json", schema.ID)
	assert.Equal(t, "Helm Show Values Output", schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "values")
	assert.Contains(t, schema.Properties, "raw")
	assert.Contains(t, schema.Required, "raw")
}
