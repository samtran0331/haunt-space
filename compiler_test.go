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
	homeDir := "/users/test"

	gotResult := BuildGhosttyCommand(mockTree, targetDir, homeDir)
	expected := ` split-surface --vertical --working-directory="/users/test/project" --command="lazygit" split-surface --horizontal --working-directory="/users/test/project" --command="claude"`

	if gotResult != expected {
		t.Errorf("Compiler layout mismatch.\nGot: %q\nExpected: %q", gotResult, expected)
	}
}

func TestBuildGhosttyCommandNilNode(t *testing.T) {
	if got := BuildGhosttyCommand(nil, "/tmp", "/home"); got != "" {
		t.Errorf("expected empty string for nil node, got %q", got)
	}
}

func TestBuildGhosttyCommandLeafNoCommand(t *testing.T) {
	leaf := &LayoutNode{Direction: None, Command: ""}
	if got := BuildGhosttyCommand(leaf, "/tmp", "/home"); got != "" {
		t.Errorf("expected empty string for leaf with no command or folder, got %q", got)
	}
}

func TestBuildGhosttyCommandLeafWithCommand(t *testing.T) {
	leaf := &LayoutNode{Direction: None, Command: "vim"}
	expected := ` --command="vim"`
	if got := BuildGhosttyCommand(leaf, "/tmp", "/home"); got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestBuildGhosttyCommandLeafWithFolder(t *testing.T) {
	leaf := &LayoutNode{Direction: None, Folder: "/custom/path", Command: "vim"}
	got := BuildGhosttyCommand(leaf, "/tmp", "/home")
	expected := ` --command="sh -c 'cd '/custom/path' && exec vim'"`
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestBuildGhosttyCommandLeafHomeFolder(t *testing.T) {
	leaf := &LayoutNode{Direction: None, Folder: "~", Command: "vim"}
	got := BuildGhosttyCommand(leaf, "/tmp", "/home/user")
	expected := ` --command="sh -c 'cd '/home/user' && exec vim'"`
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}

func TestBuildGhosttyCommandLeafFolderNoCommand(t *testing.T) {
	leaf := &LayoutNode{Direction: None, Folder: "/custom/path", Command: ""}
	got := BuildGhosttyCommand(leaf, "/tmp", "/home")
	expected := ` --command="sh -c 'cd '/custom/path' && exec $SHELL'"`
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}
