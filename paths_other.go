//go:build !windows

package main

import (
	"os"
	"strings"
)

const (
	addHelpText        = "a: add"
	supportsSystemPath = false
)

// normalizePath returns a normalized path for duplicate comparison.
// On Unix: case-sensitive, trailing slashes preserved (they can be significant).
func normalizePath(path string) string {
	return path
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func loadPaths() []pathEntry {
	var entries []pathEntry
	pathEnv := os.Getenv("PATH")
	for _, p := range strings.Split(pathEnv, string(os.PathListSeparator)) {
		if p != "" {
			// On non-Windows, we can't distinguish system vs user
			entries = append(entries, pathEntry{path: p, source: "user", exists: dirExists(p)})
		}
	}
	return entries
}
