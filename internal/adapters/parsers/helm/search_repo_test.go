package helm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchRepoParser_Parse(t *testing.T) {
	t.Run("parses single chart result", func(t *testing.T) {
		input := `NAME                            CHART VERSION   APP VERSION     DESCRIPTION
bitnami/nginx                   15.0.2          1.25.3          NGINX Open Source is a web server...
`
		parser := NewSearchRepoParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		searchResult, ok := result.Data.(*SearchResult)
		require.True(t, ok, "expected *SearchResult")
		require.Len(t, searchResult.Charts, 1)

		assert.Equal(t, "bitnami/nginx", searchResult.Charts[0].Name)
		assert.Equal(t, "15.0.2", searchResult.Charts[0].ChartVersion)
		assert.Equal(t, "1.25.3", searchResult.Charts[0].AppVersion)
		assert.Contains(t, searchResult.Charts[0].Description, "NGINX")
	})

	t.Run("parses multiple chart results", func(t *testing.T) {
		input := `NAME                            CHART VERSION   APP VERSION     DESCRIPTION
bitnami/nginx                   15.0.2          1.25.3          NGINX Open Source is a web server...
bitnami/nginx-ingress-controller 9.10.1         1.9.4           NGINX Ingress Controller is an In...
stable/nginx-ldapauth-proxy     0.1.6           1.13.5          DEPRECATED - nginx proxy with lda...
`
		parser := NewSearchRepoParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		searchResult, ok := result.Data.(*SearchResult)
		require.True(t, ok, "expected *SearchResult")
		require.Len(t, searchResult.Charts, 3)

		// Check first chart
		assert.Equal(t, "bitnami/nginx", searchResult.Charts[0].Name)
		assert.Equal(t, "15.0.2", searchResult.Charts[0].ChartVersion)
		assert.Equal(t, "1.25.3", searchResult.Charts[0].AppVersion)

		// Check second chart
		assert.Equal(t, "bitnami/nginx-ingress-controller", searchResult.Charts[1].Name)
		assert.Equal(t, "9.10.1", searchResult.Charts[1].ChartVersion)
		assert.Equal(t, "1.9.4", searchResult.Charts[1].AppVersion)

		// Check third chart (from different repo)
		assert.Equal(t, "stable/nginx-ldapauth-proxy", searchResult.Charts[2].Name)
		assert.Equal(t, "0.1.6", searchResult.Charts[2].ChartVersion)
		assert.Equal(t, "1.13.5", searchResult.Charts[2].AppVersion)
		assert.Contains(t, searchResult.Charts[2].Description, "DEPRECATED")
	})

	t.Run("handles no results found", func(t *testing.T) {
		input := `No results found
`
		parser := NewSearchRepoParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		searchResult, ok := result.Data.(*SearchResult)
		require.True(t, ok, "expected *SearchResult")
		assert.Empty(t, searchResult.Charts)
	})

	t.Run("handles empty output", func(t *testing.T) {
		input := ``
		parser := NewSearchRepoParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		searchResult, ok := result.Data.(*SearchResult)
		require.True(t, ok, "expected *SearchResult")
		assert.Empty(t, searchResult.Charts)
	})

	t.Run("handles header only output", func(t *testing.T) {
		input := `NAME	CHART VERSION	APP VERSION	DESCRIPTION
`
		parser := NewSearchRepoParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		searchResult, ok := result.Data.(*SearchResult)
		require.True(t, ok, "expected *SearchResult")
		assert.Empty(t, searchResult.Charts)
	})

	t.Run("parses charts from different repos", func(t *testing.T) {
		input := `NAME                            CHART VERSION   APP VERSION     DESCRIPTION
bitnami/redis                   17.15.2         7.2.1           Redis is an open source, advanced...
stable/redis-ha                 4.4.6           6.2.5           Highly available Redis cluster...
grafana/redis-datasource        1.5.0           2.2.0           Redis datasource for Grafana
`
		parser := NewSearchRepoParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		searchResult, ok := result.Data.(*SearchResult)
		require.True(t, ok, "expected *SearchResult")
		require.Len(t, searchResult.Charts, 3)

		// Verify different repos
		assert.Equal(t, "bitnami/redis", searchResult.Charts[0].Name)
		assert.Equal(t, "stable/redis-ha", searchResult.Charts[1].Name)
		assert.Equal(t, "grafana/redis-datasource", searchResult.Charts[2].Name)
	})
}

func TestSearchRepoParser_Matches(t *testing.T) {
	parser := NewSearchRepoParser()

	t.Run("matches helm search repo", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"search", "repo"}))
	})

	t.Run("matches helm search repo with search term", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"search", "repo", "nginx"}))
	})

	t.Run("matches helm search repo with flags", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"search", "repo", "--versions"}))
		assert.True(t, parser.Matches("helm", []string{"search", "repo", "-l"}))
	})

	t.Run("does not match helm search hub", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{"search", "hub"}))
	})

	t.Run("does not match helm search only", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{"search"}))
	})

	t.Run("does not match other helm commands", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{"list"}))
		assert.False(t, parser.Matches("helm", []string{"status"}))
		assert.False(t, parser.Matches("helm", []string{"install"}))
	})

	t.Run("does not match other base commands", func(t *testing.T) {
		assert.False(t, parser.Matches("kubectl", []string{"search", "repo"}))
		assert.False(t, parser.Matches("docker", []string{"search", "repo"}))
	})

	t.Run("does not match empty subcommands", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{}))
	})
}

func TestSearchRepoParser_Schema(t *testing.T) {
	parser := NewSearchRepoParser()
	schema := parser.Schema()

	assert.Equal(t, "https://structured-cli.dev/schemas/helm-search-repo.json", schema.ID)
	assert.Equal(t, "Helm Search Repo Output", schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "charts")
	assert.Contains(t, schema.Required, "charts")
}
