// Package main implements the Beet CLI.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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
var waitForContentFn = defaultWaitForContent

var (
	waitForContentTimeout  = 5 * time.Minute
	waitForContentInterval = 200 * time.Millisecond
)

const (
	bashCompletion = `_beet_completions() {
  local cur prev
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  local commands="templates packs doctor pack template config completion"
  local global_opts="--help --dry-run --exec --force-agents -t --template -p --pack"
  case "$prev" in
    pack)
      COMPREPLY=( $(compgen -W "list init edit" -- "$cur") )
      return 0
      ;;
    template)
      COMPREPLY=( $(compgen -W "new" -- "$cur") )
      return 0
      ;;
    config)
      COMPREPLY=( $(compgen -W "restore" -- "$cur") )
      return 0
      ;;
  esac
  COMPREPLY=( $(compgen -W "${commands} ${global_opts}" -- "$cur") )
}
complete -F _beet_completions beet
`
	zshCompletion = `#compdef beet

_beet_commands() {
  _values 'commands' templates packs doctor pack template config completion
}

_beet() {
  local state
  _arguments \
    '1: :->cmds' \
    '*::arg:->args'

  case $state in
    cmds)
      _beet_commands
      ;;
    args)
      case $words[1] in
        pack)
          _values 'pack commands' list init edit
          ;;
        template)
          _values 'template commands' new
          ;;
        config)
          _values 'config commands' restore
          ;;
        *)
          _values 'options' --help --dry-run --exec --force-agents -t --template -p --pack
          ;;
      esac
      ;;
  esac
}

_beet "$@"
`
)

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
	case "pack":
		if err := handlePackCommand(configDir, args[1:]); err != nil {
			log.Fatalf("pack: %v", err)
		}
	case "template":
		if err := handleTemplateCommand(configDir, args[1:]); err != nil {
			log.Fatalf("template: %v", err)
		}
	case "config":
		if err := handleConfig(configDir, args[1:]); err != nil {
			log.Fatalf("config: %v", err)
		}
	case "completion":
		if err := handleCompletion(args[1:]); err != nil {
			log.Fatalf("completion: %v", err)
		}
	default:
		if err := handleGenerate(configDir, args); err != nil {
			log.Fatalf("generate prompt: %v", err)
		}
	}
}

func handleConfig(configDir string, args []string) error {
	if len(args) == 0 || args[0] != "restore" {
		return fmt.Errorf("usage: beet config restore")
	}

	return restoreDefaults(configDir)
}

func handleCompletion(args []string) error {
	fs := flag.NewFlagSet("completion", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	shell := fs.String("shell", "bash", "shell (bash|zsh)")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	var script string
	switch strings.ToLower(strings.TrimSpace(*shell)) {
	case "bash":
		script = bashCompletion
	case "zsh":
		script = zshCompletion
	default:
		return fmt.Errorf("unknown shell %q (supported: bash, zsh)", *shell)
	}

	fmt.Fprint(os.Stdout, script)
	return nil
}

func handlePackCommand(configDir string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: beet pack [list|init|edit]")
	}

	switch args[0] {
	case "list":
		names, err := listPacks(configDir)
		if err != nil {
			return err
		}
		for _, name := range names {
			fmt.Println(name)
		}
		return nil
	case "init":
		fs := flag.NewFlagSet("pack init", flag.ContinueOnError)
		fs.SetOutput(os.Stdout)
		name := fs.String("name", "", "pack name")
		short := fs.String("n", "", "pack name")
		if err := fs.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		packName := firstNonEmpty(*name, *short)
		if strings.TrimSpace(packName) == "" {
			return fmt.Errorf("pack name required")
		}
		filename := normalizePackName(packName)
		path := filepath.Join(configDir, packsDirName, filename)
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("pack %s already exists", filename)
		}
		content := "outputs:\n  - file: WORK_PROMPT.md\n    template: default.md\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write pack: %w", err)
		}
		return nil
	case "edit":
		if len(args) < 2 {
			return fmt.Errorf("usage: beet pack edit <name>")
		}
		filename := normalizePackName(args[1])
		path := filepath.Join(configDir, packsDirName, filename)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("pack %s not found: %w", filename, err)
		}
		return openForEdit(path)
	default:
		return fmt.Errorf("usage: beet pack [list|init|edit]")
	}
}

