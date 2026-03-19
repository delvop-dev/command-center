package templates

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

//go:embed builtin/*.toml
var builtinFS embed.FS

type Template struct {
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Sessions    []SessionTemplate `toml:"sessions"`
}

type SessionTemplate struct {
	Name          string `toml:"name"`
	Provider      string `toml:"provider"`
	Model         string `toml:"model"`
	InitialPrompt string `toml:"initial_prompt"`
}

func LoadFile(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t Template
	if _, err := toml.Decode(string(data), &t); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &t, nil
}

func LoadBuiltins() ([]*Template, error) {
	entries, err := builtinFS.ReadDir("builtin")
	if err != nil {
		return nil, err
	}
	var templates []*Template
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		data, err := builtinFS.ReadFile(filepath.Join("builtin", entry.Name()))
		if err != nil {
			return nil, err
		}
		var t Template
		if _, err := toml.Decode(string(data), &t); err != nil {
			return nil, fmt.Errorf("parse builtin/%s: %w", entry.Name(), err)
		}
		templates = append(templates, &t)
	}
	return templates, nil
}

func LoadFromDir(dir string) ([]*Template, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var templates []*Template
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		t, err := LoadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}
