//go:build !windows

package main

import (
	"os"
	"strings"
)

func loadPaths() []pathEntry {
	var entries []pathEntry
	pathEnv := os.Getenv("PATH")
	for _, p := range strings.Split(pathEnv, string(os.PathListSeparator)) {
		if p != "" {
			// On non-Windows, we can't distinguish system vs user
			entries = append(entries, pathEntry{path: p, source: "user"})
		}
	}
	return entries
}
