package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	envConfigDir        = "PF_CONFIG_DIR"
	defaultConfigFolder = ".pf"
	templatesDirName    = "templates"
	guidelinesDirName   = "guidelines"
	defaultTemplateName = "default.md"
)

var defaultTemplates = map[string]string{
	"default.md": "## Task\n{{intent}}\n\n## Guidelines\n{{guidelines}}\n",
}

var defaultGuidelines = map[string]string{
	"principles.md": "Be clear. Be concise. Prefer deterministic, reproducible instructions.",
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
