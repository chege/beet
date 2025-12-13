package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	configDir, err := prepareConfig()
	if err != nil {
		log.Fatalf("prepare config: %v", err)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
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
		log.Fatalf("doctor not implemented yet")
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

	if err := fs.Parse(args); err != nil {
		return err
	}

	intent, err := parseIntent(fs.Args())
	if err != nil {
		return err
	}

	tmplName := firstNonEmpty(*template, *templateLong)
	return writeWorkPrompt(configDir, tmplName, intent, workPromptFilename)
}

func parseIntent(remaining []string) (string, error) {
	if len(remaining) > 0 {
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

	return "", fmt.Errorf("intent is required via args or stdin")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
