package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const internalInstruction = "Internal instruction: clarify, rephrase, infer reasonable gaps, adhere to the template, and output only the final instruction text."
const workPromptFilename = "WORK_PROMPT.md"
const agentsFilename = "agents.md"
const cliTimeoutEnv = "BEET_CLI_TIMEOUT"
const defaultCLITimeout = 5 * time.Minute

type guideline struct {
	name    string
	content string
}

func renderTemplate(template, intent, guidelines string) string {
	rendered := strings.ReplaceAll(template, "{{intent}}", strings.TrimSpace(intent))
	return strings.ReplaceAll(rendered, "{{guidelines}}", guidelines)
}

func buildWorkPrompt(templateName, template string, guidelines []guideline, intent string) string {
	guidelineText := formatGuidelines(guidelines)
	body := renderTemplate(template, intent, guidelineText)

	var b strings.Builder
	b.WriteString(internalInstruction)
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Template: %s\n", templateName))
	b.WriteString(body)
	return b.String()
}

func formatGuidelines(guidelines []guideline) string {
	if len(guidelines) == 0 {
		return ""
	}

	parts := make([]string, 0, len(guidelines))
	for _, g := range guidelines {
		parts = append(parts, g.content)
	}

	return strings.Join(parts, "\n\n")
}

func buildWorkPromptContent(configDir, templateName, intent string) (string, error) {
	normalized := normalizeTemplateName(templateName)

	template, err := loadTemplate(configDir, templateName)
	if err != nil {
		return "", err
	}

	guidelines, err := loadGuidelines(configDir)
	if err != nil {
		return "", err
	}

	label := strings.TrimSuffix(normalized, filepath.Ext(normalized))
	return buildWorkPrompt(label, template, guidelines, intent), nil
}

func writeWorkPrompt(configDir, templateName, intent, outputPath string) error {
	content, err := buildWorkPromptContent(configDir, templateName, intent)
	if err != nil {
		return err
	}

	if outputPath == "" {
		outputPath = workPromptFilename
	}

	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outputPath, err)
	}

	return nil
}

func buildAgentsContent(guidelines []guideline) string {
	var b strings.Builder
	b.WriteString("## Agents\n\n")
	if text := formatGuidelines(guidelines); text != "" {
		b.WriteString(text)
		b.WriteString("\n")
	}
	return b.String()
}

func writeAgents(configDir, outputPath string, force bool) error {
	if outputPath == "" {
		outputPath = agentsFilename
	}

	if _, err := os.Stat(outputPath); err == nil && !force {
		return nil
	}

	guidelines, err := loadGuidelines(configDir)
	if err != nil {
		return err
	}

	content := buildAgentsContent(guidelines)
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outputPath, err)
	}

	return nil
}

func writeRenderedOutput(path string, content string, forceAgents bool) error {
	if strings.EqualFold(filepath.Base(path), agentsFilename) {
		if _, err := os.Stat(path); err == nil && !forceAgents {
			return nil
		}
	}

	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create output dir: %w", err)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func runCLI(ctx context.Context, cli detectedCLI, prompt string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	logVerbose("running CLI %s (%s)", cli.name, cli.path)

	cmd := exec.CommandContext(ctx, cli.path)
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stderr = os.Stderr

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s exec failed: %w", cli.name, err)
	}

	logVerbose("CLI %s completed", cli.name)

	return out.String(), nil
}

func cliTimeout() time.Duration {
	if raw := strings.TrimSpace(os.Getenv(cliTimeoutEnv)); raw != "" {
		if duration, err := time.ParseDuration(raw); err == nil && duration > 0 {
			return duration
		}
	}
	return defaultCLITimeout
}
