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

var cliPriority = []string{"codex", "copilot"}

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
		fmt.Fprintf(w, "%s: ", name)
		path := ""
		for _, cli := range found {
			if cli.name == name {
				path = cli.path
				break
			}
		}
		if path == "" {
			fmt.Fprintln(w, "not found")
		} else {
			fmt.Fprintf(w, "found at %s\n", path)
		}
	}

	if len(found) == 0 {
		fmt.Fprintln(w, "No supported CLI detected. Install Codex CLI or Copilot CLI.")
	}

	return nil
}

func runDetectedCLI(promptPath string) error {
	cli, ok := detectPreferredCLI()
	if !ok {
		return fmt.Errorf("no supported CLI found; install Codex CLI or Copilot CLI")
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
