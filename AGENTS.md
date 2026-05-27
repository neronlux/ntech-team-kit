# ntech-team-kit

A portable kit of OpenCode and Codex skills, agents, commands, and rules for CI, code review, shipping, and test reliability. Forked from [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit) and adapted for [OpenCode](https://opencode.ai) and Codex.

## Rules

### No inline imports

Always place imports at the top of the module. Avoid inline imports in function bodies, type annotations, or interface fields unless a strict circular-dependency requires it and the exception has a comment explaining why.

### TypeScript exhaustive switch

In switch statements over discriminated unions or enums, use a `never` check in the default case so newly added variants cause compile-time failures until handled.

### Every skill must have a command

When adding a new skill to `skills/`, you must also create a matching `commands/<skill-name>.md` and add the name to the `skills` slice in `internal/kit/install.go`. The `commands` slice derives from `skills` (`commands = skills[:]`) so they stay 1:1 automatically. When in this repo, verify the lists match before completing any task that touches skills or install.go.

## Interactive mode

Running `ntech-team-kit` with no arguments opens a guided menu that loops until the user quits. The menu shows version, kit root status, OpenCode installed file count, and Codex installed file count. Target-aware options include full/lite/agents/skills install, custom install, custom uninstall, status, and update with a picker for OpenCode, Codex, both, or auto-detect. OpenCode agents install as markdown files; Codex agents install as generated TOML custom agents. Custom install/uninstall uses a numbered component picker, and uninstall asks for confirmation.

When stdin is not a terminal (piped), the interactive mode runs the default action (full install) and exits without looping. This makes `echo | ntech-team-kit` safe for CI scripts.

### Component selection

The CLI supports installing and uninstalling individual components:

- `--target opencode|codex|both|auto` — install/status/update/uninstall target
- `--pack full|lite|agents|skills` — named packs
- `--only <components>` — cherry-pick components
- `--without <components>` — exclude components from a pack
- `--select` — interactive numbered component picker
- `uninstall --only <components>` — partial uninstall preserving other components

The manifest tracks component ownership (`component\tpath` format). Partial installs merge with existing entries; partial uninstalls update the manifest instead of deleting it. Empty `ComponentSet` means "all" for backward compatibility.

### Codex support

Codex receives target-specific skills in `~/.agents/skills` and generated custom agents in `~/.codex/agents`, with `NTECH_TEAM_KIT_CODEX_SKILLS_DIR` and `NTECH_TEAM_KIT_CODEX_AGENTS_DIR` available as overrides. The installer rewrites Codex-installed skill copies as Codex-compatible, generates skill `agents/openai.yaml`, generates Codex agent TOML from OpenCode markdown agents, and opens the Codex Skills view via `codex://skills` on macOS or Linux; set `NTECH_TEAM_KIT_CODEX_SKIP_APP_REFRESH=1` for CI or headless runs. The same installed skills are available to Codex CLI, IDE, and GUI. Codex custom-agent activity is currently surfaced in Codex CLI and app. The interactive menu and CLI both support `opencode`, `codex`, `both`, and `auto` targets for install, uninstall, update, and status. OpenCode remains the full target for skills, agents, commands, rules, plugin, and config. Doctor auto-detects OpenCode CLI, Codex CLI, and Codex GUI on macOS and Linux, and its install-manifest check passes for OpenCode-only, Codex-only, or both-target installs.

## Release process

`.github/workflows/ci.yml` automates releases from pushes to `main`.

Before pushing release-worthy changes:

- Run `bun run test`.
- Run `bun run vale` when public docs, agents, commands, or rules changed.
- Verify skill and command names are 1:1 when touching `skills/`, `commands/`, or `internal/kit/install.go`.
- For Homebrew-facing changes, confirm `Formula/ntech-team-kit.rb` is the source-of-truth formula shape. The CI dispatch sends a base64-encoded copy of this formula to `neronlux/homebrew-tap`; do not reintroduce hardcoded formula templates in the tap repo.

After pushing to `main`:

- Watch the source CI run for `check`, `release-needed`, and `auto-release`.
- Confirm `auto-release` bumps `VERSION`, creates the `vX.Y.Z` tag and GitHub release, computes the tarball SHA256, dispatches the formula to `neronlux/homebrew-tap`, and updates the reference formula in this repo.
- Pull `main` after the release workflow finishes so local `VERSION`, tags, and `Formula/ntech-team-kit.rb` match the generated release commits.

Homebrew tap verification:

- Confirm `neronlux/homebrew-tap` receives the `repository_dispatch` and updates `Formula/ntech-team-kit.rb` to the new tag and SHA256.
- Confirm the tap `Test Formulae` workflow runs via `workflow_run` and passes on macOS.
- If local `brew upgrade` tries to install an old version, run `brew update` first. A stale local tap checkout can keep Homebrew pinned to an older formula that no longer reflects GitHub.
- On Linux, verify with `brew info neronlux/tap/ntech-team-kit`, then run `brew upgrade neronlux/tap/ntech-team-kit` when testing the published formula.
