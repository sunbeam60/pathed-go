//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const supportsRegistry = true

// isElevated checks if the current process is running with administrator privileges
func isElevated() bool {
	var token windows.Token
	proc := windows.CurrentProcess()
	err := windows.OpenProcessToken(proc, windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	var elevation uint32
	var size uint32
	err = windows.GetTokenInformation(token, windows.TokenElevation, (*byte)(unsafe.Pointer(&elevation)), uint32(unsafe.Sizeof(elevation)), &size)
	if err != nil {
		return false
	}
	return elevation != 0
}

// buildPathString constructs the PATH string from entries (excluding deleted)
func buildPathString(paths []pathEntry) string {
	var parts []string
	for _, p := range paths {
		if !p.deleted {
			parts = append(parts, p.path)
		}
	}
	return strings.Join(parts, ";")
}

// normalizePath returns a normalized path for duplicate comparison.
// On Windows: case-insensitive, trailing backslashes removed.
func normalizePath(path string) string {
	// Remove trailing backslash (but keep root like "C:\")
	for len(path) > 3 && (path[len(path)-1] == '\\' || path[len(path)-1] == '/') {
		path = path[:len(path)-1]
	}
	return strings.ToLower(path)
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// loadPathsFromEnv reads PATH from the process environment (no system/user distinction)
func loadPathsFromEnv() []pathEntry {
	var entries []pathEntry
	pathEnv := os.Getenv("PATH")
	for _, p := range strings.Split(pathEnv, ";") {
		if p != "" {
			entries = append(entries, pathEntry{path: p, source: "", exists: dirExists(p)})
		}
	}
	return entries
}

// loadPathsFromRegistry reads PATH from Windows registry with system/user distinction
func loadPathsFromRegistry() []pathEntry {
	var entries []pathEntry

	// System PATH from HKLM
	sysKey, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE)
	if err == nil {
		defer sysKey.Close()
		sysPath, _, err := sysKey.GetStringValue("Path")
		if err == nil {
			for _, p := range strings.Split(sysPath, ";") {
				if p != "" {
					entries = append(entries, pathEntry{path: p, source: "system", exists: dirExists(p)})
				}
			}
		}
	}

	// User PATH from HKCU
	userKey, err := registry.OpenKey(registry.CURRENT_USER,
		`Environment`,
		registry.QUERY_VALUE)
	if err == nil {
		defer userKey.Close()
		userPath, _, err := userKey.GetStringValue("Path")
		if err == nil {
			for _, p := range strings.Split(userPath, ";") {
				if p != "" {
					entries = append(entries, pathEntry{path: p, source: "user", exists: dirExists(p)})
				}
			}
		}
	}

	return entries
}

// savePaths writes the modified paths back to the Windows registry
// Only writes to a registry key if the value has actually changed
func savePaths(paths []pathEntry) error {
	var systemPaths, userPaths []string
	for _, p := range paths {
		if p.deleted {
			continue
		}
		if p.source == "system" {
			systemPaths = append(systemPaths, p.path)
		} else {
			userPaths = append(userPaths, p.path)
		}
	}

	newSystemPath := strings.Join(systemPaths, ";")
	newUserPath := strings.Join(userPaths, ";")

	// Try to write system PATH if changed
	if err := saveSystemPath(newSystemPath); err != nil {
		return err
	}

	// Write user PATH if changed
	return saveUserPath(newUserPath)
}

// saveSystemPath writes the system PATH if it has changed
func saveSystemPath(newPath string) error {
	// Try to open with write access
	sysKey, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
			// No write access - check if we even need to write
			sysKeyRO, errRO := registry.OpenKey(registry.LOCAL_MACHINE,
				`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
				registry.QUERY_VALUE)
			if errRO == nil {
				currentPath, _, _ := sysKeyRO.GetStringValue("Path")
				sysKeyRO.Close()
				if currentPath == newPath {
					// No change needed, skip
					return nil
				}
			}
			return fmt.Errorf("access denied: run as Administrator to modify system PATH")
		}
		return fmt.Errorf("failed to open system PATH key: %w", err)
	}
	defer sysKey.Close()

	// Read current value and only write if different
	currentPath, _, _ := sysKey.GetStringValue("Path")
	if currentPath != newPath {
		if err := sysKey.SetStringValue("Path", newPath); err != nil {
			if errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
				return fmt.Errorf("access denied: run as Administrator to modify system PATH")
			}
			return fmt.Errorf("failed to write system PATH: %w", err)
		}
	}
	return nil
}

// saveUserPath writes the user PATH if it has changed
func saveUserPath(newPath string) error {
	userKey, err := registry.OpenKey(registry.CURRENT_USER,
		`Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open user PATH key: %w", err)
	}
	defer userKey.Close()

	// Read current value and only write if different
	currentPath, _, _ := userKey.GetStringValue("Path")
	if currentPath != newPath {
		if err := userKey.SetStringValue("Path", newPath); err != nil {
			return fmt.Errorf("failed to write user PATH: %w", err)
		}
	}
	return nil
}
