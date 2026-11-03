package main

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed skills/painel-config/SKILL.md
var configSkill string

// installSkill é o `painel skill`: instala o assistente de configuração para
// o Claude Code, ou imprime o texto com --print para colar em outro agente.
func installSkill(args []string) error {
	if len(args) > 0 && args[0] == "--print" {
		fmt.Print(configSkill)
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".claude", "skills", "painel-config", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(configSkill), 0o644); err != nil {
		return err
	}
	fmt.Printf("skill instalada em %s\nno Claude Code, peça \"configura o painel\" ou rode /painel-config\n", path)
	return nil
}
