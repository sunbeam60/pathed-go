package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// version is set via ldflags at build time
var version = "dev"

const helpText = `pathed - Interactive PATH environment editor

USAGE:
    pathed [OPTIONS]

OPTIONS:
    -h, --help        Show this help message
    -v, --version     Show version
    -r, --registry    Read from and write to Windows registry (Windows only)

DESCRIPTION:
    pathed provides a TUI for editing your PATH environment variable.

    Default mode (environment):
      Reads PATH from the process environment and outputs a PATH string
      for your shell to capture. Use this for session-based PATH editing.

    Registry mode (--registry, Windows only):
      Reads from and writes to the Windows registry. Shows system and user
      PATH entries separately. Changes are persisted directly to the registry.
      No output is produced (shell capture not needed).

USAGE EXAMPLES:
  Linux/macOS (bash/zsh):
    export PATH="$(pathed)"

  Nushell:
    $env.PATH = (pathed | str trim | split row ';')

  Windows PowerShell:
    $env:PATH = (pathed)

  Windows cmd.exe:
    for /f "delims=" %i in ('pathed') do set "PATH=%i"
  
  Or add a shell function to your profile:
    # PowerShell ($PROFILE)
    function Edit-Path { $env:PATH = (pathed) }

    # Bash/Zsh (.bashrc or .zshrc)
    pathed() { export PATH="$(command pathed "$@")"; }

KEY BINDINGS:
    j/k, Up/Down     Navigate
    J/K, Shift+Up/Dn Move entry up/down (within section in registry mode)
    Left/Right       Horizontal scroll
    g/G, Home/End    Jump to first/last
    Tab              Edit path (opens directory browser)
    a                Add PATH entry (user entry in registry mode)
    A                Add system PATH entry (registry mode only)
    c                Clean (mark duplicates & missing for deletion)
    Del              Toggle delete mark
    q                Quit (prompts if changes exist)
    Ctrl+C           Force quit

QUIT OPTIONS:
    Default mode:     "Edited" (output modified) / "Original" (output unchanged)
    Registry mode:    "Persist" (save to registry) / "Don't persist" (discard)
`

func main() {
	// Parse command-line flags
	registryMode := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			fmt.Print(helpText)
			return
		case "-v", "--version":
			fmt.Println(version)
			return
		case "-r", "--registry":
			if !supportsRegistry {
				fmt.Fprintln(os.Stderr, "Error: --registry flag is only supported on Windows")
				os.Exit(1)
			}
			registryMode = true
		default:
			fmt.Fprintf(os.Stderr, "Unknown option: %s\nUse --help for usage information.\n", arg)
			os.Exit(1)
		}
	}

	// Open terminal device directly for TUI output, keeping stdout clean for piping
	tty, err := openTTY()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening terminal: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	p := tea.NewProgram(initialModel(registryMode), tea.WithAltScreen(), tea.WithOutput(tty))
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Handle output based on mode
	if m, ok := finalModel.(model); ok {
		if m.registryMode {
			// Registry mode: persist to registry if user chose to save
			if m.saveChanges {
				if err := savePaths(m.paths); err != nil {
					fmt.Fprintf(os.Stderr, "Error saving to registry: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			// Env mode: always output a PATH string
			if m.saveChanges {
				fmt.Println(buildPathString(m.paths))
			} else {
				fmt.Println(m.originalPath)
			}
		}
	}
}
