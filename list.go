package main

// listState manages cursor position and viewport for a scrollable list
type listState struct {
	cursor     int
	offset     int // vertical scroll offset
	hOffset    int // horizontal scroll offset
	viewHeight int // effective list height (items visible)
	headerRows int // rows of chrome above the list (subtracted from total height)
}

// MoveUp moves the cursor up one item
func (l *listState) MoveUp() {
	if l.cursor > 0 {
		l.cursor--
		if l.cursor < l.offset {
			l.offset = l.cursor
		}
	}
}

// MoveDown moves the cursor down one item
func (l *listState) MoveDown(itemCount int) {
	if l.cursor < itemCount-1 {
		l.cursor++
		if l.cursor >= l.offset+l.viewHeight {
			l.offset = l.cursor - l.viewHeight + 1
		}
	}
}

// PageUp moves the cursor up by one page
func (l *listState) PageUp() {
	l.cursor -= l.viewHeight
	if l.cursor < 0 {
		l.cursor = 0
	}
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
}

// PageDown moves the cursor down by one page
func (l *listState) PageDown(itemCount int) {
	l.cursor += l.viewHeight
	if l.cursor >= itemCount {
		l.cursor = itemCount - 1
	}
	if l.cursor >= l.offset+l.viewHeight {
		l.offset = l.cursor - l.viewHeight + 1
	}
}

// Home moves the cursor to the first item
func (l *listState) Home() {
	l.cursor = 0
	l.offset = 0
}

// End moves the cursor to the last item
func (l *listState) End(itemCount int) {
	l.cursor = itemCount - 1
	maxOffset := max(0, itemCount-l.viewHeight)
	l.offset = maxOffset
}

// ScrollLeft scrolls the view left
func (l *listState) ScrollLeft() {
	if l.hOffset > 0 {
		l.hOffset--
	}
}

// ScrollRight scrolls the view right, limited by max content width
func (l *listState) ScrollRight(maxContentWidth, viewWidth int) {
	// Calculate content width available (account for gutter and scrollbar)
	contentWidth := viewWidth - 4
	// Max offset is where the longest content's end aligns with the right edge
	// Add 1 to account for the left marker (<) that appears when scrolled
	maxHOffset := max(0, maxContentWidth-contentWidth+1)
	if l.hOffset < maxHOffset {
		l.hOffset++
	}
}

// SetViewHeight updates the view height (accounting for headerRows) and adjusts offset if needed
func (l *listState) SetViewHeight(totalHeight, itemCount int) {
	l.viewHeight = totalHeight - l.headerRows
	if l.viewHeight < 1 {
		l.viewHeight = 1
	}
	// Reduce offset if there's unnecessary blank space at bottom
	maxOffset := max(0, itemCount-l.viewHeight)
	if l.offset > maxOffset {
		l.offset = maxOffset
	}
	// Ensure cursor stays visible after resize
	if l.cursor < l.offset {
		l.offset = l.cursor
	} else if l.cursor >= l.offset+l.viewHeight {
		l.offset = l.cursor - l.viewHeight + 1
	}
}

// TotalHeight returns the total height including header rows
func (l *listState) TotalHeight() int {
	return l.viewHeight + l.headerRows
}

// Reset resets cursor and offset to beginning
func (l *listState) Reset() {
	l.cursor = 0
	l.offset = 0
}

// VisibleRange returns the start and end indices of visible items
func (l *listState) VisibleRange(itemCount int) (start, end int) {
	start = l.offset
	end = l.offset + l.viewHeight
	if end > itemCount {
		end = itemCount
	}
	return start, end
}

// RenderScrollbar returns scrollbar characters for each visible row
func (l *listState) RenderScrollbar(itemCount int) []string {
	result := make([]string, l.viewHeight)

	if l.viewHeight >= itemCount {
		// No scrollbar needed
		for i := range result {
			result[i] = " "
		}
		return result
	}

	// Calculate thumb size (minimum 1)
	thumbSize := l.viewHeight * l.viewHeight / itemCount
	if thumbSize < 1 {
		thumbSize = 1
	}

	// Calculate thumb position
	scrollRange := itemCount - l.viewHeight
	thumbRange := l.viewHeight - thumbSize
	thumbPos := 0
	if scrollRange > 0 {
		thumbPos = l.offset * thumbRange / scrollRange
	}

	// Build scrollbar with ANSI colors
	for i := 0; i < l.viewHeight; i++ {
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
