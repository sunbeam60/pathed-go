# Claude Instructions for pathed-go

## Project Overview
A spartan TUI PATH editor for Windows, built with Go and Bubble Tea.

## Working together
If the user is asking questions about a problem or approach, aim to first discuss the problem/approach with the user before you start making changes. If you now understand the intent from the user, please confirm that they are ready to make changes.
Once an approach has been chosen, it's ok to make changes necessary for that approach, but when a problem domain changes, the user prefers a discussion about approach before you start making changes again.

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
- Any error messages we return to the user should use good grammer; start with capital, finish with a full stop etc.

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
Any code that is written to actually modify the path should be commented out for safety. So the code should be written, but not actually be compiled and run for now.
