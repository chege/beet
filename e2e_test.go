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
	if err := os.WriteFile(cliScript, []byte("#!/bin/sh\n/bin/cat\n"), 0o755); err != nil {
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

func TestE2EGenerateExtendedPack(t *testing.T) {
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
	if err := os.WriteFile(cliScript, []byte("#!/bin/sh\n/bin/cat\n"), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}

	cmd := exec.Command(bin, "--pack", "extended", "ship", "it")
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(),
		"BEET_CONFIG_DIR="+configDir,
		"HOME="+workdir,
		"PATH="+cliBinDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("beet execution failed: %v\n%s", err, string(out))
	}

	for _, name := range []string{"WORK_PROMPT.md", "agents.md", "PRD.md", "SRS.md", "GUIDELINES.md"} {
		if _, err := os.Stat(filepath.Join(workdir, name)); err != nil {
			t.Fatalf("%s not created: %v", name, err)
		}
	}
}

func TestE2EGenerateComprehensivePack(t *testing.T) {
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
	if err := os.WriteFile(cliScript, []byte("#!/bin/sh\n/bin/cat\n"), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}

	cmd := exec.Command(bin, "--pack", "comprehensive", "ship", "it")
	cmd.Dir = workdir
	cmd.Env = append(os.Environ(),
		"BEET_CONFIG_DIR="+configDir,
		"HOME="+workdir,
		"PATH="+cliBinDir+string(os.PathListSeparator)+os.Getenv("PATH"),
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("beet execution failed: %v\n%s", err, string(out))
	}

	files := []string{"WORK_PROMPT.md", "agents.md", "INTENT.md", "DESIGN.md", "RULES.md", "PLAN.md", "PROGRESS.md"}
	for _, name := range files {
		path := filepath.Join(workdir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		content := string(data)
		if !strings.Contains(content, "Internal instruction") {
			t.Fatalf("%s missing internal instruction", name)
		}
		switch name {
		case "WORK_PROMPT.md", "INTENT.md", "DESIGN.md", "PLAN.md":
			if !strings.Contains(content, "ship it") {
				t.Fatalf("%s missing intent", name)
			}
		}
	}
}
