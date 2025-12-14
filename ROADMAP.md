## Goal

Deliver deterministic instruction generation that accepts free-form intent and produces multiple repo-ready docs (e.g.,
WORK_PROMPT.md, agents.md, PRD, SRS) driven by configurable template packs, while actually invoking Codex or Copilot CLI
to refine the intent into the templates.

## Current State

- Pack-driven multi-output generation (default pack emits WORK_PROMPT.md + agents.md; extended pack adds PRD/SRS/GUIDELINES); packs are bootstrapped and selectable.
- Comprehensive separation pack (AGENTS/INTENT/DESIGN/RULES/PLAN/PROGRESS) is bootstrapped and selectable.
- CLI execution defaults to on (prefers Codex CLI, then Copilot) and receives the full prompt on stdin; offline mode via
  `--exec=false`.
- Docs/help describe commands, packs, shaping, and flags.
- Config preflight checks ensure templates/packs exist before generation.

## Known Gaps

- Integration coverage for richer templates is minimal.
- DX helpers: pack/template scaffolding commands are not implemented.

## Required Work

1) Add integration tests for richer templates across packs.
2) DX helpers: `beet pack init/list/edit`, `beet template new` to reduce friction.

Last updated: 2025-12-14T20:55:30.435Z
