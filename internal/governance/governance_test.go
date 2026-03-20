package governance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	gov := Default()
	if gov.Project.Language != "" {
		t.Error("expected empty default language")
	}
	if len(gov.Skills) != 0 {
		t.Error("expected no default skills")
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "governance.toml")
	content := `
[security]
disabled_rules = ["SUP001"]
custom_allowed_hosts = ["registry.npmjs.org"]

[project]
language = "typescript"
test_before_commit = true
no_commit_to_main = true

[[skills]]
name = "code-style"
instruction = "Use functional React components"

[[skills]]
name = "git"
instruction = "Write conventional commits"
`
	os.WriteFile(path, []byte(content), 0644)

	gov := Default()
	if err := loadFile(path, gov); err != nil {
		t.Fatalf("loadFile failed: %v", err)
	}

	if len(gov.Security.DisabledRules) != 1 || gov.Security.DisabledRules[0] != "SUP001" {
		t.Errorf("unexpected disabled rules: %v", gov.Security.DisabledRules)
	}
	if gov.Project.Language != "typescript" {
		t.Errorf("expected language 'typescript', got %q", gov.Project.Language)
	}
	if !gov.Project.TestBeforeCommit {
		t.Error("expected test_before_commit true")
	}
	if len(gov.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(gov.Skills))
	}
}

func TestIsRuleDisabled(t *testing.T) {
	gov := &Governance{Security: SecurityConfig{DisabledRules: []string{"SUP001", "ESC001"}}}
	if !gov.IsRuleDisabled("SUP001") {
		t.Error("SUP001 should be disabled")
	}
	if gov.IsRuleDisabled("EXF001") {
		t.Error("EXF001 should not be disabled")
	}
}

func TestIsHostAllowed(t *testing.T) {
	gov := &Governance{Security: SecurityConfig{CustomAllowedHosts: []string{"registry.npmjs.org"}}}
	if !gov.IsHostAllowed("registry.npmjs.org") {
		t.Error("registry.npmjs.org should be allowed")
	}
	if gov.IsHostAllowed("evil.com") {
		t.Error("evil.com should not be allowed")
	}
}

func TestBuildContext(t *testing.T) {
	gov := &Governance{
		Project: ProjectConfig{
			Language:         "typescript",
			TestBeforeCommit: true,
			NoCommitToMain:   true,
		},
		Skills: []Skill{
			{Name: "style", Instruction: "Use tabs"},
		},
	}
	ctx := gov.BuildContext()
	if ctx == "" {
		t.Fatal("expected non-empty context")
	}
	if !strings.Contains(ctx, "typescript") {
		t.Error("context should mention language")
	}
	if !strings.Contains(ctx, "Use tabs") {
		t.Error("context should include skill instruction")
	}
}

func TestHasContent(t *testing.T) {
	if Default().HasContent() {
		t.Error("default governance should have no content")
	}
	gov := &Governance{Project: ProjectConfig{Language: "go"}}
	if !gov.HasContent() {
		t.Error("governance with language should have content")
	}
}

func TestActiveRuleCount(t *testing.T) {
	gov := Default()
	if gov.ActiveRuleCount() != 17 {
		t.Errorf("expected 17 active rules, got %d", gov.ActiveRuleCount())
	}
	gov.Security.DisabledRules = []string{"SUP001", "ESC001"}
	if gov.ActiveRuleCount() != 15 {
		t.Errorf("expected 15 active rules, got %d", gov.ActiveRuleCount())
	}
}
