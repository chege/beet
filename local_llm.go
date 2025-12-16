package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	localLLMRelDir        = "local-llm"
	localLLMRunnerRelPath = "local-llm/runner.sh"
	localLLMModelRelPath  = "local-llm/model.gguf"
)

type localLLMBackend struct {
	configDir string
	baseDir   string
	runner    string
	model     string
}

func newLocalLLMBackend(configDir string) *localLLMBackend {
	base := filepath.Join(configDir, localLLMRelDir)
	return &localLLMBackend{
		configDir: configDir,
		baseDir:   base,
		runner:    filepath.Join(base, "runner.sh"),
		model:     filepath.Join(base, "model.gguf"),
	}
}

func (b *localLLMBackend) ensure() error {
	if err := os.MkdirAll(b.baseDir, 0o755); err != nil {
		return fmt.Errorf("create local llm dir: %w", err)
	}
	if _, err := os.Stat(b.runner); err != nil {
		if err := copyDefaultContent(localLLMRunnerRelPath, b.runner, 0o755); err != nil {
			return err
		}
	}
	if _, err := os.Stat(b.model); err != nil {
		if err := copyDefaultContent(localLLMModelRelPath, b.model, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func (b *localLLMBackend) run(ctx context.Context, prompt string) (string, error) {
	if err := b.ensure(); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "/bin/sh", b.runner)
	cmd.Env = append(os.Environ(), fmt.Sprintf("BEET_LOCAL_LLM_MODEL=%s", b.model))
	cmd.Stdin = strings.NewReader(prompt)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("local llm runner: %w", err)
	}
	return out.String(), nil
}

func copyDefaultContent(relPath, dest string, mode os.FileMode) error {
	src := filepath.Join("defaults", relPath)
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read bundled %s: %w", relPath, err)
	}
	if err := os.WriteFile(dest, data, mode); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	if err := os.Chmod(dest, mode); err != nil {
		return fmt.Errorf("chmod %s: %w", dest, err)
	}
	return nil
}
