# ntech-team-kit

A portable kit of OpenCode skills, agents, commands, and rules for CI, code review, shipping, and test reliability. Forked from [cursor/plugins/cursor-team-kit](https://github.com/cursor/plugins/tree/main/cursor-team-kit) and adapted for [OpenCode](https://opencode.ai).

## Rules

<!-- ntech-team-kit:begin:rules -->

### No inline imports

Always place imports at the top of the module. Avoid inline imports in function bodies, type annotations, or interface fields unless a strict circular-dependency requires it and the exception has a comment explaining why.

### TypeScript exhaustive switch

In switch statements over discriminated unions or enums, use a `never` check in the default case so newly added variants cause compile-time failures until handled.

<!-- ntech-team-kit:end:rules -->
