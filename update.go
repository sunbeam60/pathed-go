package main

import (
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

// saveAndQuitMsg is sent when user chooses a save option from the quit prompt
type saveAndQuitMsg struct {
	saveType int // 0 = discard, 1 = output edited (env mode), 2 = persist to registry
}

func doSaveAndQuit(saveType int) tea.Cmd {
	return func() tea.Msg {
		return saveAndQuitMsg{saveType: saveType}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewWidth = msg.Width
		// Subtract lines for: help bar (1) + warning if not elevated in registry mode (1)
		reservedLines := 1
		if m.registryMode && !m.elevated {
			reservedLines = 2
		}
		height := msg.Height - reservedLines
		if height < 1 {
			height = 1
		}
		m.list.SetViewHeight(height, len(m.paths))
		// Also update browser's list state if active (headerRows handles the header)
		if m.browser != nil {
			m.browser.list.SetViewHeight(height, len(m.browser.entries))
		}
		return m, nil

	case saveAndQuitMsg:
		switch msg.saveType {
		case 1: // Env mode: output edited PATH
			m.saveChanges = true
		case 2: // Registry mode: persist to registry (handled in main.go after TUI exits)
			m.saveChanges = true
		default: // Discard changes
			m.saveChanges = false
		}
		return m, tea.Quit

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
		// Changes exist, ask what to do
		if m.registryMode {
			// Registry mode: persist to registry or discard
			m.prompt = newPrompt("Persist changes to registry?", []string{"Persist", "Don't persist"}, func(index int) tea.Cmd {
				if index == 0 {
					return doSaveAndQuit(2) // persist to registry
				}
				return doSaveAndQuit(0) // discard
			})
		} else {
			// Env mode: output edited or original PATH
			m.prompt = newPrompt("Output edited PATH?", []string{"Edited", "Original"}, func(index int) tea.Cmd {
				if index == 0 {
					return doSaveAndQuit(1) // output edited
				}
				return doSaveAndQuit(0) // output original
			})
		}

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
		// Move entry up (within section in registry mode, free in env mode)
		canMove := m.list.cursor > 0
		if m.registryMode {
			canMove = canMove && m.paths[m.list.cursor-1].source == m.paths[m.list.cursor].source
		}
		if canMove {
			m.paths[m.list.cursor], m.paths[m.list.cursor-1] = m.paths[m.list.cursor-1], m.paths[m.list.cursor]
			m.list.MoveUp()
			m.paths[m.list.cursor].modified = true // mark the moved entry (now at new position)
		}

	case keyMoveDn, keyMoveDnAlt:
		// Move entry down (within section in registry mode, free in env mode)
		canMove := m.list.cursor < len(m.paths)-1
		if m.registryMode {
			canMove = canMove && m.paths[m.list.cursor+1].source == m.paths[m.list.cursor].source
		}
		if canMove {
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
		// Add new PATH entry (user entry in registry mode, no source in env mode)
		if m.registryMode {
			m.browser = newBrowserForAdd("user", m.list.TotalHeight())
		} else {
			m.browser = newBrowserForAdd("", m.list.TotalHeight())
		}

	case keyAddSystem:
		// Add new system PATH entry (registry mode only)
		if m.registryMode {
			m.browser = newBrowserForAdd("system", m.list.TotalHeight())
		}

	case keyClean:
		// Mark duplicates and non-existing paths for deletion
		// In registry mode: duplicates within same source
		// In env mode: duplicates globally
		seen := make(map[string]map[string]bool) // source -> normalized path -> seen
		seen["system"] = make(map[string]bool)
		seen["user"] = make(map[string]bool)
		seen[""] = make(map[string]bool) // for env mode

		for i := range m.paths {
			p := &m.paths[i]
			// Check for non-existing path
			if !p.exists {
				p.deleted = true
				p.modified = true
			}
			// Check for duplicate (within same source in registry mode, globally in env mode)
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
