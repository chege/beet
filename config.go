package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

import "gopkg.in/yaml.v3"

const (
	envConfigDir        = "BEET_CONFIG_DIR"
	defaultConfigFolder = ".beet"
	templatesDirName    = "templates"
	guidelinesDirName   = "guidelines"
	packsDirName        = "packs"
	defaultTemplateName = "default.md"
	defaultPackName     = "default.yaml"
)

var defaultTemplates = map[string]string{
	"default.md":    "## Task\n{{intent}}\n\n## Guidelines\n{{guidelines}}\n",
	"agents.md":     "## Agents\n\n{{guidelines}}\n",
	"prd.md":        "# Product Requirements\n\n{{intent}}\n\n## Guidelines\n{{guidelines}}\n",
	"srs.md":        "# Software Requirements Specification\n\n{{intent}}\n\n## Guidelines\n{{guidelines}}\n",
	"guidelines.md": "# Guidelines\n\n{{guidelines}}\n",
	"intent.md":     "# Intent\n\n{{intent}}\n",
	"design.md":     "# Design\n\n{{intent}}\n\n## Notes\n{{guidelines}}\n",
	"rules.md":      "# Rules\n\n{{guidelines}}\n",
	"plan.md":       "# Plan\n\n- [ ] {{intent}}\n",
	"progress.md":   "# Progress\n\n- [ ] Not started\n",
}

var defaultGuidelines = map[string]string{
	"principles.md": "Be clear. Be concise. Prefer deterministic, reproducible instructions.",
}

var defaultPacks = map[string]string{
	"default.yaml": "outputs:\n" +
		"  - file: WORK_PROMPT.md\n" +
		"    template: default.md\n" +
		"  - file: agents.md\n" +
		"    template: agents.md\n",
	"extended.yaml": "outputs:\n" +
		"  - file: WORK_PROMPT.md\n" +
		"    template: default.md\n" +
		"  - file: agents.md\n" +
		"    template: agents.md\n" +
		"  - file: PRD.md\n" +
		"    template: prd.md\n" +
		"  - file: SRS.md\n" +
		"    template: srs.md\n" +
		"  - file: GUIDELINES.md\n" +
		"    template: guidelines.md\n",
	"comprehensive.yaml": "outputs:\n" +
		"  - file: WORK_PROMPT.md\n" +
		"    template: default.md\n" +
		"  - file: agents.md\n" +
		"    template: agents.md\n" +
		"  - file: INTENT.md\n" +
		"    template: intent.md\n" +
		"  - file: DESIGN.md\n" +
		"    template: design.md\n" +
		"  - file: RULES.md\n" +
		"    template: rules.md\n" +
		"  - file: PLAN.md\n" +
		"    template: plan.md\n" +
		"  - file: PROGRESS.md\n" +
		"    template: progress.md\n",
}

func resolveConfigDir() (string, error) {
	if override := os.Getenv(envConfigDir); override != "" {
		return filepath.Abs(override)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}

	return filepath.Abs(filepath.Join(home, defaultConfigFolder))
}

func ensureConfigStructure(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, templatesDirName), 0o755); err != nil {
		return fmt.Errorf("create templates dir: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, guidelinesDirName), 0o755); err != nil {
		return fmt.Errorf("create guidelines dir: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, packsDirName), 0o755); err != nil {
		return fmt.Errorf("create packs dir: %w", err)
	}

	return nil
}

func bootstrapDefaults(dir string) error {
	for name, content := range defaultTemplates {
		if err := writeIfMissing(filepath.Join(dir, templatesDirName, name), content); err != nil {
			return err
		}
	}

	for name, content := range defaultGuidelines {
		if err := writeIfMissing(filepath.Join(dir, guidelinesDirName, name), content); err != nil {
			return err
		}
	}

	for name, content := range defaultPacks {
		if err := writeIfMissing(filepath.Join(dir, packsDirName, name), content); err != nil {
			return err
		}
	}

	return nil
}

func writeIfMissing(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check file %s: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write default %s: %w", path, err)
	}

	return nil
}

func listTemplates(configDir string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(configDir, templatesDirName))
	if err != nil {
		return nil, fmt.Errorf("read templates: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		names = append(names, entry.Name())
	}

	sort.Strings(names)
	return names, nil
}

func listPacks(configDir string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(configDir, packsDirName))
	if err != nil {
		return nil, fmt.Errorf("read packs: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		names = append(names, entry.Name())
	}

	sort.Strings(names)
	return names, nil
}

func loadTemplate(configDir, name string) (string, error) {
	name = normalizeTemplateName(name)

	path := filepath.Join(configDir, templatesDirName, name)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("load template %s: %w", name, err)
	}
	return string(b), nil
}

func loadGuidelines(configDir string) ([]guideline, error) {
	dir := filepath.Join(configDir, guidelinesDirName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read guidelines: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var out []guideline
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read guideline %s: %w", entry.Name(), err)
		}
		out = append(out, guideline{name: name, content: string(content)})
	}

	return out, nil
}

func normalizeTemplateName(name string) string {
	if name == "" {
		name = defaultTemplateName
	}
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	return name
}

type pack struct {
	Outputs []packOutput `yaml:"outputs"`
}

type packOutput struct {
	File     string `yaml:"file"`
	Template string `yaml:"template"`
}

func normalizePackName(name string) string {
	if name == "" {
		name = defaultPackName
	}
	if !strings.HasSuffix(name, ".yaml") {
		name += ".yaml"
	}
	return name
}

func loadPack(configDir, name string) (pack, error) {
	name = normalizePackName(name)
	path := filepath.Join(configDir, packsDirName, name)

	data, err := os.ReadFile(path)
	if err != nil {
		return pack{}, fmt.Errorf("load pack %s: %w", name, err)
	}

	var p pack
	if err := yaml.Unmarshal(data, &p); err != nil {
		return pack{}, fmt.Errorf("parse pack %s: %w", name, err)
	}

	if len(p.Outputs) == 0 {
		return pack{}, fmt.Errorf("pack %s has no outputs", name)
	}

	for i, out := range p.Outputs {
		if strings.TrimSpace(out.File) == "" {
			return pack{}, fmt.Errorf("pack %s output %d missing file", name, i)
		}
		if strings.TrimSpace(out.Template) == "" {
			return pack{}, fmt.Errorf("pack %s output %d missing template", name, i)
		}
	}

	return p, nil
}

func requireConfigState(configDir string) error {
	packs, err := listPacks(configDir)
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		return fmt.Errorf("no packs found in %s; add a pack or re-run bootstrap", filepath.Join(configDir, packsDirName))
	}

	templates, err := listTemplates(configDir)
	if err != nil {
		return err
	}
	if len(templates) == 0 {
		return fmt.Errorf("no templates found in %s; add a template or re-run bootstrap", filepath.Join(configDir, templatesDirName))
	}

	return nil
}

func prepareConfig() (string, error) {
	dir, err := resolveConfigDir()
	if err != nil {
		return "", err
	}

	if err := ensureConfigStructure(dir); err != nil {
		return "", err
	}

	if err := bootstrapDefaults(dir); err != nil {
		return "", err
	}

	return dir, nil
}
