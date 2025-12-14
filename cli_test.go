package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type browserFake func(string) error

func (f browserFake) OpenFile(path string) error { return f(path) }

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
	defer func() {
		_ = os.Chdir(origWD)
	}()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := handleGenerate(configDir, []string{"-t", "default", "--exec=false", "ship", "it"}); err != nil {
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
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

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

func TestParseIntentFromEditor(t *testing.T) {
	script := filepath.Join(t.TempDir(), "editor.sh")
	content := "#!/bin/sh\necho \"from editor\" > \"$1\"\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	t.Setenv("EDITOR", script)

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	origStdin := os.Stdin
	os.Stdin = devNull
	defer func() { os.Stdin = origStdin }()

	intent, err := parseIntent(nil)
	if err != nil {
		t.Fatalf("parseIntent returned error: %v", err)
	}
	if intent != "from editor" {
		t.Fatalf("parseIntent = %q, want from editor", intent)
	}
}

func TestIntentFromDefaultAppFallback(t *testing.T) {
	origOpen := defaultBrowser
	origWait := waitForEdit
	t.Cleanup(func() {
		defaultBrowser = origOpen
		waitForEdit = origWait
	})

	defaultBrowser = browserFake(func(path string) error {
		return os.WriteFile(path, []byte("from default app"), 0o644)
	})
	waitForEdit = func(string) error { return nil }
	t.Setenv("EDITOR", "")

	intent, err := intentFromEditor()
	if err != nil {
		t.Fatalf("intentFromEditor returned error: %v", err)
	}
	if intent != "from default app" {
		t.Fatalf("intentFromEditor = %q, want from default app", intent)
	}
}

func TestHandleGenerateDryRun(t *testing.T) {
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
	defer func() {
		_ = os.Chdir(origWD)
	}()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	if err := handleGenerate(configDir, []string{"--dry-run", "do", "it"}); err != nil {
		t.Fatalf("handleGenerate returned error: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}
	os.Stdout = origStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}

	if strings.Contains(string(out), "Template: default") == false {
		t.Fatalf("dry-run output missing template label: %s", string(out))
	}

	if _, err := os.Stat(filepath.Join(workdir, workPromptFilename)); err == nil {
		t.Fatalf("WORK_PROMPT.md should not be written in dry-run")
	}
}

func TestHandleGenerateExecRunsCLI(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(configDir); err != nil {
		t.Fatalf("ensureConfigStructure: %v", err)
	}
	if err := bootstrapDefaults(configDir); err != nil {
		t.Fatalf("bootstrapDefaults: %v", err)
	}

	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "exec.log")
	script := filepath.Join(binDir, "codex")
	content := "#!/bin/sh\n/bin/cat >> \"$PF_EXEC_LOG\"\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	t.Setenv("PATH", binDir)
	t.Setenv("PF_EXEC_LOG", logFile)

	workdir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(origWD)
	}()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := handleGenerate(configDir, []string{"--exec", "do", "it"}); err != nil {
		t.Fatalf("handleGenerate returned error: %v", err)
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read exec log: %v", err)
	}
	contentStr := strings.TrimSpace(string(data))
	if !strings.Contains(contentStr, "Internal instruction") || !strings.Contains(contentStr, "do it") {
		t.Fatalf("exec log missing prompt content: %s", contentStr)
	}
}

func TestHandleGenerateRespectsPackOutputs(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(configDir); err != nil {
		t.Fatalf("ensureConfigStructure: %v", err)
	}
	if err := bootstrapDefaults(configDir); err != nil {
		t.Fatalf("bootstrapDefaults: %v", err)
	}

	customPack := "outputs:\n  - file: first.md\n    template: default.md\n  - file: second.md\n    template: default.md\n"
	packPath := filepath.Join(configDir, packsDirName, "custom.yaml")
	if err := os.WriteFile(packPath, []byte(customPack), 0o644); err != nil {
		t.Fatalf("write pack: %v", err)
	}

	workdir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(origWD)
	}()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := handleGenerate(configDir, []string{"--pack", "custom", "--exec=false", "ship"}); err != nil {
		t.Fatalf("handleGenerate returned error: %v", err)
	}

	for _, name := range []string{"first.md", "second.md"} {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(data), "ship") {
			t.Fatalf("%s missing intent", name)
		}
	}
}

func TestHandleGenerateUsesEditorWhenNoArgs(t *testing.T) {
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
	defer func() {
		_ = os.Chdir(origWD)
	}()
	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	script := filepath.Join(t.TempDir(), "editor.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho \"from editor\" > \"$1\"\n"), 0o755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	t.Setenv("EDITOR", script)

	origStdin := os.Stdin
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	os.Stdin = devNull
	defer func() { os.Stdin = origStdin }()

	if err := handleGenerate(configDir, []string{"--exec=false"}); err != nil {
		t.Fatalf("handleGenerate returned error: %v", err)
	}

	content, err := os.ReadFile(workPromptFilename)
	if err != nil {
		t.Fatalf("read work prompt: %v", err)
	}
	if !strings.Contains(string(content), "from editor") {
		t.Fatalf("work prompt missing editor content: %s", string(content))
	}
}

func TestHandleGenerateHelpShowsUsage(t *testing.T) {
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
	defer func() { os.Stdout = origStdout }()

	if err := handleGenerate(configDir, []string{"--help"}); err != nil {
		t.Fatalf("handleGenerate returned error: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read pipe: %v", err)
	}

	output := string(out)
	if !strings.Contains(output, "Usage: beet") {
		t.Fatalf("help output missing usage line: %s", output)
	}
	if !strings.Contains(output, "-dry-run") {
		t.Fatalf("help output missing flags: %s", output)
	}
}
