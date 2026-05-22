# ntech-team-kit

A portable, production-grade collection of skills, agents, commands, and rules that bring Cursor Team Kit workflows to [OpenCode](https://opencode.ai).

Forked from [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit) and adapted to work natively with OpenCode.

## What is this?

If you like the rigorous internal workflows Cursor uses (strict code review, reliable CI loops, clean shipping processes, etc.) but prefer OpenCode as your AI coding agent, this kit gives you those same capabilities.

It includes:

- **18 reusable skills** for CI, code review, shipping, verification, and code quality
- **2 specialized agents** (including a background CI watcher)
- **8 convenient `/commands`** for the most common workflows
- **2 opinionated rules** (no inline imports + exhaustive TypeScript switches)
- **1 TypeScript plugin** for proactive CI monitoring

Everything installs locally into your `~/.config/opencode/` directory and works across machines.

## Prerequisites

Before installing, make sure you have:

- [OpenCode](https://opencode.ai) installed
- [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated (`gh auth login`) — required by most skills
- Git configured with your user email

Some skills (e.g. `run-smoke-tests`, `control-ui`) may also require tools that exist in the project you're working on (Playwright, etc.).

## Installation

```bash
git clone https://github.com/neronlux/ntech-team-kit.git ~/ntech-team-kit
cd ~/ntech-team-kit
./install.sh
```

This creates symlinks from the repo into `~/.config/opencode/`. All skills, agents, commands, and rules become immediately available in OpenCode.

### Install options

| Command / Flag     | Description                              |
|--------------------|------------------------------------------|
| `./install.sh`     | Install using symlinks (recommended)     |
| `./install.sh --copy` | Copy files instead of symlinking      |
| `./install.sh --dry-run` | Preview planned changes               |
| `./install.sh status` | Show currently installed files       |
| `./install.sh uninstall` | Remove everything this kit installed |

### Upgrade

```bash
cd ~/ntech-team-kit && git pull && ./install.sh
```

Because the default mode uses symlinks, pulling the latest version is normally all you need.

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
# In OpenCode, ask the agent to invoke the heavy reviewer
@thermo-nuclear-code-quality-review
```

Or invoke it via the Task tool after gathering a diff. This runs the famous "thermo-nuclear" maintainability audit (1k-line rule, code judo, spaghetti detection, etc.).

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
| `thermo-nuclear-code-quality-review` | Strict maintainability audit |
| `verify-this`                      | Prove or disprove a claim with local evidence |
| `make-pr-easy-to-review`           | Clean commit history and improve PR description |
| `pr-review-canvas`                 | Generate a beautiful interactive HTML PR review page |
| `control-cli` / `control-ui`       | Build local harnesses to drive CLIs or UIs |
| `weekly-review` / `what-did-i-get-done` | Summarize your recent work |

See the full list in the `skills/` directory.

### Agents

| Agent                              | Description |
|------------------------------------|-------------|
| `ci-watcher`                       | Background agent that monitors PR checks (requires the plugin below) |
| `thermo-nuclear-code-quality-review` | Hidden subagent used for deep code quality reviews |

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
