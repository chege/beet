# 🫜 beet

A lightweight CLI for project task automation, developer tooling, and workflows written in Go.

## 🚀 Quickstart

- Install Go 1.25.x
- Run tests and lint locally:

```bash
go test ./...
golangci-lint run
```

## 📥 Install

There are two common ways to install the CLI:

- From the published module (recommended):

```bash
go install github.com/chege/beet@latest
```

- From local source (installs the built binary into your $GOBIN or $GOPATH/bin):

```bash
cd /path/to/beet
go install ./...
```

Alternatively, build and move the binary to a directory in your PATH:

```bash
go build -o beet ./...
sudo mv beet /usr/local/bin/
```

## 💡 Usage

Key commands:
- `beet [intent]` — generate pack outputs (default pack emits WORK_PROMPT.md + agents.md)
- `beet -p <pack> [intent]` — use a specific pack from `~/.beet/packs` (e.g., `extended`)
- `beet templates` — list available templates
- `beet packs` — list available packs (default pack bootstrapped)
- `beet doctor` — show detected CLIs (Codex preferred, Copilot fallback)
- `beet pack list|init|edit` — list or scaffold pack files in your config dir
- `beet template new <name>` — scaffold a new template in your config dir
- `beet config restore` — recopy bundled defaults into your config directory without overwriting existing files


Flags:
- `-t, --template <name>` — override the WORK_PROMPT.md template when using the default pack
- `-p, --pack <name>` — select a pack (default: `default`)
- `--dry-run` — render all outputs to stdout with labels
- `--force-agents` — allow overwriting agents.md
- `-v, --verbose` — enable verbose diagnostics (config bootstrap, pack/template selection, and rendering) written to stderr
## ⚙️ Environment

- `BEET_CONFIG_DIR` — override the default `~/.beet` directory when bootstrapping templates, guidelines, and packs.
- `BEET_CLI_PATH` — point beet at a specific CLI binary (useful for wrappers or alternative installs); `beet doctor` surfaces whether the override resolved.
- `BEET_CLI_TIMEOUT` — change how long beet waits for the detected CLI before aborting (duration syntax, default `5m`).

## 📝 Logging

`beet` prints only fatal errors unless the verbose flag is enabled. Pass `-v` or `--verbose` to stream diagnostics to stderr covering configuration bootstrapping, pack/template discovery, and prompt rendering; your generated files remain untouched.

Beet now focuses on prompt generation: after rendering the selected pack/template it writes each output (WORK_PROMPT.md, agents.md, etc.) to disk. Execution against a downstream LLM is left to you, so `beet` stays deterministic even without Codex or Copilot.

Packs and multi-output: pack files define outputs and templates; all outputs are rendered per pack. The default pack emits WORK_PROMPT.md and agents.md; extended packs (e.g., PRD/SRS/guidelines) and comprehensive packs (AGENTS/INTENT/DESIGN/RULES/PLAN/PROGRESS) can be added to `~/.beet/packs`.
Built-in packs: `default` (WORK_PROMPT.md, agents.md), `extended` (adds PRD.md, SRS.md, GUIDELINES.md), and `comprehensive` (adds INTENT.md, DESIGN.md, RULES.md, PLAN.md, PROGRESS.md).

Defaults: bundled templates, guidelines, and pack files live under `defaults/` in the repo. On first run Beet copies these into your config directory (`~/.beet` by default) without overwriting existing files; run `beet config restore` to re-copy any missing defaults later.

## 🧩 Template packs & placeholders (for custom templates)

When creating your own pack templates, these global placeholders are available (designed for Copilot/Codex-facing prompts and personal projects):

- `{{intent}}` – the raw goal or task.
- `{{background}}` – any repo/project context the model should know.
- `{{goals}}` – the outcomes you want.
- `{{requirements}}` – must-haves or constraints to honor.
- `{{assumptions}}` – what you’re presuming is true.
- `{{constraints}}` – limits like time/scope/resources.
- `{{risks}}` – concerns and mitigations worth calling out.
- `{{deliverables}}` – files/artifacts expected.
- `{{acceptance_criteria}}` – how success is judged.
- `{{guidelines}}` – style/ops rules to follow.
- `{{open_questions}}` – unknowns to resolve.

## ⚙️ CI

The repository uses a GitHub Actions workflow (CI) that runs tests and golangci-lint. The CI supports manual runs via the workflow_dispatch trigger.

## 📦 Dependencies

Dependabot is enabled to update Go modules and GitHub Actions.

## 🤝 Contributing

Please open issues or PRs. Follow commit message conventions: `type(scope): subject`.
