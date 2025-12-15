package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	origWait := waitForContentFn
	t.Cleanup(func() {
		defaultBrowser = origOpen
		waitForContentFn = origWait
	})

	defaultBrowser = browserFake(func(path string) error {
		return os.WriteFile(path, []byte("from default app"), 0o644)
	})
	waitForContentFn = func(string) error { return nil }
	t.Setenv("EDITOR", "")

	intent, err := intentFromEditor()
	if err != nil {
		t.Fatalf("intentFromEditor returned error: %v", err)
	}
	if intent != "from default app" {
		t.Fatalf("intentFromEditor = %q, want from default app", intent)
	}
}

func TestIntentFromEditorWaitsForContent(t *testing.T) {
	script := filepath.Join(t.TempDir(), "editor.sh")
	content := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
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

	origWait := waitForContentFn
	waitCalled := false
	waitForContentFn = func(path string) error {
		waitCalled = true
		return os.WriteFile(path, []byte("from wait"), 0o644)
	}
	t.Cleanup(func() {
		waitForContentFn = origWait
	})

	intent, err := parseIntent(nil)
	if err != nil {
		t.Fatalf("parseIntent returned error: %v", err)
	}
	if intent != "from wait" {
		t.Fatalf("parseIntent = %q, want from wait", intent)
	}
	if !waitCalled {
		t.Fatalf("waitForContentFn not invoked")
	}
}

func TestHandleCompletionBash(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w
	if err := handleCompletion([]string{"--shell", "bash"}); err != nil {
		t.Fatalf("handleCompletion bash: %v", err)
	}
	_ = w.Close()
	os.Stdout = origStdout

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read completion: %v", err)
	}
	out := string(data)
	for _, expected := range []string{"complete -F _beet_completions beet", "--dry-run", "pack"} {
		if !strings.Contains(out, expected) {
			t.Fatalf("bash completion missing %q: %s", expected, out)
		}
	}
}

func TestHandleCompletionZsh(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w
	if err := handleCompletion([]string{"--shell", "zsh"}); err != nil {
		t.Fatalf("handleCompletion zsh: %v", err)
	}
	_ = w.Close()
	os.Stdout = origStdout

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read completion: %v", err)
	}
	out := string(data)
	for _, expected := range []string{"#compdef beet", "_values 'commands'"} {
		if !strings.Contains(out, expected) {
			t.Fatalf("zsh completion missing %q: %s", expected, out)
		}
	}
}

func TestHandleCompletionRejectsUnknownShell(t *testing.T) {
	if err := handleCompletion([]string{"--shell", "fish"}); err == nil {
		t.Fatalf("handleCompletion should error on unknown shell")
	}
}

func TestDefaultWaitForContentAcceptsNonWhitespace(t *testing.T) {
	origTimeout := waitForContentTimeout
	origInterval := waitForContentInterval
	waitForContentTimeout = 500 * time.Millisecond
	waitForContentInterval = 5 * time.Millisecond
	defer func() {
		waitForContentTimeout = origTimeout
		waitForContentInterval = origInterval
	}()

	tmpFile := filepath.Join(t.TempDir(), "intent.md")
	if err := os.WriteFile(tmpFile, []byte(""), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		_ = os.WriteFile(tmpFile, []byte("content"), 0o644)
	}()

	if err := defaultWaitForContent(tmpFile); err != nil {
		t.Fatalf("defaultWaitForContent returned error: %v", err)
	}
}

func TestDefaultWaitForContentRejectsWhitespace(t *testing.T) {
	origTimeout := waitForContentTimeout
	origInterval := waitForContentInterval
	waitForContentTimeout = 50 * time.Millisecond
	waitForContentInterval = 5 * time.Millisecond
	defer func() {
		waitForContentTimeout = origTimeout
		waitForContentInterval = origInterval
	}()

	tmpFile := filepath.Join(t.TempDir(), "intent.md")
	if err := os.WriteFile(tmpFile, []byte("   \n\t"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	if err := defaultWaitForContent(tmpFile); err == nil {
		t.Fatalf("defaultWaitForContent should error on whitespace-only content")
	}
}

func TestParseIntentRejectsEmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "intent.txt")
	if err := os.WriteFile(tmpFile, []byte("  \n"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	if _, err := parseIntent([]string{tmpFile}); err == nil {
		t.Fatalf("parseIntent should error on empty file")
	}
}

func TestParseIntentRejectsEmptyArgs(t *testing.T) {
	if _, err := parseIntent([]string{"   ", "\n"}); err == nil {
		t.Fatalf("parseIntent should error on empty args")
	}
}

func TestParseIntentRejectsEmptyStdin(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if _, err := w.WriteString("   \n"); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	origStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = origStdin }()

	if _, err := parseIntent(nil); err == nil {
		t.Fatalf("parseIntent should error on empty stdin")
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

func TestHandleGenerateExtendedPack(t *testing.T) {
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

	if err := handleGenerate(configDir, []string{"--pack", "extended", "--exec=false", "plan release"}); err != nil {
		t.Fatalf("handleGenerate returned error: %v", err)
	}

	checkContains := func(name, substr string) {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if !strings.Contains(string(data), substr) {
			t.Fatalf("%s missing expected content", name)
		}
	}

	checkContains("PRD.md", "plan release")
	checkContains("SRS.md", "plan release")
	checkContains("GUIDELINES.md", "Guidelines")
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
