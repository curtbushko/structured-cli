package domain

import (
	"reflect"
	"testing"
)

func TestNewCommand(t *testing.T) {
	tests := []struct {
		name        string
		cmdName     string
		subcommands []string
		args        []string
		want        Command
	}{
		{
			name:        "simple command with no subcommands",
			cmdName:     "git",
			subcommands: nil,
			args:        nil,
			want: Command{
				Name:        "git",
				Subcommands: nil,
				Args:        nil,
			},
		},
		{
			name:        "command with subcommand",
			cmdName:     "git",
			subcommands: []string{"status"},
			args:        nil,
			want: Command{
				Name:        "git",
				Subcommands: []string{"status"},
				Args:        nil,
			},
		},
		{
			name:        "command with subcommand and args",
			cmdName:     "git",
			subcommands: []string{"log"},
			args:        []string{"--oneline", "-n", "10"},
			want: Command{
				Name:        "git",
				Subcommands: []string{"log"},
				Args:        []string{"--oneline", "-n", "10"},
			},
		},
		{
			name:        "command with nested subcommands",
			cmdName:     "kubectl",
			subcommands: []string{"config", "view"},
			args:        []string{"--minify"},
			want: Command{
				Name:        "kubectl",
				Subcommands: []string{"config", "view"},
				Args:        []string{"--minify"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCommand(tt.cmdName, tt.subcommands, tt.args)
			if got.Name != tt.want.Name {
				t.Errorf("NewCommand().Name = %v, want %v", got.Name, tt.want.Name)
			}
			if !reflect.DeepEqual(got.Subcommands, tt.want.Subcommands) {
				t.Errorf("NewCommand().Subcommands = %v, want %v", got.Subcommands, tt.want.Subcommands)
			}
			if !reflect.DeepEqual(got.Args, tt.want.Args) {
				t.Errorf("NewCommand().Args = %v, want %v", got.Args, tt.want.Args)
			}
		})
	}
}

func TestCommandFromArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Command
		wantErr bool
	}{
		{
			name: "git status --short",
			args: []string{"git", "status", "--short"},
			want: Command{
				Name:        "git",
				Subcommands: []string{"status"},
				Args:        []string{"--short"},
			},
			wantErr: false,
		},
		{
			name: "simple command only",
			args: []string{"ls"},
			want: Command{
				Name:        "ls",
				Subcommands: nil,
				Args:        nil,
			},
			wantErr: false,
		},
		{
			name: "command with only args",
			args: []string{"ls", "-la", "/tmp"},
			want: Command{
				Name:        "ls",
				Subcommands: nil,
				Args:        []string{"-la", "/tmp"},
			},
			wantErr: false,
		},
		{
			name:    "empty args returns error",
			args:    []string{},
			want:    Command{},
			wantErr: true,
		},
		{
			name:    "nil args returns error",
			args:    nil,
			want:    Command{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CommandFromArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommandFromArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("CommandFromArgs().Name = %v, want %v", got.Name, tt.want.Name)
			}
			if !reflect.DeepEqual(got.Subcommands, tt.want.Subcommands) {
				t.Errorf("CommandFromArgs().Subcommands = %v, want %v", got.Subcommands, tt.want.Subcommands)
			}
			if !reflect.DeepEqual(got.Args, tt.want.Args) {
				t.Errorf("CommandFromArgs().Args = %v, want %v", got.Args, tt.want.Args)
			}
		})
	}
}

func TestCommandSpec(t *testing.T) {
	tests := []struct {
		name        string
		specName    string
		subcommand  string
		description string
	}{
		{
			name:        "simple command spec",
			specName:    "git",
			subcommand:  "status",
			description: "Show the working tree status",
		},
		{
			name:        "command spec without subcommand",
			specName:    "ls",
			subcommand:  "",
			description: "List directory contents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := CommandSpec{
				Name:        tt.specName,
				Subcommand:  tt.subcommand,
				Description: tt.description,
			}
			if spec.Name != tt.specName {
				t.Errorf("CommandSpec.Name = %v, want %v", spec.Name, tt.specName)
			}
			if spec.Subcommand != tt.subcommand {
				t.Errorf("CommandSpec.Subcommand = %v, want %v", spec.Subcommand, tt.subcommand)
			}
			if spec.Description != tt.description {
				t.Errorf("CommandSpec.Description = %v, want %v", spec.Description, tt.description)
			}
		})
	}
}

func TestCommand_FullCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  Command
		want string
	}{
		{
			name: "command with subcommands and args",
			cmd: Command{
				Name:        "git",
				Subcommands: []string{"log"},
				Args:        []string{"--oneline"},
			},
			want: "git log --oneline",
		},
		{
			name: "command only",
			cmd: Command{
				Name:        "ls",
				Subcommands: nil,
				Args:        nil,
			},
			want: "ls",
		},
		{
			name: "command with multiple subcommands",
			cmd: Command{
				Name:        "kubectl",
				Subcommands: []string{"config", "view"},
				Args:        nil,
			},
			want: "kubectl config view",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cmd.FullCommand()
			if got != tt.want {
				t.Errorf("Command.FullCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
