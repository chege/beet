// Package main implements the Beet CLI.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

import "github.com/pkg/browser"

type browserOpener interface {
	OpenFile(path string) error
}

type pkgBrowser struct{}

func (pkgBrowser) OpenFile(path string) error {
	return browser.OpenFile(path)
}

var defaultBrowser browserOpener = pkgBrowser{}
var waitForEdit = func(path string) error {
	if _, err := fmt.Fprintf(os.Stdout, "Edit intent in %s, then press Enter to continue: ", path); err != nil {
		return err
	}
	_, err := fmt.Fscanln(os.Stdin)
	if err == io.EOF {
		return nil
	}
	return err
}

func main() {
	configDir, err := prepareConfig()
	if err != nil {
		log.Fatalf("prepare config: %v", err)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		if err := handleGenerate(configDir, args); err != nil {
			log.Fatalf("generate prompt: %v", err)
		}
		return
	}

	switch args[0] {
	case "templates":
		names, err := listTemplates(configDir)
		if err != nil {
			log.Fatalf("list templates: %v", err)
		}
		for _, name := range names {
			fmt.Println(name)
		}
	case "packs":
		names, err := listPacks(configDir)
		if err != nil {
			log.Fatalf("list packs: %v", err)
		}
		for _, name := range names {
			fmt.Println(name)
		}
	case "doctor":
		if err := runDoctor(os.Stdout); err != nil {
			log.Fatalf("doctor: %v", err)
		}
	default:
		if err := handleGenerate(configDir, args); err != nil {
			log.Fatalf("generate prompt: %v", err)
		}
	}
}

func handleGenerate(configDir string, args []string) error {
	fs := flag.NewFlagSet("beet", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: beet [flags] [intent|file]")
		fmt.Fprintln(fs.Output(), "\nFlags:")
		fs.PrintDefaults()
		fmt.Fprintln(fs.Output(), "\nCommands: beet templates | beet packs | beet doctor")
		fmt.Fprintln(fs.Output(), "Notes: packs are bootstrapped and listable; generation currently writes WORK_PROMPT.md + agents.md.")
		fmt.Fprintln(fs.Output(), "       CLI execution defaults on (Codex preferred, Copilot fallback). Disable with --exec=false.")
	}

	template := fs.String("t", "", "template name")
	templateLong := fs.String("template", "", "template name")
	pack := fs.String("p", "", "pack name")
	packLong := fs.String("pack", "", "pack name")
	dryRun := fs.Bool("dry-run", false, "render without writing files")
	forceAgents := fs.Bool("force-agents", false, "overwrite agents.md")
	execFlag := fs.Bool("exec", true, "execute detected CLI with WORK_PROMPT.md")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	intent, err := parseIntent(fs.Args())
	if err != nil {
		return err
	}

	tmplName := firstNonEmpty(*template, *templateLong)
	packName := firstNonEmpty(*pack, *packLong)

	if packName == "" {
		packName = defaultPackName
	}

	p, err := loadPack(configDir, packName)
	if err != nil {
		return err
	}

	guidelines, err := loadGuidelines(configDir)
	if err != nil {
		return err
	}

	var cli detectedCLI
	if *execFlag {
		cli, err = requireCLI()
		if err != nil {
			return err
		}
	}

	for _, out := range p.Outputs {
		templateName := out.Template
		if tmplName != "" && strings.EqualFold(out.File, workPromptFilename) {
			templateName = tmplName
		}

		templateContent, err := loadTemplate(configDir, templateName)
		if err != nil {
			return err
		}

		normalized := normalizeTemplateName(templateName)
		label := strings.TrimSuffix(normalized, filepath.Ext(normalized))
		prompt := buildWorkPrompt(label, templateContent, guidelines, intent)

		if *dryRun {
			fmt.Printf("=== %s ===\n%s\n", out.File, prompt)
			continue
		}

		content := prompt
		if *execFlag {
			content, err = runCLI(cli, prompt)
			if err != nil {
				return err
			}
		}

		if err := writeRenderedOutput(out.File, content, *forceAgents); err != nil {
			return err
		}
	}

	return nil
}

func parseIntent(remaining []string) (string, error) {
	if len(remaining) > 0 {
		if len(remaining) == 1 {
			if info, err := os.Stat(remaining[0]); err == nil && !info.IsDir() {
				b, err := os.ReadFile(remaining[0])
				if err != nil {
					return "", fmt.Errorf("read intent file: %w", err)
				}
				intent := strings.TrimSpace(string(b))
				if intent != "" {
					return intent, nil
				}
			}
		}
		return strings.TrimSpace(strings.Join(remaining, " ")), nil
	}

	info, err := os.Stdin.Stat()
	if err == nil && (info.Mode()&os.ModeCharDevice) == 0 {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		intent := strings.TrimSpace(string(b))
		if intent != "" {
			return intent, nil
		}
	}

	return intentFromEditor()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func intentFromEditor() (string, error) {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor == "" {
		return intentFromDefaultApp()
	}

	tmp, err := os.CreateTemp("", "beet-intent-*.md")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}

	cmd := exec.Command(editor, tmp.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("launch editor: %w", err)
	}

	b, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("read editor output: %w", err)
	}

	intent := strings.TrimSpace(string(b))
	if intent == "" {
		return "", fmt.Errorf("intent is empty; provide input")
	}

	return intent, nil
}

func intentFromDefaultApp() (string, error) {
	tmp, err := os.CreateTemp("", "beet-intent-*.md")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("close temp file: %w", err)
	}

	if err := defaultBrowser.OpenFile(tmp.Name()); err != nil {
		return "", fmt.Errorf("open default app: %w", err)
	}

	if err := waitForEdit(tmp.Name()); err != nil {
		return "", fmt.Errorf("await default edit: %w", err)
	}

	b, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("read intent: %w", err)
	}

	intent := strings.TrimSpace(string(b))
	if intent == "" {
		return "", fmt.Errorf("intent is empty; provide input")
	}

	return intent, nil
}
