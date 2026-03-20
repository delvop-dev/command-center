package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	General  GeneralConfig  `toml:"general"`
	Tmux     TmuxConfig     `toml:"tmux"`
	Worktree WorktreeConfig `toml:"worktree"`
	Notify   NotifyConfig   `toml:"notify"`
	Cost     CostConfig     `toml:"cost"`
}

type GeneralConfig struct {
	DefaultProvider  string `toml:"default_provider"`
	DefaultModel     string `toml:"default_model"`
	AutoCompact      bool   `toml:"auto_compact"`
	CompactThreshold int    `toml:"compact_threshold"`
	PollIntervalMs   int    `toml:"poll_interval_ms"`
}

type TmuxConfig struct {
	Prefix string `toml:"prefix"`
}

type WorktreeConfig struct {
	Enabled     bool   `toml:"enabled"`
	BaseDir     string `toml:"base_dir"`
	AutoCleanup bool   `toml:"auto_cleanup"`
}

type NotifyConfig struct {
	Channels      []string    `toml:"channels"`
	FocusSuppress bool        `toml:"focus_suppress"`
	DebounceMs    int         `toml:"debounce_ms"`
	Sound         SoundConfig `toml:"sound"`
}

type SoundConfig struct {
	InputNeeded string `toml:"input_needed"`
	TaskDone    string `toml:"task_done"`
	Error       string `toml:"error"`
}

type CostConfig struct {
	DailyBudget float64 `toml:"daily_budget"`
	LogFile     string  `toml:"log_file"`
}

func Default() *Config {
	return &Config{
		General: GeneralConfig{
			DefaultProvider:  "claude",
			DefaultModel:     "opus",
			AutoCompact:      true,
			CompactThreshold: 80,
			PollIntervalMs:   500,
		},
		Tmux: TmuxConfig{
			Prefix: "dv-",
		},
		Worktree: WorktreeConfig{
			Enabled:     false,
			BaseDir:     ".worktrees",
			AutoCleanup: true,
		},
		Notify: NotifyConfig{
			Channels:      []string{},
			FocusSuppress: true,
			DebounceMs:    500,
			Sound: SoundConfig{
				InputNeeded: "Basso",
				TaskDone:    "Glass",
				Error:       "Sosumi",
			},
		},
		Cost: CostConfig{
			DailyBudget: 50.0,
			LogFile:     "~/.delvop/cost_log.jsonl",
		},
	}
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".delvop")
}

func DefaultPath() string {
	return filepath.Join(configDir(), "config.toml")
}

func Load() (*Config, error) {
	return LoadFrom(DefaultPath())
}

func LoadFrom(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
