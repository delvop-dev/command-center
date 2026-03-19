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
	if tmpl.Description != "A test template" {
		t.Errorf("expected description 'A test template', got %q", tmpl.Description)
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
	if tmpl.Sessions[0].Name != "worker" {
		t.Errorf("expected session name 'worker', got %q", tmpl.Sessions[0].Name)
	}
	if tmpl.Sessions[0].Model != "opus" {
		t.Errorf("expected model 'opus', got %q", tmpl.Sessions[0].Model)
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := LoadFile("/nonexistent/path.toml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadFileInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	os.WriteFile(path, []byte(`this is {{{ not valid toml`), 0644)

	_, err := LoadFile(path)
	if err == nil {
		t.Error("expected error for invalid TOML")
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
		if tmpl.Name == "" {
			t.Error("template has empty name")
		}
		if len(tmpl.Sessions) == 0 {
			t.Errorf("template %q has no sessions", tmpl.Name)
		}
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
	// Non-TOML file should be ignored
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte(`# readme`), 0644)

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

func TestLoadFromDirWithInvalidTOML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.toml"), []byte(`{{{ invalid`), 0644)

	_, err := LoadFromDir(dir)
	if err == nil {
		t.Error("expected error for dir with invalid TOML")
	}
}

func TestLoadFromEmptyDir(t *testing.T) {
	dir := t.TempDir()

	templates, err := LoadFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	// No TOML files = nil result
	if templates != nil {
		t.Errorf("expected nil for empty dir, got %d templates", len(templates))
	}
}

func TestLoadFromDirUnreadable(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "unreadable")
	os.Mkdir(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "a.toml"), []byte(`name = "a"`), 0644)
	os.Chmod(subdir, 0000)
	defer os.Chmod(subdir, 0755)

	_, err := LoadFromDir(subdir)
	if err == nil {
		t.Error("expected error for unreadable dir")
	}
}

func TestLoadFileMultipleSessions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.toml")
	os.WriteFile(path, []byte(`
name = "multi"
description = "Multiple sessions"

[[sessions]]
name = "frontend"
provider = "claude"
model = "opus"
initial_prompt = "Build UI"

[[sessions]]
name = "backend"
provider = "gemini"
model = "pro"
initial_prompt = "Build API"
`), 0644)

	tmpl, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(tmpl.Sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(tmpl.Sessions))
	}
	if tmpl.Sessions[0].Name != "frontend" {
		t.Errorf("expected first session 'frontend', got %q", tmpl.Sessions[0].Name)
	}
	if tmpl.Sessions[1].Name != "backend" {
		t.Errorf("expected second session 'backend', got %q", tmpl.Sessions[1].Name)
	}
}
