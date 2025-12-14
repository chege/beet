## Goal

Deliver deterministic instruction generation that accepts free-form intent and produces multiple repo-ready docs (e.g.,
WORK_PROMPT.md, agents.md, PRD, SRS) driven by configurable template packs, while actually invoking Codex or Copilot CLI
to refine the intent into the templates.

## Current State

- Single template, single output (WORK_PROMPT.md) plus optional agents.md.
- CLI execution defaults to on (prefers Codex CLI, falls back to Copilot CLI) and runs with the rendered WORK_PROMPT.md.
- Templates and guidelines are bootstrapped; packs are seeded and listable, but generation remains single-file (no
  multi-output emission).

## Known Gaps

- Pack selection/generation missing: cannot select grouped outputs or emit PRD/SRS/guidelines sets.
- No multi-output loop: only WORK_PROMPT.md (+agents.md) is written; no per-output handling, labelling, or force rules
  beyond agents.md.
- Config/schema gaps: missing-template validation; docs/help do not cover packs/exec behavior.
- Testing gaps: no coverage for pack parsing, multi-file generation, or e2e with packs.

## Required Work

1) Template packs: support selecting packs via `-p/--pack` and render outputs from pack definitions; ship an extended
   pack for PRD/SRS/guidelines.
2) Multi-output render: for each pack output, render intent + guidelines through the specified template; honor
   `--dry-run` (print all), `--force-agents` (only gates agents.md), and fail if a template is missing.
3) Determinism and parity: ensure identical intent and pack yield identical outputs regardless of input source (
   args/file/stdin/editor). Add tests for pack parsing, multi-file writes, dry-run output labelling, force-agents
   behavior, and CLI invocation wiring.
4) Docs/help: update README and `--help` to describe packs, multi-output behavior, CLI shaping, and the new flags.
5) Integration tests: extend e2e to cover pack selection and multi-file creation (with a fake CLI).
6) DX helpers: keep config files editable directly, but add optional helpers (`beet pack init/list/edit`,
    `beet template new`) to scaffold and list packs/templates to reduce user friction and errors.

Last updated: 2025-12-14T17:58:23.088Z
