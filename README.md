# pathed

A spartan TUI for editing your PATH environment variable.

![pathed screenshot](https://github.com/user-attachments/assets/43815fb6-355c-4069-84f2-560dd49a6881)

⚠️ On Windows, this tool can make permanent and undoable changes to your registry. 

⚠️ While it's been tested and seems to work, you could screw up your computer when running with -r/--registry.

## Features

- **Two modes of operation:**
  - **Environment mode** (default): Reads PATH from the process environment, outputs modified PATH for shell capture
  - **Registry mode** (`-r`): Reads/writes directly to Windows registry with separate system and user PATH sections

- **Visual indicators:**
  - Modified entries marked with `*`
  - Added entries marked with `+`
  - Deleted entries marked with `-`
  - Non-existent paths marked with `?`
  - System PATH entries shown with distinct background (registry mode)

- **Directory browser** for editing and adding paths with keyboard navigation

- **Clean command** to mark duplicates and non-existent paths for deletion

- **Windows integration:**
  - Broadcasts `WM_SETTINGCHANGE` after registry writes so Explorer picks up changes immediately
  - Elevation detection with warning when running without Administrator privileges

## Installation

Download the latest release from the [releases page](https://github.com/sunbeam60/pathed-go/releases).

Or use [eget](https://github.com/zyedidia/eget) and do 
```bash
eget sunbeam60/pathed-go
```

### macOS Note

macOS binaries are unsigned. On first run, you may see a Gatekeeper warning. To bypass:
- Right-click the binary and select "Open", then click "Open" in the dialog
- Or run: `xattr -d com.apple.quarantine pathed`

## Usage

### Environment Mode (Linux/macOS/Windows)

Reads PATH from environment, outputs result for shell capture:

```bash
# Bash/Zsh
export PATH="$(pathed)"

# PowerShell
$env:PATH = (pathed)

# Nushell
$env.PATH = (pathed | str trim | split row ':')
```

Add to your shell profile for convenience:

```bash
# Bash/Zsh (.bashrc or .zshrc)
pathed() { export PATH="$(command pathed "$@")"; }

# PowerShell ($PROFILE)
function Edit-Path { $env:PATH = (pathed) }
```

### Registry Mode (Windows only)

Reads from and writes directly to the Windows registry:

```
pathed -r
```

- Shows system PATH (from HKLM) and user PATH (from HKCU) separately
- Changes are persisted directly to the registry
- Run as Administrator (sudo pathed -r) to persist changes to system path.

## Key Bindings

| Key | Action |
|-----|--------|
| `j`/`k`, `↑`/`↓` | Navigate up/down |
| `J`/`K`, `Shift+↑`/`↓` | Move entry up/down |
| `←`/`→` | Horizontal scroll |
| `g`/`G`, `Home`/`End` | Jump to first/last |
| `PgUp`/`PgDn`, `Ctrl+U`/`D` | Page up/down |
| `Tab` | Edit path (opens directory browser) |
| `a` | Add PATH entry (user entry in registry mode) |
| `A` | Add system PATH entry (registry mode only) |
| `c` | Clean (mark duplicates & missing for deletion) |
| `Del` | Toggle delete mark |
| `?` or `h` | Show help |
| `q` | Quit (prompts if changes exist) |
| `Ctrl+C` | Force quit |

### Directory Browser

| Key | Action |
|-----|--------|
| `Enter` | Open directory |
| `a-z` | Jump to next entry starting with letter |
| `A-Z` | Jump to previous entry starting with letter |
| `Tab` | Select current directory |
| `Esc` | Cancel |

## Building from Source

Requires Go 1.24+:

```bash
go build -ldflags "-s -w" -o pathed .
```

## License

MIT - See [LICENSE](LICENSE) for details.
