//go:build windows

package main

import (
	"errors"
	"fmt"
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

// normalizePath returns a normalized path for duplicate comparison.
// On Windows: case-insensitive, trailing backslashes removed.
func normalizePath(path string) string {
	// Remove trailing backslash (but keep root like "C:\")
	for len(path) > 3 && (path[len(path)-1] == '\\' || path[len(path)-1] == '/') {
		path = path[:len(path)-1]
	}
	return strings.ToLower(path)
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
	if err := saveUserPath(newUserPath); err != nil {
		return err
	}

	// Notify other applications that environment has changed
	broadcastEnvironmentChange()
	return nil
}

// writeRegistryPathIfChanged compares current value and writes if different
func writeRegistryPathIfChanged(key registry.Key, newPath, desc string) error {
	currentPath, _, _ := key.GetStringValue("Path")
	if currentPath == newPath {
		return nil
	}
	if err := key.SetStringValue("Path", newPath); err != nil {
		if errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
			return fmt.Errorf("access denied: run as Administrator to modify %s PATH", desc)
		}
		return fmt.Errorf("failed to write %s PATH: %w", desc, err)
	}
	return nil
}

// saveSystemPath writes the system PATH if it has changed
func saveSystemPath(newPath string) error {
	const subKey = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`

	// Try to open with write access
	sysKey, err := registry.OpenKey(registry.LOCAL_MACHINE, subKey, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		if errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
			// No write access - check if we even need to write
			sysKeyRO, errRO := registry.OpenKey(registry.LOCAL_MACHINE, subKey, registry.QUERY_VALUE)
			if errRO == nil {
				currentPath, _, _ := sysKeyRO.GetStringValue("Path")
				sysKeyRO.Close()
				if currentPath == newPath {
					return nil // No change needed, skip
				}
			}
			return fmt.Errorf("access denied: run as Administrator to modify system PATH")
		}
		return fmt.Errorf("failed to open system PATH key: %w", err)
	}
	defer sysKey.Close()

	return writeRegistryPathIfChanged(sysKey, newPath, "system")
}

// saveUserPath writes the user PATH if it has changed
func saveUserPath(newPath string) error {
	userKey, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open user PATH key: %w", err)
	}
	defer userKey.Close()

	return writeRegistryPathIfChanged(userKey, newPath, "user")
}

// broadcastEnvironmentChange notifies all windows that environment variables have changed.
// This allows Explorer and other apps to pick up PATH changes without restart.
func broadcastEnvironmentChange() {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	// WM_SETTINGCHANGE = 0x001A, HWND_BROADCAST = 0xFFFF
	// SMTO_ABORTIFHUNG = 0x0002
	const (
		HWND_BROADCAST   = 0xFFFF
		WM_SETTINGCHANGE = 0x001A
		SMTO_ABORTIFHUNG = 0x0002
	)

	envStr, _ := syscall.UTF16PtrFromString("Environment")
	sendMessageTimeout.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(envStr)),
		uintptr(SMTO_ABORTIFHUNG),
		uintptr(5000), // 5 second timeout
		0,
	)
}
