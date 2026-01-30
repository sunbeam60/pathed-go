package main

import "strings"

func (m model) View() string {
	var b strings.Builder

	// Calculate visible range
	end := m.offset + m.viewHeight
	if end > len(m.paths) {
		end = len(m.paths)
	}

	// Calculate scrollbar
	scrollbar := m.renderScrollbar()

	// Render visible paths with scrollbar
	for i := m.offset; i < end; i++ {
		entry := m.paths[i]

		// Build cursor/marker prefix: [modified marker][cursor]
		var prefix string
		if entry.modified {
			prefix = ansiRed + "*" + ansiReset // red asterisk
		} else {
			prefix = " "
		}
		if i == m.cursor {
			prefix += ">"
		} else {
			prefix += " "
		}
		path := entry.path
		// Available width for path content: total - cursor(2) - scrollbar(2) - possible markers(2)
		contentWidth := m.viewWidth - 4 // cursor + scrollbar + space

		// Apply horizontal offset
		visiblePath := path
		if m.hOffset > 0 {
			if m.hOffset < len(path) {
				visiblePath = path[m.hOffset:]
			} else {
				visiblePath = ""
			}
		}

		// Determine if we need left/right markers
		hasLeft := m.hOffset > 0 && len(path) > 0

		// Adjust content width for markers
		displayWidth := contentWidth
		if hasLeft {
			displayWidth--
		}
		hasRight := len(visiblePath) > displayWidth
		if hasRight {
			displayWidth--
		}

		// Truncate to display width
		if len(visiblePath) > displayWidth {
			visiblePath = visiblePath[:displayWidth]
		}

		// Build line with markers (green colored)
		var line strings.Builder
		line.WriteString(prefix)
		if hasLeft {
			line.WriteString(ansiGreen + "<" + ansiReset)
		}
		// System paths in bold white with dark grey background, user paths in regular white
		if entry.source == "system" {
			line.WriteString(ansiBold + ansiBgGrey) // bold + dark grey background
		}
		line.WriteString(visiblePath)
		if entry.source == "system" {
			line.WriteString(ansiReset) // reset
		}

		// Pad to align right marker and scrollbar
		currentLen := 2 + len(visiblePath) // cursor + content
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
		scrollIdx := i - m.offset
		if scrollIdx < len(scrollbar) {
			line.WriteString(" " + scrollbar[scrollIdx])
		}

		b.WriteString(line.String() + "\n")
	}

	// Pad remaining lines if list is shorter than viewport
	for i := end - m.offset; i < m.viewHeight; i++ {
		scrollIdx := i
		scrollChar := " "
		if scrollIdx < len(scrollbar) {
			scrollChar = scrollbar[scrollIdx]
		}
		b.WriteString(strings.Repeat(" ", m.viewWidth-1) + scrollChar + "\n")
	}

	// Help bar or modal
	if m.modal != nil {
		b.WriteString(m.modal.View())
	} else {
		b.WriteString(" q: quit")
	}

	return b.String()
}

func (m model) renderScrollbar() []string {
	result := make([]string, m.viewHeight)

	if m.viewHeight >= len(m.paths) {
		// No scrollbar needed
		for i := range result {
			result[i] = " "
		}
		return result
	}

	// Calculate thumb size (minimum 1)
	thumbSize := m.viewHeight * m.viewHeight / len(m.paths)
	if thumbSize < 1 {
		thumbSize = 1
	}

	// Calculate thumb position
	scrollRange := len(m.paths) - m.viewHeight
	thumbRange := m.viewHeight - thumbSize
	thumbPos := 0
	if scrollRange > 0 {
		thumbPos = m.offset * thumbRange / scrollRange
	}

	// Build scrollbar with ANSI colors
	for i := 0; i < m.viewHeight; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			// Bright white background for thumb
			result[i] = ansiBgWhite + " " + ansiReset
		} else {
			// Dim grey background for track
			result[i] = ansiBgGrey + " " + ansiReset
		}
	}
	return result
}
