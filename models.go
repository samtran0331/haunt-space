package main

// SplitDirection defines how a layout node splits its surface.
type SplitDirection string

const (
	Vertical   SplitDirection = "vertical"
	Horizontal SplitDirection = "horizontal"
	None       SplitDirection = "none" // Represents a leaf node (an actual terminal pane)
)

// LayoutNode is a single node in the binary window topology tree.
type LayoutNode struct {
	Direction  SplitDirection `json:"direction"`             // "vertical", "horizontal", or "none"
	Size       int            `json:"size_pct"`              // Split percentage ratio (e.g., 30, 50, 70)
	Folder     string         `json:"folder,omitempty"`      // Working directory for this pane; "" = cwd, "~" = home
	Command    string         `json:"command,omitempty"`     // Executable to run; empty = default shell
	LeftChild  *LayoutNode    `json:"left_child,omitempty"`  // Left or Top pane branch
	RightChild *LayoutNode    `json:"right_child,omitempty"` // Right or Bottom pane branch
}

// GlobalBlueprint is the top-level structure stored as a JSON template file.
type GlobalBlueprint struct {
	TemplateName string     `json:"template_name"`
	Root         LayoutNode `json:"root_node"`
}
