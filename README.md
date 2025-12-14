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
- `beet [intent]` — generate WORK_PROMPT.md and agents.md (default template)
- `beet templates` — list available templates
- `beet packs` — list available packs (default pack bootstrapped; generation currently uses the default single-output flow)
- `beet doctor` — show detected CLIs (Codex preferred, Copilot fallback)

Flags:
- `-t, --template <name>` — choose template (single-output flow)
- `--dry-run` — render to stdout without writing files
- `--exec` (default true) — run detected CLI with the rendered WORK_PROMPT.md
- `--force-agents` — allow overwriting agents.md

CLI shaping: execution defaults to on; the detected CLI (Codex first, then Copilot) receives WORK_PROMPT.md. If no CLI is found and `--exec` is true, generation fails.

Packs and multi-output: pack files are bootstrapped and listable, but generation currently writes only WORK_PROMPT.md plus agents.md. Pack-based multi-output generation will extend this to PRD/SRS/guidelines sets.

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
