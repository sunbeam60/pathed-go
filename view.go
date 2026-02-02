package main

import "strings"

// renderEntryPrefix returns the 2-character prefix for a path entry (state marker + cursor/exists marker)
func renderEntryPrefix(entry pathEntry, isCursor bool) string {
	// First char: modification state (priority: deleted > added > modified)
	var prefix string
	if entry.deleted {
		prefix = ansiRed + "-" + ansiReset
	} else if entry.added {
		prefix = ansiGreen + "+" + ansiReset
	} else if entry.modified {
		prefix = ansiRed + "*" + ansiReset
	} else {
		prefix = " "
	}

	// Second char: cursor or exists indicator
	if isCursor {
		if !entry.exists {
			prefix += ansiBlue + ">" + ansiReset
		} else {
			prefix += ">"
		}
	} else if !entry.exists {
		prefix += ansiBlue + "?" + ansiReset
	} else {
		prefix += " "
	}
	return prefix
}

// renderEntryStyle returns ANSI style codes for an entry based on its state
func renderEntryStyle(entry pathEntry) (style string, needsReset bool) {
	if entry.deleted && entry.source == "system" {
		return ansiRed + ansiBgRed, true
	} else if entry.deleted {
		return ansiRed, true
	} else if entry.added && entry.source == "system" {
		return ansiGreen + ansiBgGreen, true
	} else if entry.added {
		return ansiGreen, true
	} else if entry.source == "system" {
		return ansiBold + ansiBgGrey, true
	}
	return "", false
}

// renderHelpBar returns the help bar text for the main view
func renderHelpBar(registryMode bool, width int) string {
	var addHelp string
	if registryMode {
		addHelp = "a/A: add user/system"
	} else {
		addHelp = "a: add"
	}
	helpBar := " Tab: edit | " + addHelp + " | c: clean | Del: delete | q: quit"
	if len(helpBar) > width {
		helpBar = helpBar[:width-3] + "..."
	}
	return helpBar
}

func (m model) View() string {
	var b strings.Builder

	// If browser is active, render it instead of the path list
	if m.browser != nil {
		b.WriteString(m.browser.View(m.viewWidth))
		helpBar := " Enter: open | a-z/A-Z: jump fwd/back | Tab: select | Esc: cancel"
		if len(helpBar) > m.viewWidth {
			helpBar = helpBar[:m.viewWidth-3] + "..."
		}
		b.WriteString(helpBar)
		return b.String()
	}

	// Calculate visible range and scrollbar
	start, end := m.list.VisibleRange(len(m.paths))
	scrollbar := m.list.RenderScrollbar(len(m.paths))

	// Render visible paths with scrollbar
	for i := start; i < end; i++ {
		entry := m.paths[i]
		prefix := renderEntryPrefix(entry, i == m.list.cursor)
		path := entry.path
		pathRunes := []rune(path)
		pathLen := len(pathRunes)
		// Available width for path content: total - cursor(2) - scrollbar(2) - possible markers(2)
		contentWidth := m.viewWidth - 4 // cursor + scrollbar + space

		// Apply horizontal offset (in runes, not bytes)
		var visibleRunes []rune
		if m.list.hOffset > 0 {
			if m.list.hOffset < pathLen {
				visibleRunes = pathRunes[m.list.hOffset:]
			} else {
				visibleRunes = nil
			}
		} else {
			visibleRunes = pathRunes
		}

		// Determine if we need left/right markers
		hasLeft := m.list.hOffset > 0 && pathLen > 0

		// Adjust content width for markers
		displayWidth := contentWidth
		if hasLeft {
			displayWidth--
		}
		hasRight := len(visibleRunes) > displayWidth
		if hasRight {
			displayWidth--
		}

		// Truncate to display width (in runes)
		if len(visibleRunes) > displayWidth {
			visibleRunes = visibleRunes[:displayWidth]
		}
		visiblePath := string(visibleRunes)

		// Build line with markers (green colored)
		var line strings.Builder
		line.WriteString(prefix)
		if hasLeft {
			line.WriteString(ansiGreen + "<" + ansiReset)
		}
		style, needsReset := renderEntryStyle(entry)
		line.WriteString(style)
		line.WriteString(visiblePath)
		if needsReset {
			line.WriteString(ansiReset)
		}

		// Pad to align right marker and scrollbar
		currentLen := 2 + len(visibleRunes) // cursor + content (rune count)
		if hasLeft {
			currentLen++
		}
		targetLen := m.viewWidth - 2 // leave room for scrollbar
		if hasRight {
			targetLen-- // leave room for > marker
		}
		if padding := targetLen - currentLen; padding > 0 {
			line.WriteString(strings.Repeat(" ", padding))
		}
		if hasRight {
			line.WriteString(ansiGreen + ">" + ansiReset)
		}

		// Add scrollbar character
		scrollIdx := i - start
		if scrollIdx < len(scrollbar) {
			line.WriteString(" " + scrollbar[scrollIdx])
		}

		b.WriteString(line.String() + "\n")
	}

	// Pad remaining lines if list is shorter than viewport
	for i := end - start; i < m.list.viewHeight; i++ {
		scrollIdx := i
		scrollChar := " "
		if scrollIdx < len(scrollbar) {
			scrollChar = scrollbar[scrollIdx]
		}
		b.WriteString(strings.Repeat(" ", m.viewWidth-1) + scrollChar + "\n")
	}

	// Warning if not elevated in registry mode
	if m.registryMode && !m.elevated {
		warning := ansiYellow + " Warning: Not running as Administrator - system PATH changes will fail" + ansiReset
		b.WriteString(warning + "\n")
	}

	// Help bar or prompt
	if m.prompt != nil {
		b.WriteString(m.prompt.View())
	} else {
		b.WriteString(renderHelpBar(m.registryMode, m.viewWidth))
	}

	return b.String()
}
