# ntech-team-kit

A portable kit of OpenCode skills, agents, commands, and rules for CI, code review, shipping, and test reliability. Forked from [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit) and adapted for [OpenCode](https://opencode.ai).

## Rules

### No inline imports

Always place imports at the top of the module. Avoid inline imports in function bodies, type annotations, or interface fields unless a strict circular-dependency requires it and the exception has a comment explaining why.

### TypeScript exhaustive switch

In switch statements over discriminated unions or enums, use a `never` check in the default case so newly added variants cause compile-time failures until handled.

### Every skill must have a command

When adding a new skill to `skills/`, you must also create a matching `commands/<skill-name>.md` and add the name to the `commands` slice in `internal/kit/install.go`. Skills and commands must stay 1:1 so users can invoke every skill via `/command`. When in this repo, verify the lists match before completing any task that touches skills or install.go.

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
