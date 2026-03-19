package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()
	if cfg.General.DefaultProvider != "claude" {
		t.Errorf("expected default provider 'claude', got %q", cfg.General.DefaultProvider)
	}
	if cfg.General.DefaultModel != "opus" {
		t.Errorf("expected default model 'opus', got %q", cfg.General.DefaultModel)
	}
	if cfg.General.AutoCompact != true {
		t.Error("expected auto_compact true")
	}
	if cfg.General.CompactThreshold != 80 {
		t.Errorf("expected compact threshold 80, got %d", cfg.General.CompactThreshold)
	}
	if cfg.General.PollIntervalMs != 500 {
		t.Errorf("expected poll interval 500, got %d", cfg.General.PollIntervalMs)
	}
	if cfg.Tmux.Prefix != "dv-" {
		t.Errorf("expected tmux prefix 'dv-', got %q", cfg.Tmux.Prefix)
	}
	if cfg.Worktree.Enabled != false {
		t.Error("expected worktree disabled")
	}
	if cfg.Worktree.BaseDir != ".worktrees" {
		t.Errorf("expected worktree base dir '.worktrees', got %q", cfg.Worktree.BaseDir)
	}
	if cfg.Worktree.AutoCleanup != true {
		t.Error("expected worktree auto cleanup true")
	}
	if cfg.Notify.DebounceMs != 500 {
		t.Errorf("expected debounce 500, got %d", cfg.Notify.DebounceMs)
	}
	if cfg.Notify.FocusSuppress != true {
		t.Error("expected focus suppress true")
	}
	if len(cfg.Notify.Channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(cfg.Notify.Channels))
	}
	if cfg.Notify.Sound.InputNeeded != "Basso" {
		t.Errorf("expected sound 'Basso', got %q", cfg.Notify.Sound.InputNeeded)
	}
	if cfg.Notify.Sound.TaskDone != "Glass" {
		t.Errorf("expected sound 'Glass', got %q", cfg.Notify.Sound.TaskDone)
	}
	if cfg.Notify.Sound.Error != "Sosumi" {
		t.Errorf("expected sound 'Sosumi', got %q", cfg.Notify.Sound.Error)
	}
	if cfg.Cost.DailyBudget != 50.0 {
		t.Errorf("expected budget 50.0, got %f", cfg.Cost.DailyBudget)
	}
	if cfg.Cost.LogFile != "~/.delvop/cost_log.jsonl" {
		t.Errorf("expected log file '~/.delvop/cost_log.jsonl', got %q", cfg.Cost.LogFile)
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
	// Defaults should still be present for fields not overridden
	if cfg.Notify.DebounceMs != 500 {
		t.Errorf("expected default debounce 500, got %d", cfg.Notify.DebounceMs)
	}
}

func TestLoadFromMissingFile(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	// Should return defaults
	if cfg.General.DefaultProvider != "claude" {
		t.Errorf("expected default provider 'claude', got %q", cfg.General.DefaultProvider)
	}
	if cfg.Tmux.Prefix != "dv-" {
		t.Errorf("expected default prefix 'dv-', got %q", cfg.Tmux.Prefix)
	}
}

func TestLoadFromInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	err := os.WriteFile(path, []byte(`this is not valid toml {{{`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	_, err = LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
}

func TestLoadFromUnreadableFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unreadable.toml")
	err := os.WriteFile(path, []byte(`[general]`), 0644)
	if err != nil {
		t.Fatal(err)
	}
	os.Chmod(path, 0000)
	defer os.Chmod(path, 0644)

	_, err = LoadFrom(path)
	if err == nil {
		t.Fatal("expected error for unreadable file")
	}
}

func TestConfigDir(t *testing.T) {
	dir := configDir()
	if dir == "" {
		t.Error("expected non-empty config dir")
	}
	if !strings.HasSuffix(dir, ".delvop") {
		t.Errorf("expected config dir to end with '.delvop', got %q", dir)
	}
}

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	if path == "" {
		t.Error("expected non-empty default path")
	}
	if !strings.HasSuffix(path, "config.toml") {
		t.Errorf("expected path to end with 'config.toml', got %q", path)
	}
	if !strings.Contains(path, ".delvop") {
		t.Errorf("expected path to contain '.delvop', got %q", path)
	}
}

func TestLoad(t *testing.T) {
	// Load() uses DefaultPath() which may or may not exist
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should not error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestLoadFromPartialConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "partial.toml")
	err := os.WriteFile(path, []byte(`
[cost]
daily_budget = 100.0
log_file = "/tmp/costs.jsonl"
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Cost.DailyBudget != 100.0 {
		t.Errorf("expected budget 100.0, got %f", cfg.Cost.DailyBudget)
	}
	if cfg.Cost.LogFile != "/tmp/costs.jsonl" {
		t.Errorf("expected log file '/tmp/costs.jsonl', got %q", cfg.Cost.LogFile)
	}
	// Defaults for other sections should remain
	if cfg.General.DefaultProvider != "claude" {
		t.Errorf("expected default provider 'claude', got %q", cfg.General.DefaultProvider)
	}
}
