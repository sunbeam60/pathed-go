package main

import (
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewWidth = msg.Width
		height := msg.Height - 1 // subtract 1 for help bar
		if height < 1 {
			height = 1
		}
		m.list.SetViewHeight(height, len(m.paths))
		// Also update browser's list state if active (headerRows handles the header)
		if m.browser != nil {
			m.browser.list.SetViewHeight(height, len(m.browser.entries))
		}
		return m, nil

	case tea.KeyMsg:
		if m.browser != nil {
			return m.updateBrowser(msg)
		}
		if m.prompt != nil {
			return m.updatePrompt(msg)
		}
		return m.updateMain(msg)
	}

	return m, nil
}

func (m model) updatePrompt(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	newPrompt, cmd, _ := m.prompt.Update(msg)
	m.prompt = newPrompt
	return m, cmd
}

func (m model) updateBrowser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	newBrowser, cmd, selectedPath := m.browser.Update(msg)
	if newBrowser == nil {
		// Browser closed
		if selectedPath != "" {
			if m.browser.editingIndex == -1 {
				// Add mode - create new path entry
				newEntry := pathEntry{
					path:     selectedPath,
					source:   m.browser.addSource,
					modified: true,
					deleted:  false,
					added:    true,
					exists:   dirExists(selectedPath),
				}
				// Insert at appropriate position based on source
				m.paths = insertPathEntry(m.paths, newEntry)
				// Move cursor to the new entry
				for i, p := range m.paths {
					if p.path == selectedPath && p.source == newEntry.source {
						m.list.cursor = i
						break
					}
				}
			} else {
				// Edit mode - update the existing path entry
				idx := m.browser.editingIndex
				if m.paths[idx].path != selectedPath {
					m.paths[idx].path = selectedPath
					m.paths[idx].modified = true
					m.paths[idx].deleted = false // clear deletion mark when editing
					m.paths[idx].exists = dirExists(selectedPath)
				}
			}
		}
		m.browser = nil
	} else {
		m.browser = newBrowser
	}
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
		m.prompt = newPrompt("Save changes?", []string{"Yes", "No"}, func(index int) tea.Cmd {
			if index == 0 {
				// TODO: Actually save to registry/environment
				// savePaths(m.paths)
			}
			return tea.Quit
		})

	case keyUp, keyUpAlt:
		m.list.MoveUp()

	case keyDown, keyDownAlt:
		m.list.MoveDown(len(m.paths))

	case keyLeft:
		m.list.ScrollLeft()

	case keyRight:
		// Find max path length to limit scrolling (in runes, not bytes)
		maxLen := 0
		for _, p := range m.paths {
			runeLen := utf8.RuneCountInString(p.path)
			if runeLen > maxLen {
				maxLen = runeLen
			}
		}
		m.list.ScrollRight(maxLen, m.viewWidth)

	case keyPgDown, keyPgDownAlt:
		m.list.PageDown(len(m.paths))

	case keyPgUp, keyPgUpAlt:
		m.list.PageUp()

	case keyHome, keyHomeAlt:
		m.list.Home()

	case keyEnd, keyEndAlt:
		m.list.End(len(m.paths))

	case keyMoveUp, keyMoveUpAlt:
		// Move entry up within its section (system or user)
		if m.list.cursor > 0 && m.paths[m.list.cursor-1].source == m.paths[m.list.cursor].source {
			m.paths[m.list.cursor], m.paths[m.list.cursor-1] = m.paths[m.list.cursor-1], m.paths[m.list.cursor]
			m.list.MoveUp()
			m.paths[m.list.cursor].modified = true // mark the moved entry (now at new position)
		}

	case keyMoveDn, keyMoveDnAlt:
		// Move entry down within its section (system or user)
		if m.list.cursor < len(m.paths)-1 && m.paths[m.list.cursor+1].source == m.paths[m.list.cursor].source {
			m.paths[m.list.cursor], m.paths[m.list.cursor+1] = m.paths[m.list.cursor+1], m.paths[m.list.cursor]
			m.list.MoveDown(len(m.paths))
			m.paths[m.list.cursor].modified = true // mark the moved entry (now at new position)
		}

	case keyDelete:
		// Toggle deleted state on current entry
		m.paths[m.list.cursor].deleted = !m.paths[m.list.cursor].deleted

	case keySelect:
		// Open directory browser for the current entry
		m.browser = newBrowser(m.paths[m.list.cursor].path, m.list.cursor, m.list.TotalHeight())

	case keyAddUser:
		// Add new user PATH entry
		m.browser = newBrowserForAdd("user", m.list.TotalHeight())

	case keyAddSystem:
		// Add new system PATH entry (Windows only)
		if supportsSystemPath {
			m.browser = newBrowserForAdd("system", m.list.TotalHeight())
		}

	case keyClean:
		// Mark duplicates (within same source) and non-existing paths for deletion
		seen := make(map[string]map[string]bool) // source -> normalized path -> seen
		seen["system"] = make(map[string]bool)
		seen["user"] = make(map[string]bool)

		for i := range m.paths {
			p := &m.paths[i]
			// Check for non-existing path
			if !p.exists {
				p.deleted = true
				p.modified = true
			}
			// Check for duplicate within same source (using normalized path)
			normalizedPath := normalizePath(p.path)
			if seen[p.source][normalizedPath] {
				p.deleted = true
				p.modified = true
			}
			seen[p.source][normalizedPath] = true
		}
	}
	return m, nil
}
