package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstalaSkill(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := installSkill(nil); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(home, ".claude", "skills", "painel-config", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), "---\nname: painel-config\n") {
		t.Errorf("a skill deveria começar com o frontmatter:\n%.80s", b)
	}
}
