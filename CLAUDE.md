# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build -o hsp .          # Build binary
go test ./...              # Run all tests
go test -run TestName .    # Run a single test
go vet ./...               # Static analysis
```

The binary entry point is `hsp`. Two subcommands: `boo` (interactive TUI) and `summon <template>`.

## Architecture

`haunt-space` is a single-package Go CLI (`package main`) that declaratively manages Ghostty terminal window layouts.

**Data model (`models.go`):** A `GlobalBlueprint` wraps a binary tree of `LayoutNode`s. Each node is either a split (`Vertical`/`Horizontal` with a `Size` percentage) or a leaf (`None` with an optional `Command`). Templates are serialized to JSON at `~/.config/haunt-space/templates/<name>.json`.

**Compiler (`compiler.go`):** `BuildGhosttyCommand` does a pre-order traversal of the `LayoutNode` tree, emitting `split-surface --vertical/--horizontal` and `--command=` flags that are passed to `ghostty` via `sh -c`.

**Launcher (`launcher.go`):** Reads a JSON template, calls the compiler, then spawns Ghostty as a detached background process (`cmd.Process.Release()`).

**Wizard (`wizard.go`):** A Bubble Tea TUI that builds the layout tree interactively. Uses a BFS queue (`pending []pendingPath`) of tree paths still needing configuration. When a split node is chosen, its two child paths are pushed onto the queue. When all nodes are configured, `finalize()` serializes and saves the blueprint.

**Canvas (`canvas.go`):** A 2-D rune grid that renders ASCII box previews of the in-progress layout tree during the wizard. `RenderNodePreview` mirrors the compiler's recursive structure to draw proportional boxes.
