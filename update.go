package main

import tea "github.com/charmbracelet/bubbletea"

// Key bindings
const (
	keyQuit      = "q"
	keyForceQuit = "ctrl+c"
	keyUp        = "up"
	keyUpAlt     = "k"
	keyDown      = "down"
	keyDownAlt   = "j"
	keyLeft      = "left"
	keyRight     = "right"
	keyPgDown    = "pgdown"
	keyPgDownAlt = "ctrl+d"
	keyPgUp      = "pgup"
	keyPgUpAlt   = "ctrl+u"
	keyHome      = "home"
	keyHomeAlt   = "g"
	keyEnd       = "end"
	keyEndAlt    = "G"
	keyMoveUp    = "shift+up"
	keyMoveUpAlt = "K"
	keyMoveDn    = "shift+down"
	keyMoveDnAlt = "J"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewWidth = msg.Width
		m.viewHeight = msg.Height - 2
		if m.viewHeight < 1 {
			m.viewHeight = 1
		}
		// Reduce offset if there's unnecessary blank space at bottom
		maxOffset := max(0, len(m.paths)-m.viewHeight)
		if m.offset > maxOffset {
			m.offset = maxOffset
		}
		// Ensure cursor stays visible after resize
		if m.cursor < m.offset {
			m.offset = m.cursor
		} else if m.cursor >= m.offset+m.viewHeight {
			m.offset = m.cursor - m.viewHeight + 1
		}
		return m, nil

	case tea.KeyMsg:
		if m.modal != nil {
			return m.updateModal(msg)
		}
		return m.updateMain(msg)
	}

	return m, nil
}

func (m model) updateModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	newModal, cmd, _ := m.modal.Update(msg)
	m.modal = newModal
	return m, cmd
}

func (m model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case keyForceQuit:
		return m, tea.Quit

	case keyQuit:
		if !m.hasModifications() {
			// No changes, quit immediately
			return m, tea.Quit
		}
		// Changes exist, ask about saving
		m.modal = newModal("Save changes?", []string{"Yes", "No"}, func(index int) tea.Cmd {
			if index == 0 {
				// TODO: Actually save to registry/environment
				// savePaths(m.paths)
			}
			return tea.Quit
		})

	case keyUp, keyUpAlt:
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.offset {
				m.offset = m.cursor
			}
		}

	case keyDown, keyDownAlt:
		if m.cursor < len(m.paths)-1 {
			m.cursor++
			if m.cursor >= m.offset+m.viewHeight {
				m.offset = m.cursor - m.viewHeight + 1
			}
		}

	case keyLeft:
		if m.hOffset > 0 {
			m.hOffset--
		}

	case keyRight:
		// Find max path length to limit scrolling
		maxLen := 0
		for _, p := range m.paths {
			if len(p.path) > maxLen {
				maxLen = len(p.path)
			}
		}
		// Calculate content width (same as in View)
		contentWidth := m.viewWidth - 4
		// Max offset is where the longest path's end aligns with the right edge
		// Add 1 to account for the left marker (<) that appears when scrolled
		maxHOffset := max(0, maxLen-contentWidth+1)
		if m.hOffset < maxHOffset {
			m.hOffset++
		}

	case keyPgDown, keyPgDownAlt:
		m.cursor += m.viewHeight
		if m.cursor >= len(m.paths) {
			m.cursor = len(m.paths) - 1
		}
		if m.cursor >= m.offset+m.viewHeight {
			m.offset = m.cursor - m.viewHeight + 1
		}

	case keyPgUp, keyPgUpAlt:
		m.cursor -= m.viewHeight
		if m.cursor < 0 {
			m.cursor = 0
		}
		if m.cursor < m.offset {
			m.offset = m.cursor
		}

	case keyHome, keyHomeAlt:
		m.cursor = 0
		m.offset = 0

	case keyEnd, keyEndAlt:
		m.cursor = len(m.paths) - 1
		maxOffset := max(0, len(m.paths)-m.viewHeight)
		m.offset = maxOffset

	case keyMoveUp, keyMoveUpAlt:
		// Move entry up within its section (system or user)
		if m.cursor > 0 && m.paths[m.cursor-1].source == m.paths[m.cursor].source {
			m.paths[m.cursor], m.paths[m.cursor-1] = m.paths[m.cursor-1], m.paths[m.cursor]
			m.cursor--
			m.paths[m.cursor].modified = true // mark the moved entry (now at new position)
			if m.cursor < m.offset {
				m.offset = m.cursor
			}
		}

	case keyMoveDn, keyMoveDnAlt:
		// Move entry down within its section (system or user)
		if m.cursor < len(m.paths)-1 && m.paths[m.cursor+1].source == m.paths[m.cursor].source {
			m.paths[m.cursor], m.paths[m.cursor+1] = m.paths[m.cursor+1], m.paths[m.cursor]
			m.cursor++
			m.paths[m.cursor].modified = true // mark the moved entry (now at new position)
			if m.cursor >= m.offset+m.viewHeight {
				m.offset = m.cursor - m.viewHeight + 1
			}
		}
	}
	return m, nil
}
