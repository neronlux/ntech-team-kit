# ntech-team-kit

A portable, production-grade collection of skills, agents, commands, and rules that bring [Cursor Team Kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit) workflows to [OpenCode](https://opencode.ai).

If you like the rigorous internal workflows Cursor uses for code review, CI loops, and clean shipping processes, but prefer OpenCode as your AI coding agent, this kit gives you those same capabilities.

## What's included

| Component | Count | Description |
|-----------|-------|-------------|
| Skills | 18 | On-demand workflows for CI, code review, shipping, verification, and code quality |
| Agents | 2 | Specialized subagents, tab-selectable via `@` in OpenCode |
| Commands | 8 | Convenient `/command` shortcuts for the most common workflows |
| Rules | 2 | Automatically loaded coding standards (no inline imports, exhaustive TypeScript switches) |
| Plugin | 1 | Background CI watcher for proactive PR monitoring |

Everything installs into `~/.config/opencode/`. A single command keeps all content in sync.

## Prerequisites

- [OpenCode](https://opencode.ai) installed
- [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated (`gh auth login`)
- Git configured with your user email

Some skills require project-specific tools (e.g. `run-smoke-tests` needs Playwright).

## Installation

### Homebrew (recommended)

```bash
brew tap neronlux/tap
brew install ntech-team-kit
ntech-team-kit install
```

### From source

```bash
git clone https://github.com/neronlux/ntech-team-kit.git
cd ntech-team-kit

# Use the launcher (no build step)
./bin/ntech-team-kit install

# Or build and install the binary
go build -o /usr/local/bin/ntech-team-kit ./cmd/ntech-team-kit
ntech-team-kit install
```

## CLI reference

```
ntech-team-kit install [--link]    Install skills, agents, commands, rules, and plugin
ntech-team-kit update              Check for newer CLI + refresh all content
ntech-team-kit doctor              Health checks + daily update hint
ntech-team-kit status              Show installed files and manifest status
ntech-team-kit version             Print the CLI version
ntech-team-kit path                Print the resolved kit root directory
ntech-team-kit uninstall           Remove all installed files
```

**Options:**

- `--root <path>` — override kit root location
- `--link` — symlink instead of copy (useful for development)
- `NTECH_TEAM_KIT_ROOT` — environment variable equivalent of `--root`
- `OPENCODE_CONFIG_DIR` — override `~/.config/opencode` location

### Keeping up to date

**Homebrew:**

```bash
brew upgrade ntech-team-kit
ntech-team-kit update
```

**Source:**

```bash
cd ntech-team-kit && git pull && ntech-team-kit update
```

`ntech-team-kit update` checks GitHub for a newer CLI version and always refreshes all skills, agents, commands, and rules from the current kit tree.

`ntech-team-kit doctor` prints a one-line update hint at most once per day.

## Quick start

### Review and ship a branch

```
/review-and-ship
```

Structured review, suggests or writes tests, commits cleanly, opens or updates a PR.

### Fix failing CI

```
/loop-on-ci
```

Watches PR checks via `gh pr checks`, diagnoses failures, applies fixes, and iterates until green.

### Deep code quality audit

```
@thermo-nuclear-code-quality-review review the current branch
```

The "thermo-nuclear" maintainability audit: 1k-line rule, code-judo moves, spaghetti detection, and ambitious structural simplification. Tab-selectable via `@`.

### Verify a claim with evidence

```
/verify-this The new retry logic handles timeouts correctly
```

Captures baseline vs treatment, returns `VERIFIED`, `NOT VERIFIED`, or `INCONCLUSIVE`.

### Start fresh work with a PR

```
/new-branch-and-pr
```

Creates a descriptive branch from latest main, completes implementation, commits, and opens a PR.

## Skills

Skills load on demand when invoked via the `skill` tool or `/command`.

| Skill | Trigger | Description |
|-------|---------|-------------|
| `check-compiler-errors` | Compile or type-check failures | Run the repo's compile commands, summarize errors by file and type, fix iteratively |
| `control-cli` | CLI/TUI bugs, startup regressions, hangs | Build a local harness (tmux or PTY) to drive and inspect interactive CLIs |
| `control-ui` | UI bugs, visual verification, perf profiles | Build a local browser/CDP harness to drive and inspect web or Electron UIs |
| `deslop` | AI-generated code cleanup | Remove AI slop from the diff: unnecessary comments, defensive checks, `any` casts, deep nesting |
| `fix-ci` | Failing PR checks | Diagnose the first actionable failure, apply a minimal fix, push and re-check |
| `fix-merge-conflicts` | Unresolved merge conflicts | Resolve conflicts non-interactively with minimal edits, rebuild lockfiles, validate build |
| `get-pr-comments` | PR feedback summary | Fetch review and discussion comments, group by severity and actionability |
| `loop-on-ci` | Watch CI until green | Monitor PR checks, fix failures, iterate until all checks pass |
| `make-pr-easy-to-review` | Prepare PR for review | Clean commit history, improve PR description, add reviewer guidance, annotate the diff |
| `new-branch-and-pr` | Start new work | Create a branch, implement, test, commit, and open a PR |
| `pr-review-canvas` | Interactive PR walkthrough | Generate an HTML page with categorized files, moved-code detection, and inline diffs |
| `review-and-ship` | Ship changes safely | Review for bugs and intent fit, run or write tests, commit, push, and open a PR |
| `run-smoke-tests` | End-to-end verification | Run Playwright smoke tests, debug failures, verify fixes |
| `thermo-nuclear-code-quality-review` | Strict maintainability audit | 1k-line rule, code-judo, spaghetti elimination, ambitious structural simplification |
| `verify-this` | Prove or disprove a claim | Capture baseline and treatment, compare artifacts, return a verdict |
| `weekly-review` | Weekly work summary | Synthesize authored commits into a recap grouped by bugfix, tech debt, and net-new |
| `what-did-i-get-done` | Status update | Summarize your commits over a time range into a concise update |
| `workflow-from-chats` | Extract preferences | Mine recent sessions for durable working preferences, convert into skills or rules |

## Agents

Both agents are tab-selectable via `@` in OpenCode.

| Agent | Description |
|-------|-------------|
| `ci-watcher` | Background CI monitoring. Polls PR checks and notifies on failure or success. Requires the plugin enabled via `OPENCODE_NTECH_CI_WATCH=1`. |
| `thermo-nuclear-code-quality-review` | Deep maintainability auditor. Gathers its own context (diff + file contents) when invoked directly. Also available as a subagent via `Task`. |

## Commands

Type `/` in OpenCode to see available commands:

| Command | Skill | What it does |
|---------|-------|--------------|
| `/review-and-ship` | `review-and-ship` | Review, test, commit, and ship |
| `/loop-on-ci` | `loop-on-ci` | Watch CI and fix until green |
| `/verify-this` | `verify-this` | Prove or disprove a claim |
| `/run-smoke-tests` | `run-smoke-tests` | Run Playwright smoke tests |
| `/fix-ci` | `fix-ci` | Fix the first failing CI check |
| `/new-branch-and-pr` | `new-branch-and-pr` | Start fresh work with a PR |
| `/make-pr-easy-to-review` | `make-pr-easy-to-review` | Prepare a PR for review |
| `/fix-merge-conflicts` | `fix-merge-conflicts` | Resolve merge conflicts |

## Rules

Two rules are automatically loaded into every OpenCode session:

- **No inline imports** — Imports must stay at the top of the module
- **TypeScript exhaustive switch** — Use `never` in default cases for discriminated unions

## Plugin: Background CI Watcher

The `ci-watcher` plugin monitors your PR checks in the background while an OpenCode session is idle.

**Enable:**

```bash
export OPENCODE_NTECH_CI_WATCH=1
```

The plugin polls every 60 seconds (up to 30 times) and notifies you when:
- A CI check fails (with the failing check names)
- All checks pass
- The polling budget is exhausted

This is optional — the same functionality is available interactively via `/loop-on-ci` and `/fix-ci`.

## Configuration

The repo ships an `opencode.jsonc` with sensible defaults:

- Auto-loads the two rules via `instructions` glob
- Requires confirmation before loading `thermo-nuclear-code-quality-review` (read-only, but heavy)
- Restricts agent permissions to read-only (`gh` and `git` only)

Merge relevant parts into your own `~/.config/opencode/opencode.json` as needed.

## Architecture

The CLI is a self-contained Go program (`cmd/ntech-team-kit` + `internal/kit/`):

- **Pure Go** — all commands run natively with no shell script delegation
- **Kit root resolution** — `NTECH_TEAM_KIT_ROOT` env var, compiled ldflags (Homebrew), or auto-detection from binary location
- **Atomic file writes** — content is copied via temp + rename to handle symlink edge cases
- **Atomic manifest** — installed file list is collected in memory and written atomically at the end of each install
- **Doctor checks** — validates OpenCode, `gh`, authentication, kit layout, and manifest integrity

## Differences from cursor-team-kit

| Area | Cursor Team Kit | ntech-team-kit (OpenCode) |
|------|-----------------|---------------------------|
| Installation | `/add-plugin` | `brew install ntech-team-kit` or `git clone` + `ntech-team-kit install` |
| Plugin system | Cursor plugin manifest | OpenCode skills + agents + TypeScript plugin |
| Background agents | `is_background: true` | TypeScript plugin using OpenCode session events |
| Rules | `.mdc` files with `alwaysApply` | Loaded via `instructions` glob + `AGENTS.md` |
| Commands | Not available | First-class `/command` support |
| CLI | Shell script installer | Pure Go binary with atomic installs |

## Development

### Requirements

- [Go](https://go.dev) 1.24+
- [Bun](https://bun.sh) (for TypeScript type checking and plugin builds)

### Commands

```bash
bun install                     # Install dev dependencies
bun run typecheck               # TypeScript type checking
bun run build:plugin            # Build the CI watcher plugin
go build ./cmd/ntech-team-kit   # Build the CLI
go test ./...                   # Run Go tests (22 tests)
bun run test                    # Full suite: typecheck + build + Go tests
bun run vale                    # Lint documentation prose
```

## License

MIT License

- Original work: Copyright (c) 2026 Cursor
- Port and adaptations: Copyright (c) 2026 Nathan Luxford

Upstream source: [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit)
