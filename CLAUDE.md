# Claude Instructions for pathed-go

## Project Overview
A spartan TUI PATH editor for Windows, built with Go and Bubble Tea.

## Building
**Preferred**: Ask the user to build with `Ctrl+Shift+B` (VS Code build task).

Output goes to `bin/pathed.exe` (gitignored).

Manual command if needed:
```bash
go build -ldflags "-s -w" -o bin/pathed.exe .
```
Always use `-ldflags "-s -w"` for release builds to strip debug info.

## Code Style
- Keep the UI minimal/spartan - no unnecessary styling
- Use ANSI escape codes directly for colors (e.g., `\x1b[32m` for green)
- Avoid Unicode box-drawing characters - they don't render reliably on Windows terminals

## Architecture
Source files at root (standard for small Go CLIs):
- `main.go` - Entry point
- `model.go` - Application state (`model` struct)
- `update.go` - Input handling and key bindings
- `view.go` - Rendering logic
- `prompt.go` - Dialog component with own Update/View
- `entry.go` - `pathEntry` struct
- `ansi.go` - ANSI escape code constants
- `paths_windows.go` / `paths_other.go` - Platform-specific PATH loading

Key concepts:
- Scrolling: vertical (`offset`) and horizontal (`hOffset`)
- Scrollbar uses ANSI colors: white background for thumb, dark grey for track

## Key Bindings
- `j`/`k` or arrows: vertical navigation
- `J`/`K` or `Shift+Up/Down`: move entry up/down within section
- `left`/`right`: horizontal scroll
- `g`/`G` or `Home`/`End`: jump to first/last
- `PgUp`/`PgDn` or `Ctrl+U`/`Ctrl+D`: page up/down
- `q`: quit (asks to save if modified)
- `Ctrl+C`: force quit
- In prompts: `y`/`n` shortcuts, `Esc` to cancel

## Testing
Run the built executable in Windows Terminal for proper ANSI support.

## Safety
Any code that is written to actually modify the path should be commented out for safaty. So the code should be written, but not actually be compiled and run for now.
