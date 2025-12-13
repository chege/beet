package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const internalInstruction = "Internal instruction: clarify, rephrase, infer reasonable gaps, adhere to the template, and output only the final instruction text."
const workPromptFilename = "WORK_PROMPT.md"

type guideline struct {
	name    string
	content string
}

func renderTemplate(template, intent, guidelines string) string {
	rendered := strings.ReplaceAll(template, "{{intent}}", strings.TrimSpace(intent))
	return strings.ReplaceAll(rendered, "{{guidelines}}", strings.TrimSpace(guidelines))
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

	var b strings.Builder
	for _, g := range guidelines {
		b.WriteString(fmt.Sprintf("%s: %s\n", g.name, g.content))
	}
	return strings.TrimSpace(b.String())
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
