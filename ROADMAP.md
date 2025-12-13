## Goal

Deliver deterministic instruction generation that accepts free-form intent and produces multiple repo-ready docs (e.g.,
WORK_PROMPT.md, agents.md, PRD, SRS) driven by configurable template packs, while actually invoking Codex or Copilot CLI
to refine the intent into the templates.

## Current State

- Single template, single output (WORK_PROMPT.md) plus optional agents.md.
- Deterministic string substitution only; Codex/Copilot is invoked only when `--exec` is provided and then receives the
  already-rendered WORK_PROMPT.md.
- Templates and guidelines are bootstrapped but there is no notion of packs or multi-file emission.

## Known Gaps

- No pack support: cannot select grouped outputs or emit PRD/SRS/guidelines sets.
- No default CLI shaping: Codex/Copilot does not shape intent unless `--exec`, and even then only receives
  WORK_PROMPT.md.
- No multi-output loop: only WORK_PROMPT.md (+agents.md) is written; no per-output handling, labelling, or force rules
  beyond agents.md.
- Config/schema gaps: no pack definitions, missing-template validation, or bootstrap for pack files; docs/help do not
  cover packs/exec behavior.
- Testing gaps: no coverage for pack parsing, multi-file generation, CLI invocation wiring, or e2e with packs.

## Required Work

1) Template packs: add `~/.beet/packs/<pack>.yaml` describing outputs (file name + template file). Default pack should
   cover WORK_PROMPT.md + agents.md; ship an extended pack for PRD/SRS/guidelines. Add `beet -p/--pack` and `beet packs`
   to list available packs.
2) Multi-output render: for each pack output, render intent + guidelines through the specified template; honor
   `--dry-run` (print all), `--force-agents` (only gates agents.md), and fail if a template is missing.
3) CLI integration for shaping: introduce a mode where the rendered prompt (internal instruction + template +
   guidelines + intent) is sent to the detected CLI (Codex preferred, Copilot fallback) to produce the final file
   content. Make this the default path for generation (configurable to allow offline deterministic mode). Handle clear
   errors when no CLI is present.
4) Determinism and parity: ensure identical intent and pack yield identical outputs regardless of input source (
   args/file/stdin/editor). Add tests for pack parsing, multi-file writes, dry-run output labelling, force-agents
   behavior, and CLI invocation wiring.
5) Docs/help: update README and `--help` to describe packs, multi-output behavior, CLI shaping, and the new flags.
6) Safety/bootstrapping: bootstrap default pack files and guard against overwriting user templates/guidelines; keep
   existing overwrite rules (agents.md requires `--force-agents`).
7) Integration tests: extend e2e to cover pack selection, multi-file creation, and Codex/Copilot execution wiring (with
   a fake CLI).
8) DX helpers: keep config files editable directly, but add optional helpers (`beet pack init/list/edit`,
   `beet template new`) to scaffold and list packs/templates to reduce user friction and errors.
