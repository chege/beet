package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

type detectedCLI struct {
	name string
	path string
}

var cliPriority = []string{"codex", "copilot", "claude"}

func detectPreferredCLI() (detectedCLI, bool) {
	for _, name := range cliPriority {
		if path, err := exec.LookPath(name); err == nil {
			return detectedCLI{name: name, path: path}, true
		}
	}
	return detectedCLI{}, false
}

func detectAllCLIs() []detectedCLI {
	var out []detectedCLI
	for _, name := range cliPriority {
		if path, err := exec.LookPath(name); err == nil {
			out = append(out, detectedCLI{name: name, path: path})
		}
	}
	return out
}

func runDoctor(w io.Writer) error {
	found := detectAllCLIs()
	for _, name := range cliPriority {
		if _, err := fmt.Fprintf(w, "%s: ", name); err != nil {
			return err
		}
		path := ""
		for _, cli := range found {
			if cli.name == name {
				path = cli.path
				break
			}
		}
		if path == "" {
			if _, err := fmt.Fprintln(w, "not found"); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "found at %s\n", path); err != nil {
				return err
			}
		}
	}

	if len(found) == 0 {
		if _, err := fmt.Fprintln(w, "No supported CLI detected. Install Codex CLI, Copilot CLI, or Claude Code CLI."); err != nil {
			return err
		}
	}

	return nil
}

func runDetectedCLI(promptPath string) error {
	cli, ok := detectPreferredCLI()
	if !ok {
		return fmt.Errorf("no supported CLI found; install Codex CLI, Copilot CLI, or Claude Code CLI")
	}

	cmd := exec.Command(cli.path, promptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s exec failed: %w", cli.name, err)
	}

	return nil
}

func requireCLI() (detectedCLI, error) {
	cli, ok := detectPreferredCLI()
	if !ok {
		return detectedCLI{}, fmt.Errorf("no supported CLI found; install Codex CLI, Copilot CLI, or Claude Code CLI")
	}
	return cli, nil
}
