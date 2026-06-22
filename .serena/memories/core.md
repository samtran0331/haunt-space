# haunt-space — Core

CLI tool `hsp` — Ghostty terminal emulator layout engine. Builds binary split trees, saves as JSON templates, launches them in Ghostty.

## Commands
- `hsp wizard` — interactive Bubble Tea TUI to build and save layout templates
- `hsp launch <template>` — load saved template, compile to Ghostty CLI args, spawn detached

## Source map (all `package main`, flat structure)
- `main.go` — CLI entry, arg routing
- `models.go` — core types: `LayoutNode`, `GlobalBlueprint`, `SplitDirection`
- `compiler.go` — `BuildGhosttyCommand`: recursive pre-order tree → Ghostty split-surface args
- `launcher.go` — `saveBlueprint`/`loadBlueprint` (JSON), `launchTemplate` (detached exec)
- `wizard.go` — Bubble Tea model `wizardModel`, full wizard step machine
- `canvas.go` — `Canvas` 2D rune grid, `RenderNodePreview` for ASCII layout preview

## Template storage
`~/.config/haunt-space/templates/<name>.json` — created by `saveBlueprint`, read by `loadBlueprint`

## Key invariants
- Single `package main`; no sub-packages
- Ghostty invoked via `sh -c` so embedded quotes in the compiled arg string parse correctly
- Compiler builds Ghostty args via recursive `split-surface --<direction>` calls; leaf nodes emit `--command=<cmd>`

Tech stack: `mem:tech_stack`
Commands to run: `mem:suggested_commands`
Task completion: `mem:task_completion`
Conventions: `mem:conventions`