func handleTemplateCommand(configDir string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: beet template new <name>")
	}

	switch args[0] {
	case "new":
		fs := flag.NewFlagSet("template new", flag.ContinueOnError)
		fs.SetOutput(os.Stdout)
		name := fs.String("name", "", "template name")
		short := fs.String("n", "", "template name")
		if err := fs.Parse(args[1:]); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		templateName := firstNonEmpty(*name, *short)
		if strings.TrimSpace(templateName) == "" && len(fs.Args()) > 0 {
			templateName = fs.Args()[0]
		}
		if strings.TrimSpace(templateName) == "" {
			return fmt.Errorf("template name required")
		}
		filename := normalizeTemplateName(templateName)
		path := filepath.Join(configDir, templatesDirName, filename)
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("template %s already exists", filename)
		}
		content := "# New template\n\n{{intent}}\n\n{{guidelines}}\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write template: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("usage: beet template new <name>")
	}
}

func openForEdit(path string) error {
	editor := strings.TrimSpace(os.Getenv("EDITOR"))
	if editor != "" {
		cmd := exec.Command(editor, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	if err := defaultBrowser.OpenFile(path); err != nil {
		return fmt.Errorf("open default app: %w", err)
	}
	if err := waitForContentFn(path); err != nil {
		return fmt.Errorf("await default edit: %w", err)
	}
	return nil
}

func handleGenerate(configDir string, args []string) error {
	fs := flag.NewFlagSet("beet", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: beet [flags] [intent|file]")
		fmt.Fprintln(fs.Output(), "\nFlags:")
		fs.PrintDefaults()
		fmt.Fprintln(fs.Output(), "\nCommands: beet templates | beet packs | beet doctor | beet config restore | beet pack [list|init|edit] | beet template new | beet completion [--shell bash|zsh]")
		fmt.Fprintln(fs.Output(), "Notes: packs are bootstrapped and selectable with -p/--pack (default, extended, comprehensive).")
		fmt.Fprintln(fs.Output(), "       generation renders all outputs defined by the pack; -t/--template only overrides WORK_PROMPT.md in the default pack.")
		fmt.Fprintln(fs.Output(), "       CLI execution defaults on (Codex preferred, then Copilot, then Claude Code). Disable with --exec=false.")
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

	if err := requireConfigState(configDir); err != nil {
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
	var execCtx context.Context
	var cancelExec context.CancelFunc
	if *execFlag {
		cli, err = requireCLI()
		if err != nil {
			return err
		}

		baseCtx, baseCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		timeoutCtx, timeoutCancel := context.WithTimeout(baseCtx, cliTimeout())
		execCtx = timeoutCtx
		cancelExec = func() {
			timeoutCancel()
			baseCancel()
		}
		defer cancelExec()
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
			if execCtx == nil {
				return fmt.Errorf("missing CLI context")
			}
			content, err = runCLI(execCtx, cli, prompt)
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
				if intent == "" {
					return "", fmt.Errorf("intent is empty; provide input")
				}
				return intent, nil
			}
		}
		intent := strings.TrimSpace(strings.Join(remaining, " "))
		if intent == "" {
			return "", fmt.Errorf("intent is empty; provide input")
		}
		return intent, nil
	}

	info, err := os.Stdin.Stat()
	if err == nil && (info.Mode()&os.ModeCharDevice) == 0 {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		intent := strings.TrimSpace(string(b))
		if intent == "" {
			return "", fmt.Errorf("intent is empty; provide input")
		}
		return intent, nil
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

	if err := waitForContentFn(tmp.Name()); err != nil {
		return "", fmt.Errorf("await editor content: %w", err)
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

	if err := waitForContentFn(tmp.Name()); err != nil {
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

func defaultWaitForContent(path string) error {
	deadline := time.Now().Add(waitForContentTimeout)
	for {
		data, err := os.ReadFile(path)
		if err == nil && strings.TrimSpace(string(data)) != "" {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("no content written to %s", path)
		}
		time.Sleep(waitForContentInterval)
	}
}
