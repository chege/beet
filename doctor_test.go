package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectPreferredCLIPriority(t *testing.T) {
	binDir := t.TempDir()

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
	if !strings.Contains(out, "codex:") || !strings.Contains(out, "copilot:") {
		t.Fatalf("doctor output missing entries: %s", out)
	}
	if !strings.Contains(out, "No supported CLI") {
		t.Fatalf("doctor should warn when none found: %s", out)
	}
}
