package main

import tea "github.com/charmbracelet/bubbletea"

type model struct {
	paths     []pathEntry
	list      listState
	viewWidth int
	prompt    *prompt
	browser   *browser // directory browser for editing paths
}

func initialModel() model {
	return model{
		paths: loadPaths(),
		list: listState{
			viewHeight: 20,
		},
		viewWidth: 80,
	}
}

// hasModifications returns true if any path entry has been modified
func (m model) hasModifications() bool {
	for _, p := range m.paths {
		if p.modified {
			return true
		}
	}
	return false
}

func (m model) Init() tea.Cmd {
	return nil
}
