//go:build windows

package main

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

func loadPaths() []pathEntry {
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
					entries = append(entries, pathEntry{path: p, source: "system"})
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
					entries = append(entries, pathEntry{path: p, source: "user"})
				}
			}
		}
	}

	return entries
}

// savePaths writes the modified paths back to the Windows registry
// TODO: Uncomment and implement when ready to persist changes
// func savePaths(paths []pathEntry) error {
// 	var systemPaths, userPaths []string
// 	for _, p := range paths {
// 		if p.source == "system" {
// 			systemPaths = append(systemPaths, p.path)
// 		} else {
// 			userPaths = append(userPaths, p.path)
// 		}
// 	}
//
// 	systemPath := strings.Join(systemPaths, ";")
// 	userPath := strings.Join(userPaths, ";")
//
// 	// Write system PATH (requires admin privileges)
// 	sysKey, err := registry.OpenKey(registry.LOCAL_MACHINE,
// 		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
// 		registry.SET_VALUE)
// 	if err == nil {
// 		defer sysKey.Close()
// 		sysKey.SetStringValue("Path", systemPath)
// 	}
//
// 	// Write user PATH
// 	userKey, err := registry.OpenKey(registry.CURRENT_USER,
// 		`Environment`, registry.SET_VALUE)
// 	if err == nil {
// 		defer userKey.Close()
// 		userKey.SetStringValue("Path", userPath)
// 	}
//
// 	return nil
// }
