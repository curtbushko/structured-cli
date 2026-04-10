# RTK Token Savings Architecture

This document explains how RTK (Rust Token Killer) saves tokens across all commands and tracks telemetry for analytics.

## Overview

RTK is a CLI proxy that intercepts command output and applies filtering strategies to reduce token consumption by 60-90%. Every filtered command is tracked in a local SQLite database for analytics.

## Architecture Flow

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐     ┌──────────────┐
│  rtk <cmd>  │────>│  main.rs     │────>│  Filter     │────>│  Tracking    │
│  (user)     │     │  (routing)   │     │  Module     │     │  (SQLite)    │
└─────────────┘     └──────────────┘     └─────────────┘     └──────────────┘
                           │                    │                    │
                           v                    v                    v
                    Commands enum         Execute cmd         Record metrics
                    routes to             + filter output     + cleanup old
                    appropriate           + print result
                    module
```

## Token Saving Strategies

RTK uses 4 complementary strategies to reduce tokens:

### 1. Hardcoded Rust Filters (`src/cmds/`)

Native Rust modules with regex-based filtering. Used for high-traffic commands needing complex logic.

**Location**: `src/cmds/<ecosystem>/<cmd>.rs`

**Ecosystems**:
- `git/` - git, gh, gt, diff
- `rust/` - cargo, runner (err/test)
- `js/` - npm, pnpm, vitest, tsc, next, prettier, playwright, prisma
- `python/` - ruff, pytest, mypy, pip
- `go/` - go, golangci-lint
- `dotnet/` - dotnet, binlog, trx
- `cloud/` - aws, docker, kubectl, curl, wget, psql
- `system/` - ls, tree, read, grep, find, wc, env, json, log
- `ruby/` - rake, rspec, rubocop

**Example** (`src/cmds/git/gh_cmd.rs`):
```rust
lazy_static! {
    static ref HTML_COMMENT_RE: Regex = Regex::new(r"(?s)<!--.*?-->").unwrap();
    static ref BADGE_LINE_RE: Regex = Regex::new(r"(?m)^\s*\[!\[[^\]]*\]\([^)]*\)\]\([^)]*\)\s*$").unwrap();
}

