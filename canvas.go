package main

import (
	"fmt"
	"strings"
)

// Canvas is a 2-D rune grid used for ASCII layout previews.
type Canvas struct {
	Width  int
	Height int
	Grid   [][]rune
}

// NewCanvas allocates a blank canvas filled with spaces.
func NewCanvas(w, h int) Canvas {
	grid := make([][]rune, h)
	for i := range grid {
		grid[i] = make([]rune, w)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}
	return Canvas{Width: w, Height: h, Grid: grid}
}

// DrawBox draws a Unicode box around the region (x, y, w, h).
func (c *Canvas) DrawBox(x, y, w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	c.Grid[y][x], c.Grid[y][x+w-1] = '┌', '┐'
	c.Grid[y+h-1][x], c.Grid[y+h-1][x+w-1] = '└', '┘'
	for i := x + 1; i < x+w-1; i++ {
		c.Grid[y][i], c.Grid[y+h-1][i] = '─', '─'
	}
	for i := y + 1; i < y+h-1; i++ {
		c.Grid[i][x], c.Grid[i][x+w-1] = '│', '│'
	}
}

// WriteText centers text inside the given region of the canvas.
func (c *Canvas) WriteText(x, y, w, h int, text string) {
	if text == "" || w <= 2 || h <= 2 {
		return
	}

	r := []rune(text)
	if len(r) > w-2 {
		maxLen := w - 3
		if maxLen < 0 {
			return
		}
		r = append(r[:maxLen], '…')
	}

	targetY := y + (h / 2)
	targetX := x + ((w - len(r)) / 2)
	for i, ch := range r {
		xi := targetX + i
		if xi >= 0 && xi < c.Width && targetY >= 0 && targetY < c.Height {
			c.Grid[targetY][xi] = ch
		}
	}
}

// Render converts the canvas grid to a printable multi-line string.
func (c *Canvas) Render() string {
	rows := make([]string, c.Height)
	for i, row := range c.Grid {
		rows[i] = string(row)
	}
	return strings.Join(rows, "\n")
}

// pendingLabel returns the display label for a node region, checking whether
// the given path is a pending pane before falling back to "[ empty ]".
func pendingLabel(path []int, pending []pendingPath, selectedPath []int) string {
	for _, p := range pending {
		if pathsEqual(p.path, path) {
			if pathsEqual(path, selectedPath) {
				return fmt.Sprintf("►%d◄", p.paneNum)
			}
			return fmt.Sprintf("[%d]", p.paneNum)
		}
	}
	return "[ empty ]"
}

// pathsEqual reports whether two tree paths are identical.
func pathsEqual(a, b []int) bool {
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// RenderNodePreview recursively draws the layout tree onto the canvas.
// nodePath is the path to the current node (built up during recursion).
// pending is the list of unconfigured panes; each is labelled with its number.
// selectedPath is the currently active pane (highlighted with ►◄).
func RenderNodePreview(c *Canvas, node *LayoutNode, x, y, w, h int, nodePath []int, pending []pendingPath, selectedPath []int) {
	if node == nil || w <= 1 || h <= 1 {
		return
	}

	if node.Direction == None {
		c.DrawBox(x, y, w, h)
		label := ""
		isPending := false
		for _, p := range pending {
			if pathsEqual(p.path, nodePath) {
				isPending = true
				if pathsEqual(nodePath, selectedPath) {
					label = fmt.Sprintf("►%d◄", p.paneNum)
				} else {
					label = fmt.Sprintf("[%d]", p.paneNum)
				}
				break
			}
		}
		if !isPending {
			label = node.Command
			if label == "" {
				label = "zsh"
			}
		}
		c.WriteText(x, y, w, h, label)
		return
	}

	pct := node.Size
	if pct == 0 {
		pct = 50
	}

	leftPath := append(copyPath(nodePath), 0)
	rightPath := append(copyPath(nodePath), 1)

	if node.Direction == Vertical {
		splitW := (w * pct) / 100
		if splitW <= 1 {
			splitW = 2
		}
		if node.LeftChild == nil {
			c.DrawBox(x, y, splitW, h)
			c.WriteText(x, y, splitW, h, pendingLabel(leftPath, pending, selectedPath))
		} else {
			RenderNodePreview(c, node.LeftChild, x, y, splitW, h, leftPath, pending, selectedPath)
		}
		if node.RightChild == nil {
			c.DrawBox(x+splitW-1, y, w-splitW+1, h)
			c.WriteText(x+splitW-1, y, w-splitW+1, h, pendingLabel(rightPath, pending, selectedPath))
		} else {
			RenderNodePreview(c, node.RightChild, x+splitW-1, y, w-splitW+1, h, rightPath, pending, selectedPath)
		}
	} else if node.Direction == Horizontal {
		splitH := (h * pct) / 100
		if splitH <= 1 {
			splitH = 2
		}
		if node.LeftChild == nil {
			c.DrawBox(x, y, w, splitH)
			c.WriteText(x, y, w, splitH, pendingLabel(leftPath, pending, selectedPath))
		} else {
			RenderNodePreview(c, node.LeftChild, x, y, w, splitH, leftPath, pending, selectedPath)
		}
		if node.RightChild == nil {
			c.DrawBox(x, y+splitH-1, w, h-splitH+1)
			c.WriteText(x, y+splitH-1, w, h-splitH+1, pendingLabel(rightPath, pending, selectedPath))
		} else {
			RenderNodePreview(c, node.RightChild, x, y+splitH-1, w, h-splitH+1, rightPath, pending, selectedPath)
		}
	}
}
