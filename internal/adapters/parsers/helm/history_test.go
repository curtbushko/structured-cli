package helm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryParser_Parse(t *testing.T) {
	t.Run("parses single revision", func(t *testing.T) {
		input := `REVISION	UPDATED                 	STATUS  	CHART       	APP VERSION	DESCRIPTION
1       	Mon Jan 10 08:15:30 2024	deployed	nginx-1.2.0 	1.24.0     	Install complete
`
		parser := NewHistoryParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		historyResult, ok := result.Data.(*HistoryResult)
		require.True(t, ok, "expected *HistoryResult")
		require.Len(t, historyResult.Revisions, 1)

		assert.Equal(t, 1, historyResult.Revisions[0].Revision)
		assert.Contains(t, historyResult.Revisions[0].Updated, "Mon Jan 10")
		assert.Equal(t, "deployed", historyResult.Revisions[0].Status)
		assert.Equal(t, "nginx-1.2.0", historyResult.Revisions[0].Chart)
		assert.Equal(t, "1.24.0", historyResult.Revisions[0].AppVersion)
		assert.Equal(t, "Install complete", historyResult.Revisions[0].Description)
	})

	t.Run("parses multiple revisions", func(t *testing.T) {
		input := `REVISION	UPDATED                 	STATUS    	CHART       	APP VERSION	DESCRIPTION
1       	Mon Jan 10 08:15:30 2024	superseded	nginx-1.2.0 	1.24.0     	Install complete
2       	Fri Jan 12 14:20:15 2024	superseded	nginx-1.2.1 	1.24.0     	Upgrade complete
3       	Mon Jan 15 10:30:45 2024	deployed  	nginx-1.2.3 	1.25.0     	Upgrade complete
`
		parser := NewHistoryParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		historyResult, ok := result.Data.(*HistoryResult)
		require.True(t, ok)
		require.Len(t, historyResult.Revisions, 3)

		// Check first revision (superseded)
		assert.Equal(t, 1, historyResult.Revisions[0].Revision)
		assert.Equal(t, "superseded", historyResult.Revisions[0].Status)
		assert.Equal(t, "nginx-1.2.0", historyResult.Revisions[0].Chart)

		// Check second revision (superseded)
		assert.Equal(t, 2, historyResult.Revisions[1].Revision)
		assert.Equal(t, "superseded", historyResult.Revisions[1].Status)
		assert.Equal(t, "nginx-1.2.1", historyResult.Revisions[1].Chart)

		// Check third revision (deployed)
		assert.Equal(t, 3, historyResult.Revisions[2].Revision)
		assert.Equal(t, "deployed", historyResult.Revisions[2].Status)
		assert.Equal(t, "nginx-1.2.3", historyResult.Revisions[2].Chart)
		assert.Equal(t, "1.25.0", historyResult.Revisions[2].AppVersion)
	})

	t.Run("parses with various statuses", func(t *testing.T) {
		input := `REVISION	UPDATED                 	STATUS         	CHART       	APP VERSION	DESCRIPTION
1       	Mon Jan 10 08:15:30 2024	superseded     	myapp-1.0.0 	1.0.0      	Install complete
2       	Tue Jan 11 09:00:00 2024	failed         	myapp-1.1.0 	1.1.0      	Upgrade "myapp" failed
3       	Wed Jan 12 10:00:00 2024	pending-upgrade	myapp-1.2.0 	1.2.0      	Preparing upgrade
`
		parser := NewHistoryParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		historyResult, ok := result.Data.(*HistoryResult)
		require.True(t, ok)
		require.Len(t, historyResult.Revisions, 3)

		assert.Equal(t, "superseded", historyResult.Revisions[0].Status)
		assert.Equal(t, "failed", historyResult.Revisions[1].Status)
		assert.Equal(t, "pending-upgrade", historyResult.Revisions[2].Status)
	})

	t.Run("handles rollback descriptions", func(t *testing.T) {
		input := `REVISION	UPDATED                 	STATUS    	CHART       	APP VERSION	DESCRIPTION
1       	Mon Jan 10 08:15:30 2024	superseded	nginx-1.2.0 	1.24.0     	Install complete
2       	Fri Jan 12 14:20:15 2024	superseded	nginx-1.2.1 	1.24.0     	Upgrade complete
3       	Sat Jan 13 16:00:00 2024	deployed  	nginx-1.2.0 	1.24.0     	Rollback to 1
`
		parser := NewHistoryParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		historyResult, ok := result.Data.(*HistoryResult)
		require.True(t, ok)
		require.Len(t, historyResult.Revisions, 3)

		assert.Equal(t, "Rollback to 1", historyResult.Revisions[2].Description)
	})

	t.Run("handles empty output", func(t *testing.T) {
		input := ``
		parser := NewHistoryParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		historyResult, ok := result.Data.(*HistoryResult)
		require.True(t, ok)
		assert.Empty(t, historyResult.Revisions)
	})

	t.Run("handles header only output", func(t *testing.T) {
		input := `REVISION	UPDATED	STATUS	CHART	APP VERSION	DESCRIPTION
`
		parser := NewHistoryParser()
		result, err := parser.Parse(strings.NewReader(input))

		require.NoError(t, err)
		require.NotNil(t, result.Data)

		historyResult, ok := result.Data.(*HistoryResult)
		require.True(t, ok)
		assert.Empty(t, historyResult.Revisions)
	})
}

func TestHistoryParser_Matches(t *testing.T) {
	parser := NewHistoryParser()

	t.Run("matches helm history", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"history", "myrelease"}))
	})

	t.Run("matches helm history with flags", func(t *testing.T) {
		assert.True(t, parser.Matches("helm", []string{"history", "myrelease", "--max", "10"}))
	})

	t.Run("does not match other commands", func(t *testing.T) {
		assert.False(t, parser.Matches("helm", []string{"list"}))
		assert.False(t, parser.Matches("helm", []string{"status"}))
		assert.False(t, parser.Matches("helm", []string{"install"}))
		assert.False(t, parser.Matches("kubectl", []string{"history"}))
		assert.False(t, parser.Matches("helm", []string{}))
	})
}

func TestHistoryParser_Schema(t *testing.T) {
	parser := NewHistoryParser()
	schema := parser.Schema()

	assert.Equal(t, "https://structured-cli.dev/schemas/helm-history.json", schema.ID)
	assert.Equal(t, "Helm History Output", schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "revisions")
	assert.Contains(t, schema.Required, "revisions")
}
