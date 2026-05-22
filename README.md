# ntech-team-kit

A portable, production-grade collection of skills, agents, commands, and rules that bring Cursor Team Kit workflows to [OpenCode](https://opencode.ai).

Forked from [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit) and adapted to work natively with OpenCode.

## What is this?

If you like the rigorous internal workflows Cursor uses (strict code review, reliable CI loops, clean shipping processes, etc.) but prefer OpenCode as your AI coding agent, this kit gives you those same capabilities.

It includes:

- **18 reusable skills** for CI, code review, shipping, verification, and code quality
- **2 specialized agents**, both directly tab-selectable via `@`
- **8 convenient `/commands`** for the most common workflows
- **2 opinionated rules** (no inline imports + exhaustive TypeScript switches)
- **1 TypeScript plugin** for proactive CI monitoring

Everything installs locally into your `~/.config/opencode/` directory. A single `ntech-team-kit update` command keeps the CLI and all content (skills, agents, commands) in sync after a `git pull` or `brew upgrade`.

## Prerequisites

Before installing, make sure you have:

- [OpenCode](https://opencode.ai) installed
- [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated (`gh auth login`) — required by most skills
- Git configured with your user email

Some skills (e.g. `run-smoke-tests`, `control-ui`) may also require tools that exist in the project you're working on (Playwright, etc.).

## Installation

### Homebrew (recommended)

```bash
brew tap neronlux/tap
brew install ntech-team-kit
ntech-team-kit install
```

This is the easiest way to install and keep the tool up to date.

### From source (for development)

```bash
git clone https://github.com/neronlux/ntech-team-kit.git ~/ntech-team-kit
cd ~/ntech-team-kit

# Convenient launcher (no build step needed)
./bin/ntech-team-kit install

# Or build the Go binary
go build -o /usr/local/bin/ntech-team-kit ./cmd/ntech-team-kit
ntech-team-kit install
```

You can also use the original shell script directly:

```bash
./install.sh
```

### CLI reference

```bash
ntech-team-kit install                 # Install / refresh skills, agents, commands, rules
ntech-team-kit update                  # Check for updates + refresh all content (recommended)
ntech-team-kit doctor                  # Health check + daily update hint
ntech-team-kit status                  # Show what is currently installed
ntech-team-kit version                 # Print the CLI version
ntech-team-kit path                    # Show the resolved kit root
ntech-team-kit uninstall
```

Global options:
- `--root <path>` — override kit location (useful in development)
- `NTECH_TEAM_KIT_ROOT` and `OPENCODE_CONFIG_DIR` environment variables also work

### Keeping the kit up to date

**Homebrew users (recommended):**

```bash
brew upgrade ntech-team-kit
ntech-team-kit update          # optional but recommended — refreshes latest skills/agents
```

**Source / development users:**

```bash
cd ~/ntech-team-kit
git pull
ntech-team-kit update
```

`ntech-team-kit update` does two things:
- Reports whether a newer CLI binary is available on GitHub
- Always copies the latest skills, agents, commands, and rules from the current kit tree into your OpenCode config

`ntech-team-kit doctor` will also print a polite one-line hint (at most once per day) when it detects a newer version.

After any `git pull` or `brew upgrade`, running `update` (or just `doctor`) is the fastest way to stay current.

See the full CLI reference in the Installation section above. Homebrew ships the exact same binary.

## Quick Start

Here are the most common ways people use the kit:

### 1. Review and ship changes

After finishing work on a branch:

```
/review-and-ship
```

This runs a structured review, suggests or writes tests, commits cleanly, and opens/updates a PR.

### 2. Fix failing CI

When your PR has red checks:

```
/loop-on-ci
```

The agent will watch the checks, diagnose failures using `gh pr checks`, apply fixes, and iterate until everything is green.

### 3. Get a strict code quality review

```bash
@thermo-nuclear-code-quality-review review the current branch for maintainability issues
```

Or from a parent agent, invoke it as a subagent with `Task(subagent_type: "thermo-nuclear-code-quality-review", ...)`.

This is the famous "thermo-nuclear" maintainability audit (1k-line rule, code judo, spaghetti detection, ambitious structural simplification). The agent is fully tab-discoverable via `@` and gathers its own context (diff + file contents) when invoked directly.

### 4. Verify a claim with evidence

```
/verify-this The new retry logic is 3x faster than before
```

Captures baseline vs treatment and returns a clear `VERIFIED` / `NOT VERIFIED` / `INCONCLUSIVE` verdict.

## Components

### Skills

Skills load on demand via the `skill` tool. The most popular ones include:

| Skill                              | Purpose |
|------------------------------------|---------|
| `review-and-ship`                  | Full review → test → commit → PR workflow |
| `loop-on-ci`                       | Watch PR checks and fix failures until green |
| `fix-ci`                           | Diagnose and fix the first failing check |
| `thermo-nuclear-code-quality-review` | Deep maintainability audit (1k-line rule, ambitious code-judo, no spaghetti) |
| `verify-this`                      | Prove or disprove a claim with local evidence |
| `make-pr-easy-to-review`           | Clean commit history and improve PR description |
| `pr-review-canvas`                 | Generate a beautiful interactive HTML PR review page |
| `control-cli` / `control-ui`       | Build local harnesses to drive CLIs or UIs |
| `weekly-review` / `what-did-i-get-done` | Summarize your recent work |

See the full list in the `skills/` directory.

### Agents

Both agents are directly discoverable via `@` in OpenCode (press Tab after `@`).

| Agent                              | Description |
|------------------------------------|-------------|
| `ci-watcher`                       | Background CI monitoring agent (requires the plugin below) |
| `thermo-nuclear-code-quality-review` | Deep, rigorous maintainability auditor (1k-line rule, code-judo, spaghetti elimination). Fully tab-selectable via `@` and gathers its own context when invoked directly. |

### Commands

These are the most frequently used commands. Type `/` in OpenCode to see them:

- `/review-and-ship`
- `/loop-on-ci`
- `/verify-this`
- `/run-smoke-tests`
- `/fix-ci`
- `/new-branch-and-pr`
- `/make-pr-easy-to-review`
- `/fix-merge-conflicts`

**Note:** Commands are thin, discoverable wrappers around the underlying skills. You can also invoke skills directly via the `skill` tool or by mentioning agents with `@`.

### Rules

Two rules are automatically loaded into every OpenCode session:

- **No inline imports** — Imports must stay at the top of the file
- **TypeScript exhaustive switch** — Use `never` in default cases for discriminated unions

These rules are also available in `AGENTS.md` (for projects that want to commit them).

### Plugin: Background CI Watcher

The `ci-watcher` plugin gives you proactive CI monitoring without having to ask.

Enable it with:

```bash
export OPENCODE_NTECH_CI_WATCH=1
```

When a session becomes idle, the plugin will poll your PR checks in the background and notify you of failures or success.

The plugin is optional — all the same functionality is available interactively via `/loop-on-ci` and `/fix-ci`.

## Configuration

An example `opencode.jsonc` ships with the repo. It configures:

- Automatic loading of the two rules
- Safe permissions for the heavy `thermo-nuclear` review skill (asks before loading)
- Read-only permissions for the CI watcher and code review agents

Merge the relevant parts into your own `~/.config/opencode/opencode.json` if desired.

## Differences from cursor-team-kit

| Area                    | Cursor Team Kit                  | ntech-team-kit (OpenCode)                  |
|-------------------------|----------------------------------|--------------------------------------------|
| Installation            | `/add-plugin`                    | `git clone` + `./install.sh`               |
| Plugin system           | Cursor plugin manifest           | OpenCode skills + agents + local plugin    |
| Background agents       | `is_background: true`            | Real TypeScript plugin using session events|
| Rules                   | `.mdc` files with `alwaysApply`  | Loaded via `instructions` glob + `AGENTS.md` |
| Commands                | Not available                    | First-class `/command` support             |

## Development

This kit includes automated validation for the TypeScript plugin and documentation.

### Requirements

- [Bun](https://bun.sh) (for type checking and building the plugin)

### Commands

```bash
bun install          # Install dev dependencies
bun run test         # Type-check + build the plugin
bun run typecheck    # TypeScript only
bun run vale         # Lint documentation
```

The test suite exists so the plugin can be properly validated with Bun (addressing the common situation where the machine running the review didn't have Bun installed).

## License

MIT License

- Original work: Copyright (c) 2026 Cursor
- Port and adaptations: Copyright (c) 2026 Nathan Luxford

Upstream source: [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit)
