package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")
	os.WriteFile(path, []byte(`
name = "test-template"
description = "A test template"

[[sessions]]
name = "worker"
provider = "claude"
model = "opus"
initial_prompt = "Do the work"
`), 0644)

	tmpl, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.Name != "test-template" {
		t.Errorf("expected 'test-template', got %q", tmpl.Name)
	}
	if len(tmpl.Sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(tmpl.Sessions))
	}
	if tmpl.Sessions[0].Provider != "claude" {
		t.Errorf("expected provider 'claude', got %q", tmpl.Sessions[0].Provider)
	}
	if tmpl.Sessions[0].InitialPrompt != "Do the work" {
		t.Errorf("wrong initial_prompt: %q", tmpl.Sessions[0].InitialPrompt)
	}
}

func TestLoadBuiltins(t *testing.T) {
	templates, err := LoadBuiltins()
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) < 2 {
		t.Errorf("expected at least 2 builtin templates, got %d", len(templates))
	}
	names := make(map[string]bool)
	for _, tmpl := range templates {
		names[tmpl.Name] = true
	}
	if !names["solo"] {
		t.Error("missing 'solo' template")
	}
	if !names["fullstack"] {
		t.Error("missing 'fullstack' template")
	}
}

func TestLoadFromDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.toml"), []byte(`name = "a"
description = "Template A"
[[sessions]]
name = "s1"
provider = "claude"
model = "opus"
`), 0644)
	os.WriteFile(filepath.Join(dir, "b.toml"), []byte(`name = "b"
description = "Template B"
[[sessions]]
name = "s1"
provider = "gemini"
model = "pro"
`), 0644)

	templates, err := LoadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 2 {
		t.Errorf("expected 2 templates, got %d", len(templates))
	}
}

func TestLoadFromNonexistentDir(t *testing.T) {
	templates, err := LoadFromDir("/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if templates != nil {
		t.Error("expected nil for nonexistent dir")
	}
}
