package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// templateDir returns the path to the templates directory.
func templateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".config", "haunt-space", "templates")
}

func templatePath(name string) string {
	name = filepath.Base(name)
	return filepath.Join(templateDir(), name+".json")
}

// saveBlueprint serialises a GlobalBlueprint to its JSON template file,
// creating the templates directory if it does not exist.
func saveBlueprint(bp GlobalBlueprint) error {
	dir := templateDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create template directory: %w", err)
	}
	data, err := json.MarshalIndent(bp, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal blueprint: %w", err)
	}
	if err := os.WriteFile(templatePath(bp.TemplateName), data, 0o600); err != nil {
		return fmt.Errorf("write template file: %w", err)
	}
	return nil
}

// loadBlueprint reads and deserialises a template JSON file by name.
func loadBlueprint(name string) (GlobalBlueprint, error) {
	data, err := os.ReadFile(templatePath(name))
	if err != nil {
		return GlobalBlueprint{}, fmt.Errorf("read template %q: %w", name, err)
	}
	var bp GlobalBlueprint
	if err := json.Unmarshal(data, &bp); err != nil {
		return GlobalBlueprint{}, fmt.Errorf("parse template %q: %w", name, err)
	}
	return bp, nil
}

// listTemplateNames returns all saved template names sorted alphabetically.
func listTemplateNames() ([]string, error) {
	dir := templateDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return names, nil
}

// launchTemplate loads the named template, compiles it into a Ghostty command,
// and spawns Ghostty as a detached background process.
func launchTemplate(name string) error {
	bp, err := loadBlueprint(name)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = cwd
	}
	args := BuildGhosttyCommand(&bp.Root, cwd, homeDir)

	// Invoke ghostty via the shell so the pre-formatted argument string
	// (including embedded quotes) is parsed correctly by the shell.
	ghosttyCmd := "ghostty" + args
	cmd := exec.Command("sh", "-c", ghosttyCmd)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ghostty: %w", err)
	}

	// Detach: release the process so it outlives this parent process.
	if err := cmd.Process.Release(); err != nil {
		return fmt.Errorf("release ghostty process: %w", err)
	}

	fmt.Printf("Launched template %q in Ghostty (working directory: %s)\n", name, cwd)
	return nil
}
