package main

import "strings"

func (m model) View() string {
	var b strings.Builder

	// If browser is active, render it instead of the path list
	if m.browser != nil {
		b.WriteString(m.browser.View(m.viewWidth))
		helpBar := " Enter: open | a-z: jump | Tab: select | Esc: cancel"
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

		// Build cursor/marker prefix: [modified marker][cursor/exists marker]
		// Priority: deleted > added > modified
		var prefix string
		if entry.deleted {
			prefix = ansiRed + "-" + ansiReset // red - for deleted entries
		} else if entry.added {
			prefix = ansiGreen + "+" + ansiReset // green + for added entries
		} else if entry.modified {
			prefix = ansiRed + "*" + ansiReset // red asterisk for modified
		} else {
			prefix = " "
		}
		if i == m.list.cursor {
			if !entry.exists {
				prefix += ansiBlue + ">" + ansiReset // blue cursor for non-existent
			} else {
				prefix += ">"
			}
		} else if !entry.exists {
			prefix += ansiBlue + "?" + ansiReset // blue ? for non-existent paths
		} else {
			prefix += " "
		}
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
		// Style based on state: deleted > added > normal
		needsReset := false
		if entry.deleted && entry.source == "system" {
			line.WriteString(ansiRed + ansiBgRed) // red text, light red background
			needsReset = true
		} else if entry.deleted {
			line.WriteString(ansiRed) // just red text for user entries
			needsReset = true
		} else if entry.added && entry.source == "system" {
			line.WriteString(ansiGreen + ansiBgGreen) // green text, light green background
			needsReset = true
		} else if entry.added {
			line.WriteString(ansiGreen) // just green text for user entries
			needsReset = true
		} else if entry.source == "system" {
			line.WriteString(ansiBold + ansiBgGrey) // bold + dark grey background
			needsReset = true
		}
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

	// Help bar or prompt
	if m.prompt != nil {
		b.WriteString(m.prompt.View())
	} else {
		helpBar := " Tab: edit | " + addHelpText + " | c: clean | Del: delete | q: quit"
		if len(helpBar) > m.viewWidth {
			helpBar = helpBar[:m.viewWidth-3] + "..."
		}
		b.WriteString(helpBar)
	}

	return b.String()
}
