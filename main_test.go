package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	template := "Intent: {{intent}}\nRules: {{guidelines}}\n"
	out := renderTemplate(template, " ship feature X ", "rule1")
	want := "Intent: ship feature X\nRules: rule1\n"
	if out != want {
		t.Fatalf("renderTemplate mismatch\nwant: %q\ngot:  %q", want, out)
	}
}

func TestBuildWorkPromptInjectsGuidelines(t *testing.T) {
	template := "Hello {{intent}}\n{{guidelines}}\n"
	guides := []guideline{
		{name: "alpha", content: "keep it short"},
		{name: "beta", content: "ship it"},
	}
	out := buildWorkPrompt("default", template, guides, "do the thing")
	if !containsAll(out, []string{"do the thing", "keep it short", "beta", "Internal instruction"}) {
		t.Fatalf("work prompt missing expected content: %s", out)
	}
}

func containsAll(haystack string, needles []string) bool {
	for _, n := range needles {
		if !strings.Contains(haystack, n) {
			return false
		}
	}
	return true
}

func TestWriteWorkPrompt(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(configDir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}
	if err := bootstrapDefaults(configDir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	output := filepath.Join(t.TempDir(), "WORK_PROMPT.md")
	intent := "ship the feature"
	if err := writeWorkPrompt(configDir, "", intent, output); err != nil {
		t.Fatalf("writeWorkPrompt returned error: %v", err)
	}

	b, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read WORK_PROMPT: %v", err)
	}
	content := string(b)

	for _, expected := range []string{intent, "Internal instruction", "Template: default"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("expected content to contain %q; got %s", expected, content)
		}
	}
}
