package main

import tea "github.com/charmbracelet/bubbletea"

type model struct {
	paths        []pathEntry
	originalPath string // PATH at startup, for "don't save" case
	list         listState
	viewWidth    int
	prompt       *prompt
	browser      *browser  // directory browser for editing paths
	helpView     *helpView // help screen
	saveChanges  bool      // true if user chose to save changes
	registryMode bool      // true when reading from Windows registry (system/user split)
	elevated     bool      // true if running with administrator privileges (Windows)
}

func initialModel(registryMode bool) model {
	var paths []pathEntry
	if registryMode && supportsRegistry {
		paths = loadPathsFromRegistry()
	} else {
		paths = loadPathsFromEnv()
	}

	return model{
		paths:        paths,
		originalPath: buildPathString(paths),
		list: listState{
			viewHeight: 20,
		},
		viewWidth:    80,
		registryMode: registryMode && supportsRegistry,
		elevated:     isElevated(),
	}
}

// hasModifications returns true if any path entry has been modified, deleted, or added
func (m model) hasModifications() bool {
	for _, p := range m.paths {
		if p.modified || p.deleted || p.added {
			return true
		}
	}
	return false
}

func (m model) Init() tea.Cmd {
	return nil
}
