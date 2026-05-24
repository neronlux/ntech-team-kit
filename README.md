<p align="center">
  <img src="https://assets.ntek.app/ntechteamkit.jpg" alt="ntech-team-kit" width="400" />
</p>

# ntech-team-kit

A portable, production-grade collection of skills, agents, commands, and rules that bring [Cursor Team Kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit) workflows to [OpenCode](https://opencode.ai).

If you like the rigorous internal workflows Cursor uses for code review, CI loops, and clean shipping processes, but prefer OpenCode as your AI coding agent, this kit gives you those same capabilities.

## What's included

| Component | Count | Description |
|-----------|-------|-------------|
| Skills | 18 | On-demand workflows for CI, code review, shipping, verification, and code quality |
| Agents | 2 | Specialized subagents, invoked via `@agent-name` in OpenCode |
| Commands | 18 | `/command` shortcuts for every skill (1:1 with skills) |
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
ntech-team-kit
```

Running `ntech-team-kit` with no arguments opens an interactive setup menu. Press Enter to install the full pack.

### From source

```bash
git clone https://github.com/neronlux/ntech-team-kit.git
cd ntech-team-kit

# Use the launcher (no build step)
./bin/ntech-team-kit

# Or build and install the binary
go build -o /usr/local/bin/ntech-team-kit ./cmd/ntech-team-kit
ntech-team-kit
```

## CLI reference

```
ntech-team-kit                         Interactive setup (recommended)
ntech-team-kit install [options]       Install selected kit components
ntech-team-kit uninstall [options]     Remove installed files, optionally by component
ntech-team-kit update                  Check for newer CLI + refresh all content
ntech-team-kit doctor                  Health checks + daily update hint
ntech-team-kit status                  Show installed files and manifest status
ntech-team-kit version                 Print the CLI version
ntech-team-kit path                    Print the resolved kit root directory
```

### Interactive mode

Running `ntech-team-kit` with no arguments opens a guided menu:

```
  ntech-team-kit 0.1.31
  ─────────────────────────────
  Kit root: /opt/homebrew/Cellar/ntech-team-kit/0.1.31
  Installed: 44 files
  ─────────────────────────────

  What would you like to do?

    1) Install full pack (recommended)
    2) Install lite pack
    3) Install agents only
    4) Install skills only
    5) Custom install  (pick components)
    6) Custom uninstall (pick components)
    7) Check status
    8) Run doctor
    9) Update (refresh all content)
    0) Quit
```

The menu loops so you can run more than one action per session. After each action it asks whether to continue or exit.

**Custom install/uninstall** shows a numbered component picker:

```
  Select components to install.

    [*] 1) skills     On-demand workflows (review, CI, shipping)
    [*] 2) agents     Specialized subagents (ci-watcher, code-quality)
    [*] 3) commands   /command shortcuts for every skill
    [*] 4) rules      Auto-loaded coding standards
    [*] 5) plugin     Background CI watcher plugin
    [*] 6) config     opencode.jsonc defaults (first-time only)

  Examples: "1 3 5" or "skills,commands" or "lite" or Enter to confirm
  Components:
