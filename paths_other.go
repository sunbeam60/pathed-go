//go:build !windows

package main

import (
	"os"
	"strings"
)

const supportsRegistry = false

// isElevated is a stub for non-Windows platforms (always returns true to suppress warnings)
func isElevated() bool {
	return true
}

// buildPathString constructs the PATH string from entries (excluding deleted)
func buildPathString(paths []pathEntry) string {
	var parts []string
	for _, p := range paths {
		if !p.deleted {
			parts = append(parts, p.path)
		}
	}
	return strings.Join(parts, ":")
}

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

// loadPathsFromEnv reads PATH from the process environment
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

// loadPathsFromRegistry is a stub for non-Windows platforms.
// It should never be called since supportsRegistry is false.
func loadPathsFromRegistry() []pathEntry {
	return loadPathsFromEnv()
}

// savePaths is a stub for non-Windows platforms.
func savePaths(_ []pathEntry) error {
	return nil
}
