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

## ⚙️ CI

The repository uses a GitHub Actions workflow (CI) that runs tests and golangci-lint. The CI supports manual runs via the workflow_dispatch trigger.

## 📦 Dependencies

Dependabot is enabled to update Go modules and GitHub Actions.

## 🤝 Contributing

Please open issues or PRs. Follow commit message conventions: `type(scope): subject`.
