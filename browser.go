package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

// browser represents a full-screen directory selector
type browser struct {
	currentDir    string
	entries       []string // directory names only (or drive letters in drive mode)
	list          listState
	editingIndex  int    // which path entry we're editing (-1 for add mode)
	addSource     string // "user" or "system" when adding new entry (empty when editing)
	showingDrives bool   // true when showing drive selector (Windows only)
}

func newBrowser(startPath string, editingIndex int, height int) *browser {
	b := &browser{
		editingIndex: editingIndex,
		list:         listState{headerRows: 1}, // 1 header row for directory path
	}
	b.currentDir = findValidStartDir(startPath)
	b.loadEntries()
	b.list.SetViewHeight(height, len(b.entries))
	return b
}

// newBrowserForAdd creates a browser in add mode, starting at the drive root
func newBrowserForAdd(source string, height int) *browser {
	b := &browser{
		editingIndex: -1, // -1 indicates add mode
		addSource:    source,
		list:         listState{headerRows: 1},
	}
	// Start at the first available drive root
	b.currentDir = findFirstDrive()
	b.loadEntries()
	b.list.SetViewHeight(height, len(b.entries))
	return b
}

// findValidStartDir walks up the path until it finds an existing directory.
// If no valid directory is found (e.g., drive doesn't exist), returns the first available drive.
func findValidStartDir(path string) string {
	// Try the path itself
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return path
	}

	// Walk up the directory tree
	current := path
	for {
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root, check if it exists
			break
		}
		if info, err := os.Stat(parent); err == nil && info.IsDir() {
			return parent
		}
		current = parent
	}

	// Try the volume root
	volRoot := filepath.VolumeName(path) + string(filepath.Separator)
	if volRoot != string(filepath.Separator) {
		// Windows-style path with drive letter
		if info, err := os.Stat(volRoot); err == nil && info.IsDir() {
			return volRoot
		}
		// Drive doesn't exist, find first available drive
		return findFirstDrive()
	}

	// Unix-style, just use root
	return "/"
}

// findFirstDrive finds the first available drive letter on Windows
func findFirstDrive() string {
	for c := 'C'; c <= 'Z'; c++ {
		drive := string(c) + ":\\"
		if info, err := os.Stat(drive); err == nil && info.IsDir() {
			return drive
		}
	}
	// Try A and B as last resort
	for c := 'A'; c <= 'B'; c++ {
		drive := string(c) + ":\\"
		if info, err := os.Stat(drive); err == nil && info.IsDir() {
			return drive
		}
	}
	// Fallback (shouldn't happen on a working Windows system)
	return "C:\\"
}

// getAvailableDrives returns a list of available drive letters on Windows
func getAvailableDrives() []string {
	var drives []string
	// Check C-Z first (most common)
	for c := 'C'; c <= 'Z'; c++ {
		drive := string(c) + ":\\"
		if info, err := os.Stat(drive); err == nil && info.IsDir() {
			drives = append(drives, string(c)+":")
		}
	}
	// Check A and B (floppy drives, rare but possible)
	for c := 'A'; c <= 'B'; c++ {
		drive := string(c) + ":\\"
		if info, err := os.Stat(drive); err == nil && info.IsDir() {
			drives = append(drives, string(c)+":")
		}
	}
	return drives
}

// isAtDriveRoot returns true if currentDir is a Windows drive root (e.g., "C:\")
func isAtDriveRoot(path string) bool {
	clean := filepath.Clean(path)
	// Check if it's a drive root like "C:\"
	if len(clean) == 3 && clean[1] == ':' && (clean[2] == '\\' || clean[2] == '/') {
		return true
	}
	// Also handle "C:" without trailing slash
	if len(clean) == 2 && clean[1] == ':' {
		return true
	}
	return false
}

func (b *browser) loadEntries() {
	b.entries = nil
	b.list.Reset()

	if b.showingDrives {
		// Show available drives
		b.entries = getAvailableDrives()
		return
	}

	// Add parent directory option - on Windows at drive root, this will go to drive list
	if filepath.Dir(b.currentDir) != b.currentDir {
		// Not at root, normal parent navigation
		b.entries = append(b.entries, "..")
	} else if isAtDriveRoot(b.currentDir) {
		// At drive root on Windows, ".." goes to drive selector
		b.entries = append(b.entries, "..")
	}
	// On Unix at "/", no ".." entry (true root)

	entries, err := os.ReadDir(b.currentDir)
	if err != nil {
		return // Can't read directory, but ".." is still available
	}

	// Collect directories only
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}

	// Sort case-insensitively
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i]) < strings.ToLower(dirs[j])
	})

	b.entries = append(b.entries, dirs...)
}

