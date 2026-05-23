# ntech-team-kit

A portable kit of OpenCode skills, agents, commands, and rules for CI, code review, shipping, and test reliability. Forked from [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit) and adapted for [OpenCode](https://opencode.ai).

## Rules

### No inline imports

Always place imports at the top of the module. Avoid inline imports in function bodies, type annotations, or interface fields unless a strict circular-dependency requires it and the exception has a comment explaining why.

### TypeScript exhaustive switch

In switch statements over discriminated unions or enums, use a `never` check in the default case so newly added variants cause compile-time failures until handled.

### Every skill must have a command

When adding a new skill to `skills/`, you must also create a matching `commands/<skill-name>.md` and add the name to the `commands` slice in `internal/kit/install.go`. Skills and commands must stay 1:1 so users can invoke every skill via `/command`. When in this repo, verify the lists match before completing any task that touches skills or install.go.
