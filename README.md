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

Run the CLI with --help to see available commands and options:

```bash
beet --help
```

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