// Update handles input for the browser
// Returns: updated browser (nil if closed), tea.Cmd, selected path (empty if cancelled)
func (b *browser) Update(msg tea.KeyMsg) (*browser, tea.Cmd, string) {
	switch msg.String() {
	case keyUp, keyUpAlt:
		b.list.MoveUp()

	case keyDown, keyDownAlt:
		b.list.MoveDown(len(b.entries))

	case keyPgUp, keyPgUpAlt:
		b.list.PageUp()

	case keyPgDown, keyPgDownAlt:
		b.list.PageDown(len(b.entries))

	case keyHome, keyHomeAlt:
		b.list.Home()

	case keyEnd, keyEndAlt:
		b.list.End(len(b.entries))

	case keyEnter:
		// Descend into selected directory or select drive
		if len(b.entries) > 0 {
			selected := b.entries[b.list.cursor]

			if b.showingDrives {
				// Drive selected - switch to that drive's root
				b.currentDir = selected + "\\"
				b.showingDrives = false
				b.loadEntries()
				return b, nil, ""
			}

			if selected == ".." {
				if isAtDriveRoot(b.currentDir) {
					// At drive root, switch to drive selector
					currentDrive := strings.ToUpper(string(b.currentDir[0])) + ":"
					b.showingDrives = true
					b.loadEntries()
					// Select the drive we came from
					for i, entry := range b.entries {
						if entry == currentDrive {
							b.list.cursor = i
							break
						}
					}
				} else {
					// Remember the directory we're leaving
					exitedDir := filepath.Base(filepath.Clean(b.currentDir))
					// Go up to parent
					b.currentDir = filepath.Dir(filepath.Clean(b.currentDir))
					b.loadEntries()
					// Find and select the directory we just exited
					for i, entry := range b.entries {
						if entry == exitedDir {
							b.list.cursor = i
							break
						}
					}
				}
				b.list.EnsureVisible()
			} else {
				b.currentDir = filepath.Join(b.currentDir, selected)
				b.loadEntries()
			}
		}

	case keySelect:
		// Select current directory (or drive root in drive mode)
		if b.showingDrives && len(b.entries) > 0 {
			// Select the highlighted drive's root
			return nil, nil, b.entries[b.list.cursor] + "\\"
		}
		return nil, nil, b.currentDir

	case keyEsc:
		// Cancel
		return nil, nil, ""

	default:
		// Jump to entry starting with pressed letter (a-z), cycling through matches
		key := msg.String()
		if len(key) == 1 && key[0] >= 'a' && key[0] <= 'z' {
			if b.showingDrives {
				// In drive mode, jump directly to the drive letter
				driveLetter := strings.ToUpper(key) + ":"
				for i, entry := range b.entries {
					if entry == driveLetter {
						b.list.cursor = i
						break
					}
				}
			} else {
				// Find all entries starting with this letter
				var matches []int
				for i, entry := range b.entries {
					if entry == ".." {
						continue
					}
					entryLower := strings.ToLower(entry)
					if len(entryLower) > 0 && entryLower[0] == key[0] {
						matches = append(matches, i)
					}
				}

				if len(matches) > 0 {
					// Find current position in matches (if any)
					nextIdx := 0
					for i, idx := range matches {
						if idx == b.list.cursor {
							// Currently on a match, move to next (wrap around)
							nextIdx = (i + 1) % len(matches)
							break
						}
					}
					b.list.cursor = matches[nextIdx]
				}
			}
			b.list.EnsureVisible()
		}
	}

	return b, nil, ""
}

// View renders the browser
func (b *browser) View(viewWidth int) string {
	var sb strings.Builder

	// Header
	var header string
	if b.showingDrives {
		header = "Select drive:"
	} else {
		header = "Select directory: " + b.currentDir
	}
	if len(header) > viewWidth-1 {
		header = header[:viewWidth-4] + "..."
	}
	sb.WriteString(ansiBold + header + ansiReset + "\n")

	start, end := b.list.VisibleRange(len(b.entries))
	scrollbar := b.list.RenderScrollbar(len(b.entries))

	// Render directory entries
	for i := start; i < end; i++ {
		entry := b.entries[i]

		var prefix string
		if i == b.list.cursor {
			prefix = ">"
		} else {
			prefix = " "
		}

		// Leave room for scrollbar
		line := prefix + " " + entry
		maxLen := viewWidth - 3 // prefix + space + scrollbar
		lineLen := utf8.RuneCountInString(line)
		if lineLen > maxLen {
			// Truncate by runes, not bytes
			runes := []rune(line)
			line = string(runes[:maxLen-3]) + "..."
		}

		// Pad to align scrollbar (use rune count)
		lineLen = utf8.RuneCountInString(line)
		if lineLen < maxLen {
			line += strings.Repeat(" ", maxLen-lineLen)
		}

		scrollIdx := i - start
		sb.WriteString(line + " " + scrollbar[scrollIdx] + "\n")
	}

	// Pad remaining lines
	totalHeight := b.list.TotalHeight()
	rendered := end - start + 1 // +1 for header
	for i := rendered; i < totalHeight; i++ {
		scrollIdx := i - 1 // -1 for header
		scrollChar := " "
		if scrollIdx >= 0 && scrollIdx < len(scrollbar) {
			scrollChar = scrollbar[scrollIdx]
		}
		sb.WriteString(strings.Repeat(" ", viewWidth-1) + scrollChar + "\n")
	}

	return sb.String()
}
