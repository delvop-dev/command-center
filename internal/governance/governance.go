package governance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Governance struct {
	Security SecurityConfig `toml:"security"`
	Project  ProjectConfig  `toml:"project"`
	Skills   []Skill        `toml:"skills"`
}

type SecurityConfig struct {
	DisabledRules      []string `toml:"disabled_rules"`
	CustomAllowedHosts []string `toml:"custom_allowed_hosts"`
}

type ProjectConfig struct {
	Language         string `toml:"language"`
	TestBeforeCommit bool   `toml:"test_before_commit"`
	NoCommitToMain   bool   `toml:"no_commit_to_main"`
	LintOnSave       bool   `toml:"lint_on_save"`
	MaxFileSizeKB    int    `toml:"max_file_size_kb"`
}

type Skill struct {
	Name        string `toml:"name"`
	Instruction string `toml:"instruction"`
}

func Default() *Governance {
	return &Governance{}
}

func Load() (*Governance, error) {
	gov := Default()

	home, _ := os.UserHomeDir()
	globalPath := filepath.Join(home, ".delvop", "governance.toml")
	if err := loadFile(globalPath, gov); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("global governance: %w", err)
	}

	localPath := filepath.Join(".delvop", "governance.toml")
	if err := loadFile(localPath, gov); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("local governance: %w", err)
	}

	return gov, nil
}

func loadFile(path string, gov *Governance) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	_, err = toml.Decode(string(data), gov)
	return err
}

func (g *Governance) IsRuleDisabled(ruleID string) bool {
	for _, id := range g.Security.DisabledRules {
		if id == ruleID {
			return true
		}
	}
	return false
}

func (g *Governance) IsHostAllowed(host string) bool {
	for _, h := range g.Security.CustomAllowedHosts {
		if h == host {
			return true
		}
	}
	return false
}

func (g *Governance) BuildContext() string {
	var parts []string

	if g.Project.Language != "" || g.Project.TestBeforeCommit || g.Project.NoCommitToMain || g.Project.LintOnSave {
		parts = append(parts, "Project rules:")
		if g.Project.Language != "" {
			parts = append(parts, fmt.Sprintf("- Language: %s", g.Project.Language))
		}
		if g.Project.TestBeforeCommit {
			parts = append(parts, "- Always run tests before committing")
		}
		if g.Project.NoCommitToMain {
			parts = append(parts, "- Never commit directly to main branch")
		}
		if g.Project.LintOnSave {
			parts = append(parts, "- Run linter before saving files")
		}
		if g.Project.MaxFileSizeKB > 0 {
			parts = append(parts, fmt.Sprintf("- Keep files under %d KB", g.Project.MaxFileSizeKB))
		}
	}

	if len(g.Skills) > 0 {
		parts = append(parts, "")
		parts = append(parts, "Shared team conventions:")
		for _, s := range g.Skills {
			parts = append(parts, fmt.Sprintf("- %s: %s", s.Name, s.Instruction))
		}
	}

	return strings.Join(parts, "\n")
}

func (g *Governance) HasContent() bool {
	return g.Project.Language != "" ||
		g.Project.TestBeforeCommit ||
		g.Project.NoCommitToMain ||
		g.Project.LintOnSave ||
		len(g.Skills) > 0
}

func (g *Governance) ActiveRuleCount() int {
	total := 17
	return total - len(g.Security.DisabledRules)
}
