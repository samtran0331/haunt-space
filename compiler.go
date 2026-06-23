package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// resolveFolder resolves a pane's Folder field to an absolute path.
// ""  → cwd (use whatever the split-surface --working-directory sets)
// "~" or "~/..." → homeDir (plus any suffix)
// relative → joined with cwd
// absolute → used as-is
func resolveFolder(folder, cwd, homeDir string) string {
	if folder == "" {
		return ""
	}
	if folder == "~" {
		return homeDir
	}
	if strings.HasPrefix(folder, "~/") {
		return filepath.Join(homeDir, folder[2:])
	}
	if filepath.IsAbs(folder) {
		return folder
	}
	return filepath.Join(cwd, folder)
}

// BuildGhosttyCommand performs a recursive pre-order traversal of the layout
// tree and produces the Ghostty split-surface CLI argument string.
// currentDir is the directory where hsp was invoked; homeDir is the user's home.
func BuildGhosttyCommand(node *LayoutNode, currentDir, homeDir string) string {
	if node == nil {
		return ""
	}

	// Leaf node — emit command and/or working-directory override.
	if node.Direction == None {
		folder := resolveFolder(node.Folder, currentDir, homeDir)

		if folder == "" && node.Command == "" {
			// No overrides needed; the split's --working-directory covers it.
			return ""
		}

		if folder == "" {
			// No folder override, just a command.
			return fmt.Sprintf(" --command=%q", node.Command)
		}

		// Custom folder: wrap in a shell so we can cd before exec.
		cmd := node.Command
		if cmd == "" {
			cmd = "$SHELL"
		}
		inner := fmt.Sprintf("cd %s && exec %s", shellQuotePath(folder), cmd)
		return fmt.Sprintf(" --command=%q", "sh -c '"+inner+"'")
	}

	// Split node.
	result := fmt.Sprintf(" split-surface --%s --working-directory=%q", node.Direction, currentDir)
	result += BuildGhosttyCommand(node.LeftChild, currentDir, homeDir)
	result += BuildGhosttyCommand(node.RightChild, currentDir, homeDir)
	return result
}

// shellQuotePath wraps a path in single quotes, escaping any embedded single quotes.
func shellQuotePath(p string) string {
	return "'" + strings.ReplaceAll(p, "'", "'\\''") + "'"
}
