package application

import (
	"reflect"
	"testing"
)

func TestMatcher_DetectSubcommands(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantCmd         string
		wantSubcommands []string
		wantRemaining   []string
		wantErr         bool
	}{
		{
			name:            "git log with flags",
			args:            []string{"git", "log", "--oneline", "-n", "5"},
			wantCmd:         "git",
			wantSubcommands: []string{"log"},
			wantRemaining:   []string{"--oneline", "-n", "5"},
			wantErr:         false,
		},
		{
			name:            "git status no flags",
			args:            []string{"git", "status"},
			wantCmd:         "git",
			wantSubcommands: []string{"status"},
			wantRemaining:   nil,
			wantErr:         false,
		},
		{
			name:            "docker compose up with flags",
			args:            []string{"docker", "compose", "up", "-d"},
			wantCmd:         "docker",
			wantSubcommands: []string{"compose", "up"},
			wantRemaining:   []string{"-d"},
			wantErr:         false,
		},
		{
			name:            "kubectl get pods with flags",
			args:            []string{"kubectl", "get", "pods", "-n", "default"},
			wantCmd:         "kubectl",
			wantSubcommands: []string{"get", "pods"},
			wantRemaining:   []string{"-n", "default"},
			wantErr:         false,
		},
		{
			name:            "ls with flags only",
			args:            []string{"ls", "-la"},
			wantCmd:         "ls",
			wantSubcommands: nil,
			wantRemaining:   []string{"-la"},
			wantErr:         false,
		},
		{
			name:            "command only",
			args:            []string{"pwd"},
			wantCmd:         "pwd",
			wantSubcommands: nil,
			wantRemaining:   nil,
			wantErr:         false,
		},
		{
			name:    "empty args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:            "npm install package",
			args:            []string{"npm", "install", "lodash"},
			wantCmd:         "npm",
			wantSubcommands: []string{"install", "lodash"},
			wantRemaining:   nil,
			wantErr:         false,
		},
		{
			name:            "npm install with flag before package",
			args:            []string{"npm", "install", "--save-dev", "jest"},
			wantCmd:         "npm",
			wantSubcommands: []string{"install"},
			wantRemaining:   []string{"--save-dev", "jest"},
			wantErr:         false,
		},
		{
			name:            "git with double dash separator",
			args:            []string{"git", "checkout", "--", "file.txt"},
			wantCmd:         "git",
			wantSubcommands: []string{"checkout"},
			wantRemaining:   []string{"--", "file.txt"},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewCommandMatcher()

			gotCmd, gotSubcommands, gotRemaining, err := matcher.Match(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Match() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if gotCmd != tt.wantCmd {
				t.Errorf("Match() cmd = %v, want %v", gotCmd, tt.wantCmd)
			}

			if !reflect.DeepEqual(gotSubcommands, tt.wantSubcommands) {
				t.Errorf("Match() subcommands = %v, want %v", gotSubcommands, tt.wantSubcommands)
			}

			if !reflect.DeepEqual(gotRemaining, tt.wantRemaining) {
				t.Errorf("Match() remaining = %v, want %v", gotRemaining, tt.wantRemaining)
			}
		})
	}
}

func TestMatcher_MatchWithRegistry(t *testing.T) {
	// Arrange: Registry with specific parsers
	registry := NewInMemoryParserRegistry()
	gitStatusParser := &matchingParser{cmd: "git", subcommands: []string{"status"}}
	gitLogParser := &matchingParser{cmd: "git", subcommands: []string{"log"}}
	dockerComposeUpParser := &matchingParser{cmd: "docker", subcommands: []string{"compose", "up"}}

	registry.Register(gitStatusParser)
	registry.Register(gitLogParser)
	registry.Register(dockerComposeUpParser)

	matcher := NewCommandMatcher()

	tests := []struct {
		name        string
		args        []string
		wantParser  *matchingParser
		wantFound   bool
		wantRemArgs []string
	}{
		{
			name:        "matches git status",
			args:        []string{"git", "status"},
			wantParser:  gitStatusParser,
			wantFound:   true,
			wantRemArgs: nil,
		},
		{
			name:        "matches git log with args",
			args:        []string{"git", "log", "--oneline"},
			wantParser:  gitLogParser,
			wantFound:   true,
			wantRemArgs: []string{"--oneline"},
		},
		{
			name:        "matches docker compose up",
			args:        []string{"docker", "compose", "up", "-d"},
			wantParser:  dockerComposeUpParser,
			wantFound:   true,
			wantRemArgs: []string{"-d"},
		},
		{
			name:        "no match for git push",
			args:        []string{"git", "push"},
			wantParser:  nil,
			wantFound:   false,
			wantRemArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, subcommands, remaining, err := matcher.Match(tt.args)
			if err != nil {
				t.Fatalf("Match() error = %v", err)
			}

			parser, found := registry.Find(cmd, subcommands)

			if found != tt.wantFound {
				t.Errorf("registry.Find() found = %v, want %v", found, tt.wantFound)
			}

			if found && parser != tt.wantParser {
				t.Errorf("registry.Find() parser = %v, want %v", parser, tt.wantParser)
			}

			if found && !reflect.DeepEqual(remaining, tt.wantRemArgs) {
				t.Errorf("Match() remaining = %v, want %v", remaining, tt.wantRemArgs)
			}
		})
	}
}
