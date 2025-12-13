package main

import (
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
