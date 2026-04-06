package helm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListParser_Parse(t *testing.T) {
	t.Run("parses multiple releases", func(t *testing.T) {
		input := `NAME        	NAMESPACE	REVISION	UPDATED                                	STATUS  	CHART           	APP VERSION
nginx       	default  	1       	2024-01-15 10:30:45.123456789 +0000 UTC	deployed	nginx-1.0.0     	1.19.0
redis       	cache    	3       	2024-01-14 09:15:30.987654321 +0000 UTC	deployed	redis-17.3.11   	7.2.3
`
		parser := NewListParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		listResult, ok := result.Data.(*ListResult)
		require.True(t, ok, "expected *ListResult")
		require.Len(t, listResult.Releases, 2)

		// Check first release
		assert.Equal(t, "nginx", listResult.Releases[0].Name)
		assert.Equal(t, "default", listResult.Releases[0].Namespace)
		assert.Equal(t, 1, listResult.Releases[0].Revision)
		assert.Equal(t, "deployed", listResult.Releases[0].Status)
		assert.Equal(t, "nginx-1.0.0", listResult.Releases[0].Chart)
		assert.Equal(t, "1.19.0", listResult.Releases[0].AppVersion)
		assert.Contains(t, listResult.Releases[0].Updated, "2024-01-15")

		// Check second release
		assert.Equal(t, "redis", listResult.Releases[1].Name)
		assert.Equal(t, "cache", listResult.Releases[1].Namespace)
		assert.Equal(t, 3, listResult.Releases[1].Revision)
		assert.Equal(t, "deployed", listResult.Releases[1].Status)
		assert.Equal(t, "redis-17.3.11", listResult.Releases[1].Chart)
		assert.Equal(t, "7.2.3", listResult.Releases[1].AppVersion)
	})

	t.Run("parses release with failed status", func(t *testing.T) {
		input := `NAME    	NAMESPACE	REVISION	UPDATED                                	STATUS	CHART       	APP VERSION
myapp   	prod     	5       	2024-01-16 14:22:33.111222333 +0000 UTC	failed	myapp-2.1.0 	3.0.0
`
		parser := NewListParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		listResult, ok := result.Data.(*ListResult)
		require.True(t, ok)
		require.Len(t, listResult.Releases, 1)

		assert.Equal(t, "myapp", listResult.Releases[0].Name)
		assert.Equal(t, "failed", listResult.Releases[0].Status)
	})

	t.Run("handles empty output", func(t *testing.T) {
		input := ``
		parser := NewListParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		listResult, ok := result.Data.(*ListResult)
		require.True(t, ok)
		assert.Empty(t, listResult.Releases)
	})

	t.Run("handles header only output", func(t *testing.T) {
		input := `NAME	NAMESPACE	REVISION	UPDATED	STATUS	CHART	APP VERSION
`
		parser := NewListParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		listResult, ok := result.Data.(*ListResult)
		require.True(t, ok)
		assert.Empty(t, listResult.Releases)
	})
}

func TestListParser_Matches(t *testing.T) {
	parser := NewListParser()

	t.Run("matches helm list", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"list"}))
	})

	t.Run("matches helm ls alias", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"ls"}))
	})

	t.Run("matches helm list with flags", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"list", "-A"}))
		assert.True(t, parser.Matches("helm", []string{"ls", "--all-namespaces"}))
	})

	t.Run("does not match other commands", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{"status"}))
		assert.False(t, parser.Matches("helm", []string{"install"}))
		assert.False(t, parser.Matches("kubectl", []string{"list"}))
		assert.False(t, parser.Matches("helm", []string{}))
	})
}

func TestListParser_Schema(t *testing.T) {
	parser := NewListParser()
	schema := parser.Schema()

	assert.Equal(t, "https://structured-cli.dev/schemas/helm-list.json", schema.ID)
	assert.Equal(t, "Helm List Output", schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "releases")
	assert.Contains(t, schema.Required, "releases")
}