```

Type numbers (`1 3 5`), comma-separated names (`skills,commands`), or a pack name (`full`, `lite`). Press Enter to confirm. Uninstall asks for confirmation before removing files.

When you pipe stdin (not a terminal), the interactive mode runs the default action (full install) and exits, so it works in CI scripts:

```bash
echo | ntech-team-kit    # equivalent to: ntech-team-kit install
```

### Non-interactive options

**Global flags** go before the command: `ntech-team-kit --root ./kit install --pack lite`

- `--root <path>` — override kit root location
- `NTECH_TEAM_KIT_ROOT` — environment variable form of `--root`
- `OPENCODE_CONFIG_DIR` — override `~/.config/opencode` location

**Install options:**

- `--pack full|lite|agents|skills` — install a named component pack (default: full)
- `--select` — choose components interactively
- `--only <components>` — install only specific components (comma-separated)
- `--without <components>` — exclude components from the selected pack
- `--link` — symlink instead of copy (useful for development)

**Uninstall options:**

- `--select` — choose components interactively
- `--only <components>` — uninstall only specific components (comma-separated)

**Install packs:**

| Pack | Components |
|------|------------|
| `full` | `skills`, `agents`, `commands`, `rules`, `plugin`, `config` |
| `lite` | `skills`, `commands`, `rules`, `config` |
| `agents` | `agents` only |
| `skills` | `skills` only |

Component names: `skills`, `agents`, `commands`, `rules`, `plugin`, `config`.

```bash
ntech-team-kit install --pack lite
ntech-team-kit install --select
ntech-team-kit install --only skills,commands
ntech-team-kit install --without plugin,agents
ntech-team-kit uninstall --select
ntech-team-kit uninstall --only agents
```

### Keeping up to date

**Homebrew:**

```bash
brew update
brew upgrade ntech-team-kit
ntech-team-kit update
```

If Homebrew tries to upgrade to an old version, refresh the tap first:

```bash
brew update
brew info neronlux/tap/ntech-team-kit
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

The "thermo-nuclear" maintainability audit: 1k-line rule, code-judo moves, spaghetti detection, and ambitious structural simplification. Invokable via `@`.

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
| `fix-merge-conflicts` | Unresolved merge conflicts | Resolve conflicts non-interactively with minimal edits, rebuild lockfiles, check build |
| `get-pr-comments` | PR feedback summary | Fetch review and discussion comments, group by severity and actionability |
| `loop-on-ci` | Watch CI until green | Watch PR checks, fix failures, iterate until all checks pass |
| `make-pr-easy-to-review` | Prepare PR for review | Clean commit history, improve PR description, add reviewer guidance, annotate the diff |
| `new-branch-and-pr` | Start new work | Create a branch, make changes, test, commit, and open a PR |
| `pr-review-canvas` | Interactive PR walkthrough | Generate an HTML page with categorized files, moved-code detection, and inline diffs |
| `review-and-ship` | Ship changes safely | Review for bugs and intent fit, run or write tests, commit, push, and open a PR |
| `run-smoke-tests` | End-to-end verification | Run Playwright smoke tests, debug failures, verify fixes |
| `thermo-nuclear-code-quality-review` | Strict maintainability audit | 1k-line rule, code-judo, spaghetti elimination, ambitious structural simplification |
| `verify-this` | Prove or disprove a claim | Capture baseline and treatment, compare artifacts, return a verdict |
| `weekly-review` | Weekly work summary | Synthesize authored commits into a recap grouped by bugfix, tech debt, and net-new |
| `what-did-i-get-done` | Status update | Summarize your commits over a time range into a concise update |
| `workflow-from-chats` | Extract preferences | Mine recent sessions for durable working preferences, convert into skills or rules |

## Agents

Both agents are invokable via `@agent-name` in OpenCode.

| Agent | Description |
|-------|-------------|
| `ci-watcher` | Background CI monitoring. Polls PR checks and notifies on failure or success. Requires the plugin enabled via `OPENCODE_NTECH_CI_WATCH=1`. |
| `thermo-nuclear-code-quality-review` | Deep maintainability auditor. Gathers its own context (diff + file contents) when invoked directly. Also available as a subagent via `Task`. |

## Commands

Type `/` in OpenCode to see available commands:

