package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type detectedCLI struct {
	name string
	path string
}

const envCLIBinary = "BEET_CLI_PATH"

var cliPriority = []string{"codex", "copilot", "claude"}

func detectPreferredCLI() (detectedCLI, bool) {
	for _, name := range cliPriority {
		path, err := exec.LookPath(name)
		if err == nil {
			logVerbose("preferred CLI %s found at %s", name, path)
			return detectedCLI{name: name, path: path}, true
		}
		logVerbose("preferred CLI %s not found: %v", name, err)
	}
	return detectedCLI{}, false
}

func detectAllCLIs() []detectedCLI {
	var out []detectedCLI
	for _, name := range cliPriority {
		path, err := exec.LookPath(name)
		if err == nil {
			logVerbose("detected CLI %s at %s", name, path)
			out = append(out, detectedCLI{name: name, path: path})
			continue
		}
		logVerbose("CLI %s unavailable: %v", name, err)
	}
	return out
}

func runDoctor(w io.Writer) error {
	logVerbose("running doctor diagnostics")
	found := detectAllCLIs()
	logVerbose("detected %d CLI candidates", len(found))
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

	if override, ok, err := detectCLIOverride(); err != nil {
		if _, writeErr := fmt.Fprintf(w, "%s: %v\n", envCLIBinary, err); writeErr != nil {
			return writeErr
		}
	} else if ok {
		if _, writeErr := fmt.Fprintf(w, "%s override: %s at %s\n", envCLIBinary, override.name, override.path); writeErr != nil {
			return writeErr
		}
	}

	if len(found) == 0 {
		if _, err := fmt.Fprintln(w, "No supported CLI detected. Install Codex CLI, Copilot CLI, or Claude Code CLI."); err != nil {
			return err
		}
	}

	return nil
}

func requireCLI() (detectedCLI, error) {
	if override, ok, err := detectCLIOverride(); err != nil {
		return detectedCLI{}, err
	} else if ok {
		logVerbose("using CLI override %s (%s)", override.name, override.path)
		return override, nil
	}
	cli, ok := detectPreferredCLI()
	if !ok {
		return detectedCLI{}, fmt.Errorf("no supported CLI found; install Codex CLI, Copilot CLI, or Claude Code CLI")
	}
	logVerbose("selected CLI %s (%s)", cli.name, cli.path)
	return cli, nil
}

func detectCLIOverride() (detectedCLI, bool, error) {
	raw := strings.TrimSpace(os.Getenv(envCLIBinary))
	if raw == "" {
		return detectedCLI{}, false, nil
	}
	logVerbose("%s override requested: %s", envCLIBinary, raw)
	path, err := exec.LookPath(raw)
	if err != nil {
		logVerbose("%s lookup failed: %v", envCLIBinary, err)
		return detectedCLI{}, false, fmt.Errorf("%s lookup failed: %w", envCLIBinary, err)
	}
	logVerbose("%s override resolved to %s", envCLIBinary, path)
	return detectedCLI{name: filepath.Base(path), path: path}, true, nil
}
