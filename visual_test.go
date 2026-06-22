package main

import "testing"

func TestCanvasBoxSizing(t *testing.T) {
	canvas := NewCanvas(10, 5)
	canvas.DrawBox(0, 0, 10, 5)

	if canvas.Grid[0][0] != '┌' || canvas.Grid[0][9] != '┐' || canvas.Grid[4][0] != '└' || canvas.Grid[4][9] != '┘' {
		t.Errorf("Canvas rendering pipeline generated faulty layout border coordinates.")
	}
}

func TestCanvasWriteText(t *testing.T) {
	canvas := NewCanvas(20, 10)
	canvas.DrawBox(0, 0, 20, 10)
	canvas.WriteText(0, 0, 20, 10, "hello")

	// The text should appear somewhere on the canvas; just verify it's non-space.
	found := false
	for _, row := range canvas.Grid {
		for _, ch := range row {
			if ch == 'h' {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("WriteText did not place text on the canvas")
	}
}

func TestCanvasWriteTextTruncation(t *testing.T) {
	canvas := NewCanvas(10, 5)
	canvas.DrawBox(0, 0, 10, 5)
	// Text longer than the box interior should be truncated without panic.
	canvas.WriteText(0, 0, 10, 5, "this is a very long string that should be truncated")
}

func TestNewCanvas(t *testing.T) {
	c := NewCanvas(5, 3)
	if c.Width != 5 || c.Height != 3 {
		t.Errorf("NewCanvas dimensions wrong: got %dx%d", c.Width, c.Height)
	}
	for y, row := range c.Grid {
		for x, ch := range row {
			if ch != ' ' {
				t.Errorf("expected space at (%d,%d), got %q", x, y, ch)
			}
		}
	}
}

func TestRenderNodePreviewLeaf(t *testing.T) {
	c := NewCanvas(20, 10)
	node := &LayoutNode{Direction: None, Command: "vim"}
	// Should not panic and should draw a box.
	RenderNodePreview(&c, node, 0, 0, 20, 10, 0, []int{})
	if c.Grid[0][0] != '┌' {
		t.Errorf("expected box corner at (0,0)")
	}
}

func TestRenderNodePreviewSplit(t *testing.T) {
	c := NewCanvas(40, 20)
	node := &LayoutNode{
		Direction: Vertical,
		Size:      50,
		LeftChild:  &LayoutNode{Direction: None, Command: "vim"},
		RightChild: &LayoutNode{Direction: None, Command: "zsh"},
	}
	// Should not panic.
	RenderNodePreview(&c, node, 0, 0, 40, 20, 0, []int{0})
}

func TestCanvasRender(t *testing.T) {
	c := NewCanvas(5, 2)
	out := c.Render()
	lines := splitLines(out)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

// splitLines splits a string by newlines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, ch := range s {
		if ch == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}
