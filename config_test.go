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
		filepath.Join(dir, packsDirName),
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

func TestBootstrapDefaultsCreatesDefaultPack(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}

	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	path := filepath.Join(dir, packsDirName, "default.yaml")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read default pack: %v", err)
	}

	want, err := os.ReadFile(filepath.Join("defaults", "packs", "default.yaml"))
	if err != nil {
		t.Fatalf("read source default pack: %v", err)
	}

	if string(got) != string(want) {
		t.Fatalf("default pack content mismatch:\n got: %s\nwant: %s", string(got), string(want))
	}
}

func TestBootstrapDefaultsCreatesExtendedPack(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}

	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	path := filepath.Join(dir, packsDirName, "extended.yaml")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read extended pack: %v", err)
	}

	want, err := os.ReadFile(filepath.Join("defaults", "packs", "extended.yaml"))
	if err != nil {
		t.Fatalf("read source extended pack: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("extended pack content mismatch:\n got: %s\nwant: %s", string(got), string(want))
	}
}

func TestBootstrapDefaultsCreatesComprehensivePack(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}

	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	path := filepath.Join(dir, packsDirName, "comprehensive.yaml")
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read comprehensive pack: %v", err)
	}

	want, err := os.ReadFile(filepath.Join("defaults", "packs", "comprehensive.yaml"))
	if err != nil {
		t.Fatalf("read source comprehensive pack: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("comprehensive pack content mismatch:\n got: %s\nwant: %s", string(got), string(want))
	}
}

func TestRestoreDefaultsNonDestructive(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}

	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	defaultPath := filepath.Join(dir, templatesDirName, "default.md")
	custom := []byte("custom content")
	if err := os.WriteFile(defaultPath, custom, 0o644); err != nil {
		t.Fatalf("overwrite default template: %v", err)
	}

	missing := filepath.Join(dir, templatesDirName, "prd.md")
	if err := os.Remove(missing); err != nil {
		t.Fatalf("remove template: %v", err)
	}

	if err := restoreDefaults(dir); err != nil {
		t.Fatalf("restoreDefaults returned error: %v", err)
	}

	after, err := os.ReadFile(defaultPath)
	if err != nil {
		t.Fatalf("read default template: %v", err)
	}
	if !reflect.DeepEqual(after, custom) {
		t.Fatalf("default template overwritten during restore")
	}

	restored, err := os.ReadFile(missing)
	if err != nil {
		t.Fatalf("read restored template: %v", err)
	}
	source, err := os.ReadFile(filepath.Join("defaults", "templates", "prd.md"))
	if err != nil {
		t.Fatalf("read source template: %v", err)
	}

	if string(restored) != string(source) {
		t.Fatalf("restored template mismatch: got %q want %q", string(restored), string(source))
	}
}

func TestRequireConfigStateErrorsWithoutPacks(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, templatesDirName, "default.md"), []byte("content"), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	if err := requireConfigState(dir); err == nil {
		t.Fatalf("requireConfigState should error when no packs exist")
	}
}

func TestRequireConfigStateSucceedsWithTemplatesAndPacks(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, templatesDirName, "default.md"), []byte("content"), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, packsDirName, "default.yaml"), []byte("outputs: []"), 0o644); err != nil {
		t.Fatalf("write pack: %v", err)
	}

	if err := requireConfigState(dir); err != nil {
		t.Fatalf("requireConfigState returned error: %v", err)
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

	sourceNames, err := os.ReadDir(filepath.Join("defaults", "templates"))
	if err != nil {
		t.Fatalf("read default templates dir: %v", err)
	}
	var want []string
	for _, entry := range sourceNames {
		if entry.IsDir() {
			continue
		}
		want = append(want, entry.Name())
	}
	if !reflect.DeepEqual(names, sortedCopy(want)) {
		t.Fatalf("listTemplates = %v, want %v", names, sortedCopy(want))
	}
}

func TestListPacksSorted(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}
	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	extra := filepath.Join(dir, packsDirName, "custom.yaml")
	if err := os.WriteFile(extra, []byte("outputs: []"), 0o644); err != nil {
		t.Fatalf("write pack: %v", err)
	}

	names, err := listPacks(dir)
	if err != nil {
		t.Fatalf("listPacks returned error: %v", err)
	}

	want := sortedCopy([]string{"custom.yaml", "default.yaml", "extended.yaml", "comprehensive.yaml"})
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("listPacks = %v, want %v", names, want)
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

func TestLoadPackDefault(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}
	if err := bootstrapDefaults(dir); err != nil {
		t.Fatalf("bootstrapDefaults returned error: %v", err)
	}

	p, err := loadPack(dir, "")
	if err != nil {
		t.Fatalf("loadPack returned error: %v", err)
	}

	if len(p.Outputs) != 2 {
		t.Fatalf("len(outputs) = %d, want 2", len(p.Outputs))
	}
	if p.Outputs[0].File == "" || p.Outputs[0].Template == "" {
		t.Fatalf("pack outputs should include file and template")
	}
}

func TestLoadPackValidates(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "cfg")
	if err := ensureConfigStructure(dir); err != nil {
		t.Fatalf("ensureConfigStructure returned error: %v", err)
	}

	bad := "outputs:\n  - file: \"\"\n    template: \"\"\n"
	path := filepath.Join(dir, packsDirName, "bad.yaml")
	if err := os.WriteFile(path, []byte(bad), 0o644); err != nil {
		t.Fatalf("write pack: %v", err)
	}

	if _, err := loadPack(dir, "bad"); err == nil {
		t.Fatalf("expected validation error")
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
