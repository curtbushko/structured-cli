# structured-cli

A universal CLI wrapper that transforms raw CLI output into structured JSON. Built for AI coding agents that need to parse command output without regex.

## Features

- **81 Parsers** - Support for git, npm, docker, cargo, go, python, and more
- **71 JSON Schemas** - Documented, validated output formats
- **Drop-in Replacement** - Use as an alias without breaking existing workflows
- **Token Efficient** - Up to 95% reduction in tokens vs raw output
- **Exit Code Propagation** - Preserves underlying command exit codes

## Installation

### From Source

```bash
go install github.com/curtbushko/structured-cli/cmd/structured-cli@latest
```

### Build Locally

```bash
git clone https://github.com/curtbushko/structured-cli.git
cd structured-cli
go build -o structured-cli ./cmd/structured-cli
```

## Usage

### JSON Mode

Use `--json` flag or set `STRUCTURED_CLI_JSON=true`:

```bash
# With flag
structured-cli git status --json

# With environment variable
STRUCTURED_CLI_JSON=true structured-cli git status
```

**Output:**
```json
{
  "branch": "main",
  "upstream": "origin/main",
  "ahead": 2,
  "behind": 0,
  "clean": false,
  "staged": [{"file": "src/main.go", "status": "modified"}],
  "modified": ["README.md"],
  "untracked": ["temp.log"],
  "deleted": [],
  "conflicts": []
}
```

### Passthrough Mode (Default)

Without `--json`, output passes through unchanged:

```bash
structured-cli git status
# Same output as running `git status` directly
```

### As an Alias

```bash
alias git="structured-cli git"

# Normal usage - unchanged behavior
git status

# JSON mode when needed
git status --json
```

## Supported Commands

### Git (12 commands)
| Command | Description |
|---------|-------------|
| `git status` | Repository status with staged/modified/untracked files |
| `git log` | Commit history with author, date, message |
| `git diff` | File changes with additions/deletions |
| `git branch` | Branch listing with current and upstream info |
| `git show` | Commit details with diff |
| `git add` | Stage files |
| `git commit` | Create commits |
| `git push` | Push to remote |
| `git pull` | Pull from remote |
| `git checkout` | Switch branches |
| `git blame` | Line-by-line attribution |
| `git reflog` | Reference log entries |

### Go (7 commands)
| Command | Description |
|---------|-------------|
| `go build` | Build output with errors |
| `go test` | Test results with pass/fail/skip counts |
| `go vet` | Static analysis issues |
| `go run` | Program execution with stdout/stderr |
| `go mod tidy` | Module cleanup with added/removed deps |
| `gofmt` | Format check with unformatted files |
| `go generate` | Code generation results |

### NPM (7 commands)
| Command | Description |
|---------|-------------|
| `npm install` | Package installation with added/removed counts |
| `npm audit` | Security vulnerabilities by severity |
| `npm outdated` | Packages needing updates |
| `npm list` | Dependency tree |
| `npm run` | Script execution |
| `npm test` | Test execution |
| `npm init` | Package initialization |

### Docker (10 commands)
| Command | Description |
|---------|-------------|
| `docker ps` | Container listing |
| `docker build` | Build results with image ID |
| `docker logs` | Container log entries |
| `docker images` | Image listing |
| `docker run` | Container creation |
| `docker exec` | Command execution in container |
| `docker pull` | Image pull results |
| `docker compose up` | Service startup |
| `docker compose down` | Service shutdown |
| `docker compose ps` | Service listing |

### Python (5 commands)
| Command | Description |
|---------|-------------|
| `pip install` | Package installation |
| `pip-audit` | Security vulnerabilities |
| `uv pip install` | Fast package installation |
| `uv run` | Script execution |
| `black --check` | Format check |

### Cargo/Rust (9 commands)
| Command | Description |
|---------|-------------|
| `cargo build` | Build results |
| `cargo test` | Test results |
| `cargo clippy` | Lint warnings |
| `cargo run` | Binary execution |
| `cargo add` | Add dependency |
| `cargo remove` | Remove dependency |
| `cargo fmt` | Format check |
| `cargo doc` | Documentation generation |
| `cargo check` | Type checking |

### Build Tools (5 commands)
| Command | Description |
|---------|-------------|
| `tsc` | TypeScript compilation |
| `esbuild` | JavaScript bundling |
| `vite build` | Vite production build |
| `webpack` | Webpack bundling |
| `cargo build` | Rust compilation |

### Linters (6 commands)
| Command | Description |
|---------|-------------|
| `eslint` | JavaScript/TypeScript linting |
| `prettier --check` | Format check |
| `biome check` | Lint and format |
| `golangci-lint` | Go linting |
| `ruff check` | Python linting |
| `mypy` | Python type checking |

### Test Runners (5 commands)
| Command | Description |
|---------|-------------|
| `pytest` | Python tests |
| `jest` | JavaScript tests |
| `vitest` | Vite tests |
| `mocha` | Node.js tests |
| `cargo test` | Rust tests |

### Make/Just (2 commands)
| Command | Description |
|---------|-------------|
| `make` | Makefile targets |
| `just` | Justfile recipes |

### File Operations (11 commands)
| Command | Description |
|---------|-------------|
| `ls` | Directory listing |
| `find` | File search |
| `grep` | Content search |
| `rg` (ripgrep) | Fast content search |
| `fd` | Fast file search |
| `cat` | File contents |
| `head` | First N lines |
| `tail` | Last N lines |
| `wc` | Word/line/byte counts |
| `du` | Disk usage |
| `df` | Filesystem info |

## Error Handling

Errors are returned as JSON when in JSON mode:

```bash
$ structured-cli git status --json  # not in a git repo
{"error": "fatal: not a git repository", "exit_code": 128}
```

The exit code is propagated from the underlying command.

### Unsupported Commands

Commands without parsers return raw output in a JSON wrapper:

```bash
$ structured-cli git stash show --json
{"raw": "...", "parsed": false}
```

## Architecture

structured-cli follows hexagonal architecture (ports & adapters):

```
cmd/                  # Composition root
internal/
  domain/             # Pure types (Command, ParseResult, Schema)
  ports/              # Interfaces (Parser, Runner, Writer)
  application/        # Services (Executor, Registry)
  adapters/
    cli/              # Cobra CLI handler
    runner/           # os/exec wrapper
    parsers/          # 81 parser implementations
      git/
      npm/
      docker/
      ...
schemas/              # 71 JSON Schema files
```

## Development

### Prerequisites

- Go 1.22+
- golangci-lint
- go-arch-lint

### Build

```bash
go build ./cmd/structured-cli
```

### Test

```bash
go test -race ./...
```

### Lint

```bash
golangci-lint run
go-arch-lint check
```

## License

MIT
