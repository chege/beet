package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlePackList(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(configDir); err != nil {
		t.Fatalf("ensureConfigStructure: %v", err)
	}
	if err := bootstrapDefaults(configDir); err != nil {
		t.Fatalf("bootstrapDefaults: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	if err := handlePackCommand(configDir, []string{"list"}); err != nil {
		t.Fatalf("handlePackCommand list: %v", err)
	}
	_ = w.Close()
	os.Stdout = origStdout

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, "default.yaml") {
		t.Fatalf("list output missing default.yaml: %s", out)
	}
}

func TestHandlePackInitAndEdit(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(configDir); err != nil {
		t.Fatalf("ensureConfigStructure: %v", err)
	}

	if err := handlePackCommand(configDir, []string{"init", "--name", "custom"}); err != nil {
		t.Fatalf("pack init: %v", err)
	}

	path := filepath.Join(configDir, packsDirName, "custom.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read custom pack: %v", err)
	}
	if !strings.Contains(string(data), "outputs:") {
		t.Fatalf("custom pack missing scaffold content: %s", string(data))
	}

	script := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho \"edited\" > \"$1\"\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	t.Setenv("EDITOR", script)

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	defer devNull.Close()
	origStdin := os.Stdin
	os.Stdin = devNull
	defer func() { os.Stdin = origStdin }()

	if err := handlePackCommand(configDir, []string{"edit", "custom"}); err != nil {
		t.Fatalf("pack edit: %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read edited pack: %v", err)
	}
	if !strings.Contains(string(after), "edited") {
		t.Fatalf("pack edit did not update content: %s", string(after))
	}
}

func TestHandleTemplateNew(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(configDir); err != nil {
		t.Fatalf("ensureConfigStructure: %v", err)
	}

	if err := handleTemplateCommand(configDir, []string{"new", "custom"}); err != nil {
		t.Fatalf("template new: %v", err)
	}

	path := filepath.Join(configDir, templatesDirName, "custom.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read custom template: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "{{intent}}") || !strings.Contains(content, "{{guidelines}}") {
		t.Fatalf("custom template missing placeholders: %s", content)
	}
}
