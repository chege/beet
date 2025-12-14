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

Flags:
- `-t, --template <name>` — override the WORK_PROMPT.md template when using the default pack
- `-p, --pack <name>` — select a pack (default: `default`)
- `--dry-run` — render all outputs to stdout with labels
- `--exec` (default true) — send rendered prompts to the detected CLI and write its shaped output
- `--force-agents` — allow overwriting agents.md

CLI shaping: execution defaults to on; the detected CLI (Codex first, then Copilot) receives the full prompt (internal instruction + template + guidelines + intent) on stdin and its output is written to files. If no CLI is found and `--exec` is true, generation fails.

Packs and multi-output: pack files define outputs and templates; all outputs are rendered per pack. The default pack emits WORK_PROMPT.md and agents.md; extended packs (e.g., PRD/SRS/guidelines) can be added to `~/.beet/packs`.
Built-in packs: `default` (WORK_PROMPT.md, agents.md) and `extended` (adds PRD.md, SRS.md, GUIDELINES.md).

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
