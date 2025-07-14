//go:build windows
// +build windows

package constants

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/redjax/syst/internal/utils"

	"golang.org/x/sys/windows/registry"
)

func windowsRelease() string {
	// Try reading from registry first
	if ver, err := readWindowsVersionFromRegistry(); err == nil && ver != "" {
		return ver
	}

	// Fallback: try 'ver' command output
	if ver, err := readWindowsVersionFromVerCmd(); err == nil && ver != "" {
		return ver
	}

	// Last fallback
	return "unknown"
}

func readWindowsVersionFromRegistry() (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer key.Close()

	// Try to get ProductName (e.g. "Windows 10 Pro")
	productName, _, err := key.GetStringValue("ProductName")
	if err != nil {
		return "", err
	}

	// Try to get ReleaseId or CurrentBuild for version numbers
	releaseId, _, _ := key.GetStringValue("ReleaseId")       // e.g. "2009"
	currentBuild, _, _ := key.GetStringValue("CurrentBuild") // e.g. "19042"
	ubr, _, _ := key.GetIntegerValue("UBR")                  // e.g. 1237 (update build revision)

	version := productName
	if releaseId != "" {
		version += " " + releaseId
	}
	if currentBuild != "" {
		version += " (Build " + currentBuild
		if ubr != 0 {
			version += fmt.Sprintf(".%d", ubr)
		}
		version += ")"
	}

	return version, nil
}

func readWindowsVersionFromVerCmd() (string, error) {
	out, err := exec.Command("cmd", "/C", "ver").Output()
	if err != nil {
		return "", err
	}
	ver := strings.TrimSpace(string(out))
	// Output example: "Microsoft Windows [Version 10.0.19044.1706]"
	// Clean it up to just version number:
	start := strings.Index(ver, "[Version ")
	end := strings.Index(ver, "]")
	if start >= 0 && end > start {
		return ver[start+len("[Version ") : end], nil
	}
	return ver, nil
}

// Return detected package manager for Windows.
// Defaults to winget, but prefers scoop if it's installed. Precedence is:
//  1. scoop, 2. winget, 3. choco
func detectWindowsPackageManager() string {
	if utils.IsCommandAvailable("scoop") {
		return "scoop"
	}
	if utils.IsCommandAvailable("winget") {
		return "winget"
	}
	if utils.IsCommandAvailable("choco") {
		return "choco"
	}

	// default if none found
	return "winget"
}

func platformConstants() PlatformConstants {
	return PlatformConstants{
		PackageManager: detectWindowsPackageManager(),
		Family:         "windows",
		Distribution:   "windows",
		Release:        windowsRelease(),
	}
}
