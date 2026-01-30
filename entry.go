package main

// pathEntry holds a path and its source (system or user)
type pathEntry struct {
	path     string
	source   string // "system" or "user"
	modified bool   // true if entry has been moved
}
