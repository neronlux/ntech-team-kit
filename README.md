<p align="center">
  <img src="https://assets.ntek.app/teamkitcodex.png" alt="ntech-team-kit" width="400" />
</p>

# ntech-team-kit

A portable, production-grade collection of skills, agents, commands, and rules that bring [Cursor Team Kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit) workflows to [OpenCode](https://opencode.ai) and [Codex](https://developers.openai.com/codex).

If you like rigorous internal workflows for code review, CI loops, and clean shipping processes, this kit makes them easy to install for OpenCode and Codex.

## What's included

| Component | Count | Description |
|-----------|-------|-------------|
| Skills | 18 | On-demand workflows for CI, code review, shipping, verification, and code quality; installable for OpenCode or Codex |
| Agents | 2 | Specialized subagents; OpenCode gets markdown agents, Codex gets generated custom-agent TOML |
| Commands | 18 | OpenCode `/command` shortcuts for every skill (1:1 with skills) |
| Rules | 2 | OpenCode coding standards (no inline imports, exhaustive TypeScript switches) |
| Plugin | 1 | OpenCode background CI watcher for proactive PR monitoring |

OpenCode content installs into `~/.config/opencode/`. Codex skills install into `~/.agents/skills/` and Codex custom agents install into `~/.codex/agents/`.

## AI harness support

| Harness | Installer target | What users get | How to invoke |
|---------|------------------|----------------|---------------|
| OpenCode CLI/TUI | `--target opencode` | Skills, markdown agents, slash commands, rules, config, and the optional CI watcher plugin | `/review-and-ship`, `/loop-on-ci`, `@thermo-nuclear-code-quality-review` |
| Codex CLI | `--target codex` | Codex-compatible skill copies plus generated custom-agent TOML | `$review-and-ship`, `/skills`, or ask Codex to use a named custom agent |
| Codex app | `--target codex` | The same skills and custom agents, with generated `agents/openai.yaml` metadata for the Skills view | Type `$` in the composer or pick enabled skills from the slash command list |
| Codex IDE extension | `--target codex` | The same shared Codex skills in `~/.agents/skills` | Use Codex skill invocation from the extension; the installer also places generated agents in the shared Codex location for harnesses that surface custom agents |

The installer keeps one source tree and emits the right shape for each harness. OpenCode gets native markdown skills, markdown agents, commands, rules, config, and the optional plugin. Codex gets copied skill directories with `compatibility: codex`, generated Skills-view metadata, Codex-specific wording where source guidance would otherwise be OpenCode-only, and TOML custom agents with conservative read-only defaults.

## Prerequisites

- [OpenCode](https://opencode.ai) or [Codex](https://developers.openai.com/codex) installed
- [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated (`gh auth login`) for GitHub-dependent workflows such as CI, PR review, and shipping
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
ntech-team-kit status [--target ...]   Show installed files and manifest status
ntech-team-kit version                 Print the CLI version
ntech-team-kit path                    Print the resolved kit root directory
```

### Interactive mode

Running `ntech-team-kit` with no arguments opens a guided menu:

```
  ntech-team-kit 0.1.31
  ─────────────────────────────
  Kit root: /opt/homebrew/Cellar/ntech-team-kit/0.1.31
  OpenCode installed: 44 files
  Codex installed: nothing
  ─────────────────────────────

  What would you like to do?

    1) Install full pack (choose target)
    2) Install lite pack (choose target)
    3) Install agents only (choose target)
    4) Install skills only (choose target)
    5) Custom install  (choose target/components)
    6) Custom uninstall (choose target/components)
    7) Check status (choose target)
    8) Run doctor
    9) Update (choose target)
    0) Quit
```

Target-aware actions ask whether to use OpenCode, Codex, both, or auto-detection. Codex installs skills and generated custom agents for the Codex CLI, IDE extension, and app; OpenCode remains the target for commands, rules, plugin, and config. The menu loops so you can run more than one action per session. After each action it asks whether to continue or exit.

**Custom install/uninstall** shows a numbered component picker. OpenCode exposes every component:

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

Codex exposes the components it can consume directly:

```
  Select Codex components to install.

    [*] 1) skills     Codex skills in ~/.agents/skills
    [*] 2) agents     Codex custom agents in ~/.codex/agents

  Components:
```

For Codex, `full` means skills plus generated custom agents, `skills` installs only skills, and `agents` installs only generated custom agents.

When you pipe stdin (not a terminal), the interactive mode runs the default action (full install) and exits, so it works in CI scripts:

```bash
echo | ntech-team-kit    # equivalent to: ntech-team-kit install
```

### Non-interactive options

**Global flags** go before the command: `ntech-team-kit --root ./kit install --pack lite`

- `--root <path>` — override kit root location
- `NTECH_TEAM_KIT_ROOT` — environment variable form of `--root`
- `OPENCODE_CONFIG_DIR` — override `~/.config/opencode` location
- `NTECH_TEAM_KIT_CODEX_SKILLS_DIR` — override the Codex skill destination, defaulting to `~/.agents/skills`
- `NTECH_TEAM_KIT_CODEX_AGENTS_DIR` — override the Codex agent destination, defaulting to `~/.codex/agents`
- `NTECH_TEAM_KIT_CODEX_SKIP_APP_REFRESH=1` — skip opening the Codex Skills view after Codex installs

**Install options:**

- `--target opencode|codex|both|auto` — target for install/status/update/uninstall, default: `opencode`
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
ntech-team-kit install --target codex
ntech-team-kit install --target both
ntech-team-kit install --target auto
ntech-team-kit install --select
ntech-team-kit install --only skills,commands
ntech-team-kit install --without plugin,agents
ntech-team-kit status --target codex
ntech-team-kit uninstall --select
ntech-team-kit uninstall --target codex
ntech-team-kit uninstall --only agents
```

