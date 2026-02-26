package docker

import (
	"strings"
	"testing"
)

func TestLogsParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData LogsResult
	}{
		{
			name:  "empty output",
			input: "",
			wantData: LogsResult{
				Success:    true,
				Lines:      []LogLine{},
				TotalLines: 0,
			},
		},
		{
			name: "simple log lines",
			input: `Starting server...
Listening on port 8080
Connection received from 192.168.1.1`,
			wantData: LogsResult{
				Success: true,
				Lines: []LogLine{
					{Stream: "stdout", Message: "Starting server..."},
					{Stream: "stdout", Message: "Listening on port 8080"},
					{Stream: "stdout", Message: "Connection received from 192.168.1.1"},
				},
				TotalLines: 3,
			},
		},
		{
			name: "log lines with timestamps",
			input: `2024-01-15T10:30:00.000000000Z Starting server...
2024-01-15T10:30:01.000000000Z Listening on port 8080
2024-01-15T10:30:05.000000000Z Connection received`,
			wantData: LogsResult{
				Success: true,
				Lines: []LogLine{
					{Timestamp: "2024-01-15T10:30:00.000000000Z", Stream: "stdout", Message: "Starting server..."},
					{Timestamp: "2024-01-15T10:30:01.000000000Z", Stream: "stdout", Message: "Listening on port 8080"},
					{Timestamp: "2024-01-15T10:30:05.000000000Z", Stream: "stdout", Message: "Connection received"},
				},
				TotalLines: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewLogsParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*LogsResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *LogsResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("LogsResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.TotalLines != tt.wantData.TotalLines {
				t.Errorf("LogsResult.TotalLines = %d, want %d", got.TotalLines, tt.wantData.TotalLines)
			}

			if len(got.Lines) != len(tt.wantData.Lines) {
				t.Errorf("LogsResult.Lines length = %d, want %d", len(got.Lines), len(tt.wantData.Lines))
				return
			}

			for i, line := range got.Lines {
				want := tt.wantData.Lines[i]
				if line.Message != want.Message {
					t.Errorf("Line[%d].Message = %q, want %q", i, line.Message, want.Message)
				}
				if line.Stream != want.Stream {
					t.Errorf("Line[%d].Stream = %q, want %q", i, line.Stream, want.Stream)
				}
			}
		})
	}
}

func TestLogsParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker logs",
			cmd:         "docker",
			subcommands: []string{"logs"},
			want:        true,
		},
		{
			name:        "matches docker logs with container",
			cmd:         "docker",
			subcommands: []string{"logs", "mycontainer"},
			want:        true,
		},
		{
			name:        "matches docker logs with flags",
			cmd:         "docker",
			subcommands: []string{"logs", "-f", "--tail", "100", "mycontainer"},
			want:        true,
		},
		{
			name:        "matches docker container logs",
			cmd:         "docker",
			subcommands: []string{"container", "logs"},
			want:        true,
		},
		{
			name:        "does not match docker ps",
			cmd:         "docker",
			subcommands: []string{"ps"},
			want:        false,
		},
		{
			name:        "does not match podman",
			cmd:         "podman",
			subcommands: []string{"logs"},
			want:        false,
		},
	}

	parser := NewLogsParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestLogsParser_Schema(t *testing.T) {
	parser := NewLogsParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema.ID should not be empty")
	}

	if schema.Title == "" {
		t.Error("Schema.Title should not be empty")
	}

	if schema.Type != schemaTypeObject {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, schemaTypeObject)
	}

	requiredProps := []string{"success", "lines", "total_lines"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
