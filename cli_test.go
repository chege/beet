package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleGenerateCreatesWorkPrompt(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(configDir); err != nil {
		t.Fatalf("ensureConfigStructure: %v", err)
	}
	if err := bootstrapDefaults(configDir); err != nil {
		t.Fatalf("bootstrapDefaults: %v", err)
	}

	workdir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(origWD)
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := handleGenerate(configDir, []string{"-t", "default", "ship", "it"}); err != nil {
		t.Fatalf("handleGenerate returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workdir, workPromptFilename))
	if err != nil {
		t.Fatalf("read work prompt: %v", err)
	}

	for _, expected := range []string{"ship it", "Template: default", "Internal instruction"} {
		if !strings.Contains(string(content), expected) {
			t.Fatalf("work prompt missing %q; content: %s", expected, string(content))
		}
	}
}

func TestParseIntentFromStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	if _, err := w.WriteString("from stdin"); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	w.Close()

	intent, err := parseIntent(nil)
	if err != nil {
		t.Fatalf("parseIntent returned error: %v", err)
	}
	if intent != "from stdin" {
		t.Fatalf("parseIntent = %q, want from stdin", intent)
	}
}

func TestParseIntentFromFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "intent.txt")
	if err := os.WriteFile(tmpFile, []byte("from file\n"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	intent, err := parseIntent([]string{tmpFile})
	if err != nil {
		t.Fatalf("parseIntent returned error: %v", err)
	}
	if intent != "from file" {
		t.Fatalf("parseIntent = %q, want from file", intent)
	}
}