Codex receives target-specific `skills` and `agents`. OpenCode receives the full component model: skills, agents, commands, rules, plugin, and config. Codex ignores OpenCode-only commands, rules, plugin, and config.

When installing for Codex, the installer copies each skill to `~/.agents/skills/<skill-name>/`, rewrites the installed copy as Codex-compatible, applies small Codex-specific text adaptations where needed, adds extra assets and generated `agents/openai.yaml` metadata, generates Codex custom agents in `~/.codex/agents/*.toml`, and opens the Codex Skills view via `codex://skills` on macOS or Linux. If Codex does not show a new skill immediately, restart the Codex harness.

Codex users invoke skills with `$skill-name` in the composer or from `/skills`. For example, use `$verify-this` to prove a claim with baseline/treatment evidence, `$review-and-ship` to review the current branch, or `$workflow-from-chats` to extract durable preferences from recent Codex or OpenCode sessions. OpenCode users keep using the matching `/command` shortcuts and `@agent-name` markdown agents.

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

`ntech-team-kit update` checks GitHub for a newer CLI version and refreshes content from the current kit tree. Use `ntech-team-kit update --target codex` or `--target both` to refresh Codex skills too.

`ntech-team-kit doctor` detects OpenCode, Codex CLI, and Codex GUI installs on macOS and Linux, reports OpenCode and Codex install manifests, warns about missing GitHub CLI/auth for GitHub-dependent skills, then prints a one-line update hint at most once per day.

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

The "thermo-nuclear" maintainability audit: 1k-line rule, code-judo moves, spaghetti detection, and ambitious structural simplification. In OpenCode, invoke it with `@thermo-nuclear-code-quality-review` or `/thermo-nuclear-code-quality-review`. In Codex, invoke the skill with `$thermo-nuclear-code-quality-review` or ask Codex to use the generated custom agent for a dedicated maintainability review.

### Verify a claim with evidence

```
/verify-this The new retry logic handles timeouts correctly
```

Captures baseline vs treatment, returns `VERIFIED`, `NOT VERIFIED`, or `INCONCLUSIVE`.

In Codex, invoke the same workflow with `$verify-this`.

### Start fresh work with a PR

```
/new-branch-and-pr
```

Creates a descriptive branch from latest main, completes implementation, commits, and opens a PR.

## Skills

Skills load on demand when invoked via the OpenCode `skill` tool, an OpenCode `/command`, or Codex skill invocation such as `/skills` or `$skill-name`.

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

OpenCode receives markdown agents that are invokable via `@agent-name`. Codex receives generated TOML custom agents in `~/.codex/agents`; ask Codex to use the named custom agent when you want a dedicated subagent run. Codex currently surfaces custom-agent activity in the CLI and app.

| Agent | Description |
|-------|-------------|
| `ci-watcher` | Background CI monitoring. Polls PR checks and notifies on failure or success. Requires the plugin enabled via `OPENCODE_NTECH_CI_WATCH=1`. |
| `thermo-nuclear-code-quality-review` | Deep maintainability auditor. Gathers its own context (diff + file contents) when invoked directly. Also available as a subagent via `Task`. |

## Commands

Commands are OpenCode-only. Type `/` in OpenCode to see available commands. Codex users should invoke the matching skill with `$skill-name` or from `/skills`.

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
- **Interactive mode** — guided target picker, numbered component picker, confirmation prompts, and piped-stdin detection
- **Component packs** — install full, lite, agents-only, skills-only, or cherry-pick individual components
- **Partial install/uninstall** — manifest tracks component ownership so you can add or remove components without touching others
- **Kit root resolution** — `NTECH_TEAM_KIT_ROOT` env var, compiled ldflags (Homebrew), or auto-detection from binary location
- **Atomic file writes** — copies content via temp + rename to handle symlink edge cases
- **Atomic manifest** — collects the installed file list in memory and writes it atomically at the end of each install
- **Doctor checks** — validates OpenCode, Codex CLI/GUI, kit layout, and OpenCode/Codex manifest integrity; missing `gh` auth is a warning for GitHub-dependent skills

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
go test ./...                   # Run Go tests
bun run test                    # Full suite: typecheck + build + Go tests
bun run vale                    # Lint documentation prose
```

## License

MIT License

- Original work: Copyright (c) 2026 Cursor
- Port and adaptations: Copyright (c) 2026 Nathan Luxford

Upstream source: [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit)
