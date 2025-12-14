# Beet CLI — PRD & SRS
_Last updated: 2025-12-12_

---

## Part I — Product Requirements Document (PRD)

### 1. Purpose

The Beet CLI (`beet`) is a local developer tool that transforms **free‑form human intent**
into **clean, structured, repository‑ready instruction files** for Codex CLI, Copilot CLI, or similar LLM tools.

The output is committed files, not chat history.

Primary outputs:
- `agents.md` — how the AI should work
- `WORK_PROMPT.md` — what the AI should do now

---

### 2. Problem Statement

LLM‑assisted software development breaks down when:
- instructions are ad‑hoc
- prompts drift between runs
- process rules are forgotten
- context lives only in chat sessions

Writing and maintaining large instruction files manually is slow and error‑prone.

There is no simple tool that:
> Takes rough, unstructured intent and reliably produces disciplined instruction files.

---

### 3. Goals

**Primary**
- Accept free‑form, unstructured input
- Refine and clarify intent using an LLM
- Apply consistent structure via templates
- Generate Codex‑ready instruction files

**Secondary**
- Reuse shared coding and process guidelines
- Be fast, local, and deterministic
- Integrate seamlessly with Codex / Copilot CLIs

---

### 4. Non‑Goals

- Not an LLM client
- Not a chat UI
- Not a variable‑substitution template engine
- No cloud services
- No prompt marketplace

---

### 5. Target Users

- Software engineers using Codex CLI or Copilot CLI
- Teams wanting reproducible AI‑assisted workflows
- Developers who prefer writing intent, not schemas

---

### 6. Core Concept

```
free‑form input
  → LLM refinement (clarify, rephrase, structure)
  → template framing
  → guideline constraints
  → instruction files
```

The user never fills template fields.
The LLM reshapes the text.

---

### 7. Managed Files (Codex‑Facing)

```
agents.md        # stable: how Codex should work
WORK_PROMPT.md   # volatile: what Codex should do now
```

### 7.1 File-Set Design (Adopted)

We adopt Proposal 2 (Comprehensive Separation):
- AGENTS.md — entrypoint, lists auxiliary files.
- INTENT.md — structured interpretation of free-form requirements.
- DESIGN.md — optional high-level solution outline when needed.
- RULES.md — coding and quality guidelines.
- PLAN.md — detailed work breakdown.
- PROGRESS.md — optional progress tracker mirroring PLAN tasks.

---

### 8. Key Features (v1)

- Free‑form input via editor, file, or stdin
- Template‑based framing (no variables)
- Shared guideline injection
- Automatic config bootstrap
- Overwrite safety for `agents.md`
- Template discovery
- CLI detection (`doctor`)
- Dry‑run mode
- Optional execution of detected LLM CLI

---

### 9. Success Criteria

- Users stop hand‑writing large instruction files
- Instruction quality is consistent across runs
- Files are committed and reviewable
- Tool becomes a standard pre‑step before Codex/Copilot

---

## Part II — Software Requirements Specification (SRS)

### 1. Scope

This SRS defines the functional and non‑functional requirements for the Beet CLI (`beet`).

---

### 2. Terminology

- **Template** — Instructional frame that defines how intent is reshaped
- **Guideline** — Reusable rule set influencing LLM behavior
- **Instruction files** — Files consumed by Codex/Copilot

---

### 3. Functional Requirements

#### FR‑1 Input Handling
- Accept input from:
  - `$EDITOR`
  - File argument
  - `stdin`
- Input is treated as free‑form natural language

---

#### FR‑2 Internal LLM Instruction
The tool must prepend a fixed internal instruction that directs the LLM to:
- Clarify vague statements
- Rephrase informal language
- Infer reasonable missing details (without inventing features)
- Structure output according to the selected template
- Produce only the final instruction text

---

#### FR‑3 Templates
- Templates are plain text / Markdown
- Templates define output structure, not input fields
- Templates may reference one or more guideline files
- Templates are stored in the config directory

---

#### FR‑4 Guidelines
- Guidelines are plain text / Markdown
- Injected verbatim
- Deterministic order
- Extensible by adding files

---

#### FR‑5 Configuration

Default config directory:
```
~/.beet/
```

Structure:
```
~/.beet/
  templates/
  guidelines/
```

- Bootstrap automatically on first run
- Never overwrite user‑modified files

Override:
```
BEET_CONFIG_DIR
```

---

#### FR‑6 Output Files

- Generate:
  - `agents.md`
  - `WORK_PROMPT.md`
- `agents.md`:
  - Created if missing
  - Overwritten only with `--force-agents`
- `WORK_PROMPT.md`:
  - Always regenerated

---

#### FR‑7 CLI Detection

- Detect available CLIs in priority order:
  1. Codex CLI
  2. Copilot CLI
  3. Claude Code CLI
- Provide transparency via `beet doctor`
- Fail clearly if `--exec` is requested and none are found

---

#### FR‑8 Commands

```
beet [input]
beet -t <template>
beet templates
beet doctor
```

---

#### FR‑9 Flags

- `-t, --template <name>`
- `--dry-run`
- `--exec`
- `--force-agents`

---

### 4. Non‑Functional Requirements

- Language: Go
- Single binary
- Instant startup
- Deterministic output
- Local execution only
- Clear error messages

---

### 5. Safety & Determinism

- No hidden state
- No network access except via LLM CLI
- Same input + template → same output

---

### 6. Exit Conditions

The system is considered complete when:
- Instruction files fully reflect user intent
- Files are ready for Codex/Copilot consumption
- No user interaction is required beyond input

---

## Summary

`beet` is a **prompt‑to‑instruction compiler**.

Humans write intent.
Templates encode discipline.
Guidelines encode standards.
LLMs refine and clarify.
Repositories contain the truth.
