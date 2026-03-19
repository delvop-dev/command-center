package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.General.DefaultProvider != "claude" {
		t.Errorf("expected default provider 'claude', got %q", cfg.General.DefaultProvider)
	}
	if cfg.General.PollIntervalMs != 500 {
		t.Errorf("expected poll interval 500, got %d", cfg.General.PollIntervalMs)
	}
	if cfg.Tmux.Prefix != "dv-" {
		t.Errorf("expected tmux prefix 'dv-', got %q", cfg.Tmux.Prefix)
	}
	if cfg.Notify.DebounceMs != 500 {
		t.Errorf("expected debounce 500, got %d", cfg.Notify.DebounceMs)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	err := os.WriteFile(path, []byte(`
[general]
default_provider = "gemini"
default_model = "2.5-pro"
poll_interval_ms = 1000

[tmux]
prefix = "test-"
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.General.DefaultProvider != "gemini" {
		t.Errorf("expected provider 'gemini', got %q", cfg.General.DefaultProvider)
	}
	if cfg.General.DefaultModel != "2.5-pro" {
		t.Errorf("expected model '2.5-pro', got %q", cfg.General.DefaultModel)
	}
	if cfg.General.PollIntervalMs != 1000 {
		t.Errorf("expected poll 1000, got %d", cfg.General.PollIntervalMs)
	}
	if cfg.Tmux.Prefix != "test-" {
		t.Errorf("expected prefix 'test-', got %q", cfg.Tmux.Prefix)
	}
	if cfg.Notify.DebounceMs != 500 {
		t.Errorf("expected default debounce 500, got %d", cfg.Notify.DebounceMs)
	}
}
