package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestResolveConfigDirEnvOverride(t *testing.T) {
	tmp := t.TempDir()
	override := filepath.Join(tmp, "cfg")
	t.Setenv(envConfigDir, override)

	got, err := resolveConfigDir()
	if err != nil {
		t.Fatalf("resolveConfigDir returned error: %v", err)
	}

	want, err := filepath.Abs(override)
	if err != nil {
		t.Fatalf("filepath.Abs failed: %v", err)
	}

	if got != want {
		t.Fatalf("resolveConfigDir = %q, want %q", got, want)
	}
}

func TestResolveConfigDirUsesHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(envConfigDir, "")
	t.Setenv("HOME", tmp)

	got, err := resolveConfigDir()
	if err != nil {
		t.Fatalf("resolveConfigDir returned error: %v", err)
	}

	want := filepath.Join(tmp, defaultConfigFolder)
	want, err = filepath.Abs(want)
	if err != nil {
		t.Fatalf("filepath.Abs failed: %v", err)
	}

	if got != want {
		t.Fatalf("resolveConfigDir = %q, want %q", got, want)
	}
}

func TestEnsureConfigStructure(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}

	for _, path := range []string{
		dir,
		filepath.Join(dir, templatesDirName),
		filepath.Join(dir, guidelinesDirName),
	} {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			t.Fatalf("expected directory at %s, stat err: %v", path, err)
		}
	}
}

func TestBootstrapDefaultsIdempotent(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}

	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	defaultPath := filepath.Join(dir, templatesDirName, "default.md")
	original, err := os.ReadFile(defaultPath)
	if err != nil {
		t.Fatalf("read default template: %v", err)
	}

	custom := []byte("custom content")
	if err := os.WriteFile(defaultPath, custom, 0o644); err != nil {
		t.Fatalf("rewrite default template: %v", err)
	}

	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("second bootstrapDefaults error: %v", err)
	}

	after, err := os.ReadFile(defaultPath)
	if err != nil {
		t.Fatalf("read default template after: %v", err)
	}

	if !reflect.DeepEqual(after, custom) {
		t.Fatalf("bootstrapDefaults overwrote existing file; got %q want %q", string(after), string(custom))
	}

	if len(original) == 0 {
		t.Fatalf("original default template should not be empty")
	}
}

func TestListTemplatesSorted(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}
	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	names, err := listTemplates(dir)
	if err != nil {
		t.Fatalf("listTemplates returned error: %v", err)
	}

	want := make([]string, 0, len(defaultTemplates))
	for name := range defaultTemplates {
		want = append(want, name)
	}
	if len(names) != len(want) {
		t.Fatalf("listTemplates len=%d, want %d", len(names), len(want))
	}

	if !reflect.DeepEqual(names, sortedCopy(want)) {
		t.Fatalf("listTemplates = %v, want %v", names, sortedCopy(want))
	}
}

func TestLoadTemplateDefault(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}
	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	got, err := loadTemplate(dir, "")
	if err != nil {
		t.Fatalf("loadTemplate returned error: %v", err)
	}

	if got == "" {
		t.Fatalf("loadTemplate returned empty template")
	}
}

func TestLoadTemplateAddsExtension(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}

	want := "hello"
	path := filepath.Join(dir, templatesDirName, "custom.md")
	if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	got, err := loadTemplate(dir, "custom")
	if err != nil {
		t.Fatalf("loadTemplate returned error: %v", err)
	}
	if got != want {
		t.Fatalf("loadTemplate = %q, want %q", got, want)
	}
}

func TestLoadGuidelinesSorted(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}

	guidelineFiles := map[string]string{
		"b.md": "second",
		"a.md": "first",
	}
	for name, content := range guidelineFiles {
		if err := os.WriteFile(filepath.Join(dir, guidelinesDirName, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write guideline %s: %v", name, err)
		}
	}

	got, err := loadGuidelines(dir)
	if err != nil {
		t.Fatalf("loadGuidelines returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("len(guidelines) = %d, want 2", len(got))
	}

	if got[0].name != "a" || got[0].content != "first" {
		t.Fatalf("first guideline = %+v, want name a content first", got[0])
	}
	if got[1].name != "b" || got[1].content != "second" {
		t.Fatalf("second guideline = %+v, want name b content second", got[1])
	}
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
