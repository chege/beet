package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectPreferredCLIPriority(t *testing.T) {
	binDir := t.TempDir()

	claude := filepath.Join(binDir, "claude")
	if err := os.WriteFile(claude, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write claude: %v", err)
	}
	copilot := filepath.Join(binDir, "copilot")
	if err := os.WriteFile(copilot, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write copilot: %v", err)
	}
	codex := filepath.Join(binDir, "codex")
	if err := os.WriteFile(codex, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write codex: %v", err)
	}

	t.Setenv("PATH", binDir)

	cli, ok := detectPreferredCLI()
	if !ok {
		t.Fatalf("expected CLI detected")
	}
	if cli.name != "codex" {
		t.Fatalf("expected codex preferred, got %s", cli.name)
	}
}

func TestRunDoctorReportsMissing(t *testing.T) {
	t.Setenv("PATH", "")

	var b strings.Builder
	if err := runDoctor(&b); err != nil {
		t.Fatalf("runDoctor error: %v", err)
	}

	out := b.String()
	if !strings.Contains(out, "codex:") || !strings.Contains(out, "copilot:") || !strings.Contains(out, "claude:") {
		t.Fatalf("doctor output missing entries: %s", out)
	}
	if !strings.Contains(out, "No supported CLI") {
		t.Fatalf("doctor should warn when none found: %s", out)
	}
}

func TestRequireCLIRespectsOverride(t *testing.T) {
	binDir := t.TempDir()
	override := filepath.Join(binDir, "custom-cli")
	if err := os.WriteFile(override, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write override: %v", err)
	}

	t.Setenv(envCLIBinary, override)

	cli, err := requireCLI()
	if err != nil {
		t.Fatalf("requireCLI error: %v", err)
	}
	if cli.path != override {
		t.Fatalf("expected override path %s, got %s", override, cli.path)
	}
}

func TestRequireCLIOverrideMissing(t *testing.T) {
	t.Setenv(envCLIBinary, "/nonexistent/cli")

	if _, err := requireCLI(); err == nil {
		t.Fatalf("expected error when override missing")
	} else if !strings.Contains(err.Error(), envCLIBinary) {
		t.Fatalf("error should mention %s, got %v", envCLIBinary, err)
	}
}