| Command | Skill | What it does |
|---------|-------|--------------|
| `/check-compiler-errors` | `check-compiler-errors` | Run compile/type-check and fix failures |
| `/control-cli` | `control-cli` | Drive and inspect an interactive CLI or TUI |
| `/control-ui` | `control-ui` | Drive and inspect a web or Electron UI |
| `/deslop` | `deslop` | Remove AI-generated code slop |
| `/fix-ci` | `fix-ci` | Fix the first failing CI check |
| `/fix-merge-conflicts` | `fix-merge-conflicts` | Resolve merge conflicts |
| `/get-pr-comments` | `get-pr-comments` | Summarize PR review feedback |
| `/loop-on-ci` | `loop-on-ci` | Watch CI and fix until green |
| `/make-pr-easy-to-review` | `make-pr-easy-to-review` | Prepare a PR for review |
| `/new-branch-and-pr` | `new-branch-and-pr` | Start fresh work with a PR |
| `/pr-review-canvas` | `pr-review-canvas` | Generate an interactive HTML PR walkthrough |
| `/review-and-ship` | `review-and-ship` | Review, test, commit, and ship |
| `/run-smoke-tests` | `run-smoke-tests` | Run Playwright smoke tests |
| `/thermo-nuclear-code-quality-review` | `thermo-nuclear-code-quality-review` | Strict maintainability audit |
| `/verify-this` | `verify-this` | Prove or disprove a claim |
| `/weekly-review` | `weekly-review` | Weekly work summary |
| `/what-did-i-get-done` | `what-did-i-get-done` | Status update from commit history |
| `/workflow-from-chats` | `workflow-from-chats` | Extract preferences into skills or rules |

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
- The watcher reaches its polling limit

This is optional — the same functionality is available interactively via `/loop-on-ci` and `/fix-ci`.

## Configuration

The repo ships an `opencode.jsonc` with sensible defaults:

- Auto-loads the two rules via `instructions` glob
- Requires confirmation before loading `thermo-nuclear-code-quality-review` (read-only, but heavy)
- Restricts agent permissions to read-only (`gh` and targeted `git` commands only)

For first-time installs, `ntech-team-kit install` copies this config into `~/.config/opencode/opencode.jsonc` so rules and the plugin load automatically. If you already have an OpenCode config, the installer leaves it untouched; merge the relevant `instructions`, `plugin`, and `permission` entries as needed.

## Architecture

The CLI is a self-contained Go program (`cmd/ntech-team-kit/` + `internal/kit/`):

- **Pure Go** — all commands run natively with no shell script delegation
- **Modular layout** — `main.go` (entry point and dispatch), `interactive.go` (menu and prompts), `args.go` (CLI argument parsing)
- **Interactive mode** — guided menu with numbered component picker, confirmation prompts, and piped-stdin detection
- **Component packs** — install full, lite, agents-only, skills-only, or cherry-pick individual components
- **Partial install/uninstall** — manifest tracks component ownership so you can add or remove components without touching others
- **Kit root resolution** — `NTECH_TEAM_KIT_ROOT` env var, compiled ldflags (Homebrew), or auto-detection from binary location
- **Atomic file writes** — copies content via temp + rename to handle symlink edge cases
- **Atomic manifest** — collects the installed file list in memory and writes it atomically at the end of each install
- **Doctor checks** — validates OpenCode, `gh`, authentication, kit layout, and manifest integrity

## Differences from cursor-team-kit

| Area | Cursor Team Kit | ntech-team-kit (OpenCode) |
|------|-----------------|---------------------------|
| Installation | `/add-plugin` | `brew install ntech-team-kit` or `git clone` + `ntech-team-kit` |
| Plugin system | Cursor plugin manifest | OpenCode skills + agents + TypeScript plugin |
| Background agents | `is_background: true` | TypeScript plugin using OpenCode session events |
| Rules | `.mdc` files with `alwaysApply` | Loaded via `instructions` glob |
| Commands | Not available | First-class `/command` support |
| CLI | Shell script installer | Pure Go binary with interactive setup and component packs |
| Partial install | Not available | Install or uninstall individual components by name |

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
go test ./...                   # Run Go tests (29 tests)
bun run test                    # Full suite: typecheck + build + Go tests
bun run vale                    # Lint documentation prose
```

## License

MIT License

- Original work: Copyright (c) 2026 Cursor
- Port and adaptations: Copyright (c) 2026 Nathan Luxford

Upstream source: [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit)
