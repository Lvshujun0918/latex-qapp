package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromDotEnvInCurrentDir(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env")
	content := `PORT=9090
	JWT_SECRET=test-secret
	QWEN_API_KEY=test-key`
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env failed: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get cwd failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(cwd)
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("JWT_SECRET")
		_ = os.Unsetenv("QWEN_API_KEY")
	})

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cfg := Load()
	if cfg.Port != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Fatalf("expected jwt secret from .env")
	}
	if cfg.QwenAPIKey != "test-key" {
		t.Fatalf("expected qwen api key from .env")
	}
}
