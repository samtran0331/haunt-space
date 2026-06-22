package main

import "testing"

func TestBuildGhosttyCommand(t *testing.T) {
	mockTree := &LayoutNode{
		Direction: Vertical,
		Size:      30,
		LeftChild: &LayoutNode{Direction: None, Command: "lazygit"},
		RightChild: &LayoutNode{
			Direction: Horizontal,
			Size:      50,
			LeftChild:  &LayoutNode{Direction: None, Command: "claude"},
			RightChild: &LayoutNode{Direction: None, Command: ""},
		},
	}
	targetDir := "/users/test/project"

	gotResult := BuildGhosttyCommand(mockTree, targetDir)
	expected := ` split-surface --vertical --working-directory="/users/test/project" --command="lazygit" split-surface --horizontal --working-directory="/users/test/project" --command="claude"`

	if gotResult != expected {
		t.Errorf("Compiler layout mismatch.\nGot: %q\nExpected: %q", gotResult, expected)
	}
}

func TestBuildGhosttyCommandNilNode(t *testing.T) {
	if got := BuildGhosttyCommand(nil, "/tmp"); got != "" {
		t.Errorf("expected empty string for nil node, got %q", got)
	}
}

func TestBuildGhosttyCommandLeafNoCommand(t *testing.T) {
	leaf := &LayoutNode{Direction: None, Command: ""}
	if got := BuildGhosttyCommand(leaf, "/tmp"); got != "" {
		t.Errorf("expected empty string for leaf with no command, got %q", got)
	}
}

func TestBuildGhosttyCommandLeafWithCommand(t *testing.T) {
	leaf := &LayoutNode{Direction: None, Command: "vim"}
	expected := ` --command="vim"`
	if got := BuildGhosttyCommand(leaf, "/tmp"); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}
