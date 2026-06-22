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

func TestBuildGhosttyArgs(t *testing.T) {
	mockTree := &LayoutNode{
		Direction: Vertical,
		Size:      30,
		LeftChild: &LayoutNode{Direction: None, Command: "lazygit"},
		RightChild: &LayoutNode{
			Direction: Horizontal,
			Size:      50,
			LeftChild:  &LayoutNode{Direction: None, Command: "claude --resume"},
			RightChild: &LayoutNode{Direction: None, Command: ""},
		},
	}

	got := BuildGhosttyArgs(mockTree, "/users/test/project")
	want := []string{
		"split-surface",
		"--vertical",
		"--working-directory=/users/test/project",
		"--command=lazygit",
		"split-surface",
		"--horizontal",
		"--working-directory=/users/test/project",
		"--command=claude --resume",
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected arg count: got %d, want %d\nGot: %#v\nWant: %#v", len(got), len(want), got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected arg at index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestBuildGhosttyArgsNilNode(t *testing.T) {
	if got := BuildGhosttyArgs(nil, "/tmp"); got != nil {
		t.Errorf("expected nil args for nil node, got %#v", got)
	}
}

func TestBuildGhosttyArgsLeafNoCommand(t *testing.T) {
	leaf := &LayoutNode{Direction: None, Command: ""}
	if got := BuildGhosttyArgs(leaf, "/tmp"); got != nil {
		t.Errorf("expected nil args for leaf with no command, got %#v", got)
	}
}

func TestBuildGhosttyArgsLeafWithCommand(t *testing.T) {
	leaf := &LayoutNode{Direction: None, Command: "vim"}
	want := []string{"--command=vim"}
	got := BuildGhosttyArgs(leaf, "/tmp")

	if len(got) != len(want) {
		t.Fatalf("unexpected arg count: got %d, want %d", len(got), len(want))
	}
	if got[0] != want[0] {
		t.Fatalf("unexpected arg: got %q, want %q", got[0], want[0])
	}
}
