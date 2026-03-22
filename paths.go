package main

import (
	"os"
	"strings"
)

// expandEnvVars expands Windows-style %VAR% environment variable references in a string.
func expandEnvVars(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '%' {
			j := strings.Index(s[i+1:], "%")
			if j == -1 {
				result.WriteString(s[i:])
				break
			}
			varName := s[i+1 : i+1+j]
			if val := os.Getenv(varName); val != "" {
				result.WriteString(val)
			} else {
				result.WriteString(s[i : i+2+j])
			}
			i = i + 2 + j
		} else {
			result.WriteByte(s[i])
			i++
		}
	}
	return result.String()
}

// dirExists checks if a directory exists, expanding any %VAR% environment variables first.
func dirExists(path string) bool {
	info, err := os.Stat(expandEnvVars(path))
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
