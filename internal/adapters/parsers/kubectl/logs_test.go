package kubectl_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/kubectl"
)

func TestLogsParser_ParsePlainOutput(t *testing.T) {
	input := `Starting application...
Listening on port 8080
Received request from 10.0.0.1
Processing request...
Request completed successfully`

	parser := kubectl.NewLogsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	logsResult, ok := result.Data.(*kubectl.LogsResult)
	require.True(t, ok, "result.Data should be *kubectl.LogsResult")

	assert.Equal(t, 5, logsResult.LineCount)
	require.Len(t, logsResult.Lines, 5)

	assert.Equal(t, "Starting application...", logsResult.Lines[0].Message)
	assert.Equal(t, "Listening on port 8080", logsResult.Lines[1].Message)
	assert.Empty(t, logsResult.Lines[0].Timestamp)
	assert.Empty(t, logsResult.Lines[0].Level)
}

func TestLogsParser_ParseWithTimestamps(t *testing.T) {
	input := `2024-01-15T10:30:00.123456789Z Starting application...
2024-01-15T10:30:01.234567890Z Listening on port 8080
2024-01-15T10:30:02.345678901Z Received request from 10.0.0.1`

	parser := kubectl.NewLogsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	logsResult, ok := result.Data.(*kubectl.LogsResult)
	require.True(t, ok)

	assert.Equal(t, 3, logsResult.LineCount)
	require.Len(t, logsResult.Lines, 3)

	assert.Equal(t, "2024-01-15T10:30:00.123456789Z", logsResult.Lines[0].Timestamp)
	assert.Equal(t, "Starting application...", logsResult.Lines[0].Message)

	assert.Equal(t, "2024-01-15T10:30:01.234567890Z", logsResult.Lines[1].Timestamp)
	assert.Equal(t, "Listening on port 8080", logsResult.Lines[1].Message)
}

func TestLogsParser_ParseWithLogLevels(t *testing.T) {
	input := `INFO Starting application...
DEBUG Initializing database connection
WARN Connection timeout, retrying...
ERROR Failed to connect to database
FATAL Application shutting down`

	parser := kubectl.NewLogsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	logsResult, ok := result.Data.(*kubectl.LogsResult)
	require.True(t, ok)

	assert.Equal(t, 5, logsResult.LineCount)
	require.Len(t, logsResult.Lines, 5)

	assert.Equal(t, "INFO", logsResult.Lines[0].Level)
	assert.Equal(t, "Starting application...", logsResult.Lines[0].Message)

	assert.Equal(t, "DEBUG", logsResult.Lines[1].Level)
	assert.Equal(t, "WARN", logsResult.Lines[2].Level)
	assert.Equal(t, "ERROR", logsResult.Lines[3].Level)
	assert.Equal(t, "FATAL", logsResult.Lines[4].Level)
}

func TestLogsParser_ParseWithTimestampsAndLevels(t *testing.T) {
	input := `2024-01-15T10:30:00.000Z INFO Starting application...
2024-01-15T10:30:01.000Z WARN Connection slow
2024-01-15T10:30:02.000Z ERROR Failed to process request`

	parser := kubectl.NewLogsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	logsResult, ok := result.Data.(*kubectl.LogsResult)
	require.True(t, ok)

	assert.Equal(t, 3, logsResult.LineCount)

	assert.Equal(t, "2024-01-15T10:30:00.000Z", logsResult.Lines[0].Timestamp)
	assert.Equal(t, "INFO", logsResult.Lines[0].Level)
	assert.Equal(t, "Starting application...", logsResult.Lines[0].Message)

	assert.Equal(t, "2024-01-15T10:30:02.000Z", logsResult.Lines[2].Timestamp)
	assert.Equal(t, "ERROR", logsResult.Lines[2].Level)
	assert.Equal(t, "Failed to process request", logsResult.Lines[2].Message)
}

func TestLogsParser_HandleEmptyLogs(t *testing.T) {
	input := ``

	parser := kubectl.NewLogsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	logsResult, ok := result.Data.(*kubectl.LogsResult)
	require.True(t, ok)

	assert.Equal(t, 0, logsResult.LineCount)
	assert.NotNil(t, logsResult.Lines)
	assert.Empty(t, logsResult.Lines)
}

func TestLogsParser_HandleContainerName(t *testing.T) {
	// The parser doesn't set container from the log content,
	// but we ensure it can be set when needed
	input := `Starting application...`

	parser := kubectl.NewLogsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	logsResult, ok := result.Data.(*kubectl.LogsResult)
	require.True(t, ok)

	// Container is empty when not specified
	assert.Empty(t, logsResult.Container)
}

func TestLogsParser_ParseJSONLogs(t *testing.T) {
	// Some apps output JSON logs - parser should handle these gracefully
	input := `{"time":"2024-01-15T10:30:00Z","level":"info","msg":"Starting"}
{"time":"2024-01-15T10:30:01Z","level":"error","msg":"Failed"}`

	parser := kubectl.NewLogsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	logsResult, ok := result.Data.(*kubectl.LogsResult)
	require.True(t, ok)

	// JSON lines are preserved as-is in the message
	assert.Equal(t, 2, logsResult.LineCount)
	assert.Contains(t, logsResult.Lines[0].Message, `"msg":"Starting"`)
}

func TestLogsParser_ParseMultilineStackTrace(t *testing.T) {
	input := `ERROR Exception caught
java.lang.NullPointerException
    at com.example.MyClass.method(MyClass.java:42)
    at com.example.Main.main(Main.java:10)
INFO Recovered from error`

	parser := kubectl.NewLogsParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	logsResult, ok := result.Data.(*kubectl.LogsResult)
	require.True(t, ok)

	// Each line is parsed separately
	assert.Equal(t, 5, logsResult.LineCount)
	assert.Equal(t, "ERROR", logsResult.Lines[0].Level)
	assert.Equal(t, "INFO", logsResult.Lines[4].Level)
}

func TestLogsParser_Matches(t *testing.T) {
	parser := kubectl.NewLogsParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"kubectl logs", "kubectl", []string{"logs"}, true},
		{"kubectl logs with pod", "kubectl", []string{"logs", "my-pod"}, true},
		{"kubectl logs with container", "kubectl", []string{"logs", "my-pod", "-c", "app"}, true},
		{"kubectl logs with flags", "kubectl", []string{"logs", "-f", "my-pod"}, true},
		{"kubectl logs with timestamps", "kubectl", []string{"logs", "--timestamps", "my-pod"}, true},
		{"kubectl get pods", "kubectl", []string{"get", "pods"}, false},
		{"kubectl describe pod", "kubectl", []string{"describe", "pod"}, false},
		{"kubectl only", "kubectl", []string{}, false},
		{"docker logs", "docker", []string{"logs"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLogsParser_Schema(t *testing.T) {
	parser := kubectl.NewLogsParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "lines")
	assert.Contains(t, schema.Properties, "line_count")
	assert.Contains(t, schema.Properties, "container")
}
