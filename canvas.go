package main

import "strings"

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

// RenderNodePreview recursively draws the layout tree onto the canvas.
// pathIndex tracks the recursion depth; currentPath is the active wizard focus.
func RenderNodePreview(c *Canvas, node *LayoutNode, x, y, w, h int, pathIndex int, currentPath []int) {
	if node == nil || w <= 1 || h <= 1 {
		return
	}

	// Determine if this specific branch node matches the active targeted wizard path
	isActive := true
	if len(currentPath) < pathIndex {
		isActive = false
	}

	if node.Direction == None {
		c.DrawBox(x, y, w, h)
		name := node.Command
		if name == "" {
			name = "zsh prompt"
		}
		if isActive {
			name = "👉 [ " + name + " ]"
		}
		c.WriteText(x, y, w, h, name)
		return
	}

	pct := node.Size
	if pct == 0 {
		pct = 50
	}

	if node.Direction == Vertical {
		splitW := (w * pct) / 100
		if splitW <= 1 {
			splitW = 2
		}
		if node.LeftChild == nil {
			c.DrawBox(x, y, splitW, h)
			c.WriteText(x, y, splitW, h, "[ empty ]")
		} else {
			RenderNodePreview(c, node.LeftChild, x, y, splitW, h, pathIndex+1, currentPath)
		}
		if node.RightChild == nil {
			c.DrawBox(x+splitW-1, y, w-splitW+1, h)
			c.WriteText(x+splitW-1, y, w-splitW+1, h, "[ empty ]")
		} else {
			RenderNodePreview(c, node.RightChild, x+splitW-1, y, w-splitW+1, h, pathIndex+1, currentPath)
		}
	} else if node.Direction == Horizontal {
		splitH := (h * pct) / 100
		if splitH <= 1 {
			splitH = 2
		}
		if node.LeftChild == nil {
			c.DrawBox(x, y, w, splitH)
			c.WriteText(x, y, w, splitH, "[ empty ]")
		} else {
			RenderNodePreview(c, node.LeftChild, x, y, w, splitH, pathIndex+1, currentPath)
		}
		if node.RightChild == nil {
			c.DrawBox(x, y+splitH-1, w, h-splitH+1)
			c.WriteText(x, y+splitH-1, w, h-splitH+1, "[ empty ]")
		} else {
			RenderNodePreview(c, node.RightChild, x, y+splitH-1, w, h-splitH+1, pathIndex+1, currentPath)
		}
	}
}