fn filter_markdown_body(body: &str) -> String {
    // Remove HTML comments, badges, images, horizontal rules
    // Collapse multiple blank lines
    // Preserve code blocks untouched
}
```

**Techniques used**:
- Strip ANSI escape codes
- Remove HTML comments and badges from markdown
- Collapse verbose output to essential fields
- Truncate long lines
- Extract only errors/warnings from test output
- Condense git commits to hash + message

### 2. TOML Filter DSL (`src/filters/`)

Declarative filters defined in TOML. Used for simpler commands or community contributions.

**Lookup priority** (first match wins):
1. `.rtk/filters.toml` - project-local
2. `~/.config/rtk/filters.toml` - user-global
3. Built-in (`src/filters/*.toml`) - compiled into binary

**Pipeline stages** (applied in order):
1. `strip_ansi` - remove ANSI escape codes
2. `replace` - regex substitutions, line-by-line
3. `match_output` - short-circuit if pattern matches
4. `strip_lines_matching` / `keep_lines_matching` - filter lines
5. `truncate_lines_at` - limit line length
6. `head_lines` / `tail_lines` - keep first/last N lines
7. `max_lines` - absolute line cap
8. `on_empty` - message if result is empty

**Example** (`src/filters/turbo.toml`):
```toml
[filters.turbo]
description = "Compact Turborepo output"
match_command = "^turbo\\b"
strip_ansi = true
strip_lines_matching = [
  "^\\s*$",
  "^\\s*cache (hit|miss|bypass)",
  "^\\s*\\d+ packages in scope",
  "^\\s*Tasks:\\s+\\d+",
  "^\\s*Duration:\\s+",
]
truncate_lines_at = 150
max_lines = 50
on_empty = "turbo: ok"
```

### 3. Language-Aware Code Filter (`src/core/filter.rs`)

Strips comments and boilerplate from source code.

**Supported languages**:
- Rust, Python, JavaScript, TypeScript, Go, C/C++, Java, Ruby, Shell
- Data formats (JSON, YAML, TOML) - no comment stripping

**Filter levels**:
- `none` - no filtering
- `minimal` - strip line comments
- `aggressive` - strip all comments and doc strings

### 4. Error-Only Extraction (`src/cmds/rust/runner.rs`)

For test commands, extract only failures and errors.

**Pattern matching**:
```rust
lazy_static! {
    static ref ERROR_PATTERNS: Vec<Regex> = vec![
        Regex::new(r"(?i)^.*error[\s:\[].*$").unwrap(),
        Regex::new(r"(?i)^.*warning[\s:\[].*$").unwrap(),
        Regex::new(r"(?i)^.*failed.*$").unwrap(),
        Regex::new(r"^error\[E\d+\]:.*$").unwrap(),  // Rust
        Regex::new(r"^Traceback.*$").unwrap(),        // Python
        Regex::new(r"^\s*at .*:\d+:\d+.*$").unwrap(), // JS/TS
    ];
}
```

**Commands**:
- `rtk err <command>` - run command, show only errors
- `rtk test <command>` - run tests, show only failures

## Tracking System

### Token Estimation

Tokens are estimated using a 4-character heuristic:

```rust
// src/core/tracking.rs:1025
pub fn estimate_tokens(text: &str) -> usize {
    (text.len() as f64 / 4.0).ceil() as usize
}
```

### Recording Flow

Every filter module uses `TimedExecution` to track:

```rust
// Start timer before command
let timer = tracking::TimedExecution::start();

// Execute command and filter
let raw_output = execute_command(...);
let filtered = filter_output(&raw_output);

// Record to database
timer.track(
    "git log -20",           // original_cmd
    "rtk git log",           // rtk_cmd
    &raw_output,             // input (raw)
    &filtered,               // output (filtered)
);
```

### Database Schema

**Location**:
- Linux: `~/.local/share/rtk/tracking.db`
- macOS: `~/Library/Application Support/rtk/tracking.db`
- Windows: `%APPDATA%\rtk\tracking.db`

**Tables**:

```sql
CREATE TABLE commands (
    id INTEGER PRIMARY KEY,
    timestamp TEXT NOT NULL,
    original_cmd TEXT NOT NULL,
    rtk_cmd TEXT NOT NULL,
    project_path TEXT,           -- current working directory
    input_tokens INTEGER NOT NULL,
    output_tokens INTEGER NOT NULL,
    saved_tokens INTEGER NOT NULL,
    savings_pct REAL NOT NULL,
    exec_time_ms INTEGER
);

CREATE TABLE parse_failures (
    id INTEGER PRIMARY KEY,
    timestamp TEXT NOT NULL,
    raw_command TEXT NOT NULL,
    error_message TEXT NOT NULL,
    fallback_succeeded INTEGER NOT NULL DEFAULT 0
);
```

**Retention**: 90-day automatic cleanup on every insert.

### Savings Calculation

```rust
// src/core/tracking.rs:351
pub fn record(&self, ..., input_tokens: usize, output_tokens: usize, ...) {
    let saved = input_tokens.saturating_sub(output_tokens);
    let pct = if input_tokens > 0 {
        (saved as f64 / input_tokens as f64) * 100.0
    } else {
        0.0
    };
    // Insert into database...
}
```

### Query APIs

**Summary** (`rtk gain`):
```rust
pub struct GainSummary {
    pub total_commands: usize,
    pub total_input: usize,
    pub total_output: usize,
    pub total_saved: usize,
    pub avg_savings_pct: f64,
    pub total_time_ms: u64,
    pub by_command: Vec<(String, usize, usize, f64, u64)>,  // top 10
    pub by_day: Vec<(String, usize)>,  // last 30 days
}
```

**Export formats**:
- `rtk gain --format json` - JSON export
- `rtk gain --format csv` - CSV export
- `rtk gain --daily/--weekly/--monthly` - aggregated stats
- `rtk gain --project` - scoped to current project
- `rtk gain --history` - recent commands

## Telemetry System

### Purpose

Optional usage ping for analytics (compile-time opt-in).

**Location**: `src/core/telemetry.rs`

### Data Collected

```json
{
  "device_hash": "sha256(salt + hostname + username)",
  "version": "0.34.3",
  "os": "linux",
  "arch": "x86_64",
  "install_method": "cargo|homebrew|nix|script|other",
  "commands_24h": 42,
  "top_commands": ["git", "cargo", "pnpm"],
  "savings_pct": 75.5,
  "tokens_saved_24h": 15000,
  "tokens_saved_total": 250000
}
```

### Privacy Controls

- **Disabled by default**: Requires `RTK_TELEMETRY_URL` at compile time
- **Opt-out**: `RTK_TELEMETRY_DISABLED=1` or `telemetry.enabled = false` in config
- **Rate-limited**: Maximum 1 ping per 23 hours
- **Anonymized**: Device hash uses random salt, no PII transmitted
- **Fire-and-forget**: 2-second timeout, errors silently ignored

### Device Hash Generation

```rust
fn generate_device_hash() -> String {
    let salt = get_or_create_salt();  // random 32-byte hex, stored in ~/.local/share/rtk/.device_salt
    let mut hasher = Sha256::new();
    hasher.update(salt.as_bytes());
    hasher.update(b":");
    hasher.update(hostname.as_bytes());
    hasher.update(b":");
    hasher.update(username.as_bytes());
    format!("{:x}", hasher.finalize())
}
```

## Expected Savings by Command Type

| Command Category | Typical Savings | Strategy |
|-----------------|-----------------|----------|
| `git log` | 80%+ | Condense to hash + message |
| `cargo test` | 90%+ | Show failures only |
| `gh pr view` | 87%+ | Remove ASCII art, badges |
| `pnpm list` | 70%+ | Compact dependency tree |
| `docker ps` | 60%+ | Essential fields only |
| `npm install` | 75%+ | Strip progress bars |
| `turbo` | 70%+ | Strip cache noise |

## Performance Targets

- **Startup time**: <10ms
- **Memory usage**: <5MB
- **Binary size**: <5MB

These are enforced through:
- No async runtime (tokio adds 5-10ms startup)
- `lazy_static!` for all regex compilation
- Single-threaded execution
- Minimal dependencies

## Fallback Behavior

If any filter fails, RTK falls back to passthrough:

```rust
let filtered = filter_output(&output.stdout)
    .unwrap_or_else(|e| {
        eprintln!("rtk: filter warning: {}", e);
        output.stdout.clone()  // Return unfiltered
    });
```

This ensures RTK never blocks the user workflow.
