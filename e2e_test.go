package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2EGenerateCreatesOutputs(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	workdir := t.TempDir()
	configDir := filepath.Join(t.TempDir(), "cfg")
	bin := filepath.Join(t.TempDir(), "beet-e2e")
	cliBinDir := t.TempDir()

	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, string(out))
	}

	cliScript := filepath.Join(cliBinDir, "codex")
	if err := os.WriteFile(cliScript, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}

	cmd := exec.Command(bin, "ship", "it")
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(),
		"BEET_CONFIG_DIR="+configDir,
		"HOME="+workdir,
		"PATH="+cliBinDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("beet execution failed: %v\n%s", err, string(out))
	}

	wp := filepath.Join(workdir, workPromptFilename)
	agents := filepath.Join(workdir, agentsFilename)

	if _, err := os.Stat(wp); err != nil {
		t.Fatalf("WORK_PROMPT.md not created: %v", err)
	}
	if _, err := os.Stat(agents); err != nil {
		t.Fatalf("agents.md not created: %v", err)
	}

	content, err := os.ReadFile(wp)
	if err != nil {
		t.Fatalf("read WORK_PROMPT.md: %v", err)
	}
	if !strings.Contains(string(content), "ship it") {
		t.Fatalf("work prompt missing intent: %s", string(content))
	}
	if !strings.Contains(string(content), "Internal instruction") {
		t.Fatalf("work prompt missing internal instruction: %s", string(content))
	}
}
