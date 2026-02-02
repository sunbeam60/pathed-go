package main

import (
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

// helpView displays scrollable help text
type helpView struct {
	lines []string
	list  listState
}

// newHelpView creates a help view from the helpText constant
func newHelpView(height int) *helpView {
	lines := strings.Split(helpText, "\n")
	h := &helpView{
		lines: lines,
		list:  listState{},
	}
	h.list.SetViewHeight(height, len(lines))
	return h
}

// Update handles input for the help view
// Returns nil to close the help view
func (h *helpView) Update(msg tea.KeyMsg) *helpView {
	switch msg.String() {
	case keyUp, keyUpAlt:
		h.list.ScrollUp()
	case keyDown, keyDownAlt:
		h.list.ScrollDown(len(h.lines))
	case keyPgUp, keyPgUpAlt:
		h.list.ScrollPageUp()
	case keyPgDown, keyPgDownAlt:
		h.list.ScrollPageDown(len(h.lines))
	case keyHome, keyHomeAlt:
		h.list.ScrollHome()
	case keyEnd, keyEndAlt:
		h.list.ScrollEnd(len(h.lines))
	case keyEsc, "?", "h":
		return nil // Close help view
	}
	return h
}

// View renders the help view
func (h *helpView) View(viewWidth int) string {
	var sb strings.Builder

	start, end := h.list.VisibleRange(len(h.lines))
	scrollbar := h.list.RenderScrollbar(len(h.lines))

	for i := start; i < end; i++ {
		line := h.lines[i]

		// Truncate if too long
		maxLen := viewWidth - 2 // leave room for scrollbar
		lineLen := utf8.RuneCountInString(line)
		if lineLen > maxLen {
			runes := []rune(line)
			line = string(runes[:maxLen-3]) + "..."
			lineLen = maxLen
		}

		// Pad to align scrollbar
		if lineLen < maxLen {
			line += strings.Repeat(" ", maxLen-lineLen)
		}

		scrollIdx := i - start
		sb.WriteString(line + " " + scrollbar[scrollIdx] + "\n")
	}

	// Pad remaining lines if content is shorter than viewport
	for i := end - start; i < h.list.viewHeight; i++ {
		scrollIdx := i
		scrollChar := " "
		if scrollIdx < len(scrollbar) {
			scrollChar = scrollbar[scrollIdx]
		}
		sb.WriteString(strings.Repeat(" ", viewWidth-1) + scrollChar + "\n")
	}

	return sb.String()
}
