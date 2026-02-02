//go:build !windows

package main

const supportsRegistry = false

// isElevated is a stub for non-Windows platforms (always returns true to suppress warnings)
func isElevated() bool {
	return true
}

// normalizePath returns a normalized path for duplicate comparison.
// On Unix: case-sensitive, trailing slashes preserved (they can be significant).
func normalizePath(path string) string {
	return path
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
