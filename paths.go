package main

import (
	"os"
	"strings"
)

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// buildPathString constructs the PATH string from entries (excluding deleted)
func buildPathString(paths []pathEntry) string {
	var parts []string
	for _, p := range paths {
		if !p.deleted {
			parts = append(parts, p.path)
		}
	}
	return strings.Join(parts, string(os.PathListSeparator))
}

// loadPathsFromEnv reads PATH from the process environment (no system/user distinction)
func loadPathsFromEnv() []pathEntry {
	var entries []pathEntry
	pathEnv := os.Getenv("PATH")
	for _, p := range strings.Split(pathEnv, string(os.PathListSeparator)) {
		if p != "" {
			entries = append(entries, pathEntry{path: p, source: "", exists: dirExists(p)})
		}
	}
	return entries
}
