## Goal

Deliver deterministic instruction generation that accepts free-form intent and produces multiple repo-ready docs (e.g.,
WORK_PROMPT.md, agents.md, PRD, SRS) driven by configurable template packs and ready for downstream LLMs.

## Current State

- Pack-driven multi-output generation (default pack emits WORK_PROMPT.md + agents.md; extended pack adds PRD/SRS/GUIDELINES); packs are bootstrapped and selectable.
- Comprehensive separation pack (AGENTS/INTENT/DESIGN/RULES/PLAN/PROGRESS) is bootstrapped and selectable.
- CLI detection exists only for diagnostics (`beet doctor`) and does not change generation behavior.
- Docs/help describe commands, packs, shaping, and flags.
- Config preflight checks ensure templates/packs exist before generation.
- DX helpers implemented: `beet pack list|init|edit` and `beet template new`.
- Integration tests cover default, extended, and comprehensive pack outputs.

## Known Gaps

- None currently.

## Required Work

None currently.

Last updated: 2025-12-15T08:52:37.143Z
