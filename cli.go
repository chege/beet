package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
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

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: pf [input] | pf -t <template> | pf templates | pf doctor")
}

func handleGenerate(configDir string, args []string) error {
	fs := flag.NewFlagSet("pf", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	template := fs.String("t", "", "template name")
	templateLong := fs.String("template", "", "template name")
	dryRun := fs.Bool("dry-run", false, "render without writing files")
	forceAgents := fs.Bool("force-agents", false, "overwrite agents.md")
	execFlag := fs.Bool("exec", false, "execute detected CLI with WORK_PROMPT.md")

	if err := fs.Parse(args); err != nil {
		return err
	}

	intent, err := parseIntent(fs.Args())
	if err != nil {
		return err
	}

	tmplName := firstNonEmpty(*template, *templateLong)
	if *dryRun {
		content, err := buildWorkPromptContent(configDir, tmplName, intent)
		if err != nil {
			return err
		}
		fmt.Println(content)
		return nil
	}

	if err := writeAgents(configDir, agentsFilename, *forceAgents); err != nil {
		return err
	}

	if err := writeWorkPrompt(configDir, tmplName, intent, workPromptFilename); err != nil {
		return err
	}

	if *execFlag {
		if err := runDetectedCLI(workPromptFilename); err != nil {
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
		return "", fmt.Errorf("intent is required via args or stdin; set EDITOR to use an editor")
	}

	tmp, err := os.CreateTemp("", "pf-intent-*.md")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

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
