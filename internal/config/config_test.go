package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigLoadPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")
	content := []byte("api_key = \"key_from_file\"\nbase_url = \"https://example.com\"\n")
	if err := os.WriteFile(cfgPath, content, 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// 1. From file only
	t.Setenv("JULES_API_KEY", "")
	t.Setenv("JULES_BASE_URL", "")
	cfg, err := Load(cfgPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "key_from_file" {
		t.Errorf("expected key_from_file, got %s", cfg.APIKey)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", cfg.BaseURL)
	}

	// 2. Env overrides file
	t.Setenv("JULES_API_KEY", "key_from_env")
	t.Setenv("JULES_BASE_URL", "https://env.example.com")
	cfg, err = Load(cfgPath, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "key_from_env" {
		t.Errorf("expected key_from_env, got %s", cfg.APIKey)
	}
	if cfg.BaseURL != "https://env.example.com" {
		t.Errorf("expected https://env.example.com, got %s", cfg.BaseURL)
	}

	// 3. Flag overrides env and file
	cfg, err = Load(cfgPath, "key_from_flag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "key_from_flag" {
		t.Errorf("expected key_from_flag, got %s", cfg.APIKey)
	}
}
