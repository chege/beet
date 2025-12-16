package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalLLMBackendRun(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "cfg")
	backend := newLocalLLMBackend(configDir)
	if err := backend.ensure(); err != nil {
		t.Fatalf("ensure backend: %v", err)
	}

	privateLog := filepath.Join(configDir, "local-llm", "prompt.log")
	script := `#!/usr/bin/env sh
	cat > "${BEET_LOCAL_LLM_PROMPT_LOG:-` + privateLog + `}"
	printf "llm-response" 
`
	if err := os.WriteFile(backend.runner, []byte(script), 0o755); err != nil {
		t.Fatalf("override runner: %v", err)
	}

	t.Setenv("BEET_LOCAL_LLM_PROMPT_LOG", privateLog)
	ctx := context.Background()
	response, err := backend.run(ctx, "placeholder prompt")
	if err != nil {
		t.Fatalf("backend run: %v", err)
	}
	if response != "llm-response" {
		t.Fatalf("unexpected response %q", response)
	}

	data, err := os.ReadFile(privateLog)
	if err != nil {
		t.Fatalf("read prompt log: %v", err)
	}
	if string(data) != "placeholder prompt" {
		t.Fatalf("prompt log mismatch: %q", string(data))
	}
}
