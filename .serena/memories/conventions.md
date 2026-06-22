# Conventions

- Single `package main`; all files flat in repo root
- Bubble Tea MVU pattern: `Init()`, `Update(msg)`, `View()` on `wizardModel`
- Wizard steps are `wizardStep` iota enum; state machine driven by key presses and `handleEnter()`
- `pending []pendingPath` is a queue of unconfigured tree paths; BFS-style node configuration
- `nodeAt(root **LayoutNode, path []int)` traverses/creates nodes by path (0=left, 1=right)
- `SplitDirection` is `"vertical"` | `"horizontal"` | `"none"` (leaf)
- JSON tags use `snake_case` (`size_pct`, `template_name`, `root_node`, etc.)
- `filepath.Base()` sanitization applied to template names to prevent path traversal
- lipgloss styles declared as package-level `var` in `wizard.go`
- Canvas uses Unicode box-drawing chars (`┌┐└┘─│`)
