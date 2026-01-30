package main

// pathEntry holds a path and its source (system or user)
type pathEntry struct {
	path     string
	source   string // "system" or "user"
	modified bool   // true if entry has been changed
	deleted  bool   // true if entry is marked for deletion
	added    bool   // true if entry was added in this session
	exists   bool   // true if directory exists on disk
}

// insertPathEntry inserts a new entry at the end of its source section
func insertPathEntry(paths []pathEntry, entry pathEntry) []pathEntry {
	// Find the insertion point: end of the matching source section
	insertIdx := len(paths)
	if entry.source == "system" {
		// Insert at the end of system entries (before user entries start)
		for i, p := range paths {
			if p.source == "user" {
				insertIdx = i
				break
			}
		}
	}
	// For user entries, insertIdx stays at len(paths) - append at end

	// Insert at the found position
	paths = append(paths, pathEntry{})
	copy(paths[insertIdx+1:], paths[insertIdx:])
	paths[insertIdx] = entry
	return paths
}
