package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const guidelineSnippet = "Be clear. Be concise."

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

	expected := map[string]struct {
		template         string
		expectIntent     bool
		expectGuidelines bool
		extra            string
	}{
		"WORK_PROMPT.md": {template: "default", expectIntent: true},
		"agents.md":      {template: "agents", expectGuidelines: true},
		"PRD.md":         {template: "prd", expectIntent: true, expectGuidelines: true, extra: "Product Requirements"},
		"SRS.md":         {template: "srs", expectIntent: true, expectGuidelines: true, extra: "Software Requirements Specification"},
		"GUIDELINES.md":  {template: "guidelines", expectGuidelines: true, extra: "Guidelines"},
	}

	for name, exp := range expected {
		path := filepath.Join(workdir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		content := string(data)
		if !strings.Contains(content, "Internal instruction") {
			t.Fatalf("%s missing internal instruction", name)
		}
		if !strings.Contains(content, "Template: "+exp.template) {
			t.Fatalf("%s missing template label %q", name, exp.template)
		}
		if exp.expectIntent && !strings.Contains(content, "ship it") {
			t.Fatalf("%s missing intent content", name)
		}
		if exp.expectGuidelines && !strings.Contains(content, guidelineSnippet) {
			t.Fatalf("%s missing guidelines content", name)
		}
		if exp.extra != "" && !strings.Contains(content, exp.extra) {
			t.Fatalf("%s missing template text %q", name, exp.extra)
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

	expected := map[string]struct {
		template         string
		expectIntent     bool
		expectGuidelines bool
		extra            string
	}{
		"WORK_PROMPT.md": {template: "default", expectIntent: true},
		"agents.md":      {template: "agents", expectGuidelines: true},
		"INTENT.md":      {template: "intent", expectIntent: true, extra: "Intent"},
		"DESIGN.md":      {template: "design", expectIntent: true, expectGuidelines: true, extra: "Notes"},
		"RULES.md":       {template: "rules", expectGuidelines: true, extra: "Rules"},
		"PLAN.md":        {template: "plan", expectIntent: true, extra: "[ ]"},
		"PROGRESS.md":    {template: "progress", extra: "Not started"},
	}

	for name, exp := range expected {
		path := filepath.Join(workdir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		content := string(data)
		if !strings.Contains(content, "Internal instruction") {
			t.Fatalf("%s missing internal instruction", name)
		}
		if !strings.Contains(content, "Template: "+exp.template) {
			t.Fatalf("%s missing template label %q", name, exp.template)
		}
		if exp.expectIntent && !strings.Contains(content, "ship it") {
			t.Fatalf("%s missing intent content", name)
		}
		if exp.expectGuidelines && !strings.Contains(content, guidelineSnippet) {
			t.Fatalf("%s missing guidelines content", name)
		}
		if exp.extra != "" && !strings.Contains(content, exp.extra) {
			t.Fatalf("%s missing template text %q", name, exp.extra)
		}
	}
}
