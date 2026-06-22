package main

import "fmt"

// BuildGhosttyArgs performs a recursive pre-order traversal of the layout tree
// and produces Ghostty CLI arguments as an argv slice.
func BuildGhosttyArgs(node *LayoutNode, currentDir string) []string {
	if node == nil {
		return nil
	}

	// Base Case: Terminal Leaf Node
	if node.Direction == None {
		if node.Command != "" {
			return []string{"--command=" + node.Command}
		}
		return nil
	}

	// Recursive Case: Partition the surface frame
	args := []string{
		"split-surface",
		fmt.Sprintf("--%s", node.Direction),
		fmt.Sprintf("--working-directory=%s", currentDir),
	}

	if node.LeftChild != nil {
		args = append(args, BuildGhosttyArgs(node.LeftChild, currentDir)...)
	}
	if node.RightChild != nil {
		args = append(args, BuildGhosttyArgs(node.RightChild, currentDir)...)
	}

	return args
}

// BuildGhosttyCommand performs a recursive pre-order traversal of the layout
// tree and produces the Ghostty split-surface CLI argument string.
func BuildGhosttyCommand(node *LayoutNode, currentDir string) string {
	if node == nil {
		return ""
	}

	// Base Case: Terminal Leaf Node
	if node.Direction == None {
		if node.Command != "" {
			return fmt.Sprintf(" --command=%q", node.Command)
		}
		return ""
	}

	// Recursive Case: Partition the surface frame
	cmd := fmt.Sprintf(" split-surface --%s --working-directory=%q", node.Direction, currentDir)

	if node.LeftChild != nil {
		cmd += BuildGhosttyCommand(node.LeftChild, currentDir)
	}
	if node.RightChild != nil {
		cmd += BuildGhosttyCommand(node.RightChild, currentDir)
	}

	return cmd
}
