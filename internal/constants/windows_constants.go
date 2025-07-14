//go:build windows
// +build windows

package constants

import (
	"os"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/redjax/syst/internal/utils"
)

// Windows API DLL and procedures
var (
	modkernel32              = syscall.NewLazyDLL("kernel32.dll")
	procGlobalMemoryStatusEx = modkernel32.NewProc("GlobalMemoryStatusEx")
	procGetTickCount64       = modkernel32.NewProc("GetTickCount64")
)

type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

func platformConstants() PlatformConstants {
	return PlatformConstants{
		PackageManager: detectWindowsPackageManager(),
		Family:         "windows",
		Distribution:   "windows",
		Release:        windowsRelease(),

		Architecture: runtime.GOARCH,
		Hostname:     getHostname(),
		CPUModel:     getCPUModel(),
		CPUCount:     runtime.NumCPU(),
		TotalRAM:     getTotalRAM(),
		DefaultShell: getDefaultShell(),
		HomeDir:      getHomeDir(),
		Uptime:       getUptime(),
		Filesystem:   getRootFilesystemType(),
	}
}

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
	return "winget"
}

func windowsRelease() string {
	// Simplified: you can expand this by reading registry or ver command
	return "windows"
}

func getHostname() string {
	hn, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hn
}

func getCPUModel() string {
	if val := os.Getenv("PROCESSOR_IDENTIFIER"); val != "" {
		return val
	}
	return "unknown"
}

func getTotalRAM() uint64 {
	var memStatus memoryStatusEx
	memStatus.dwLength = uint32(unsafe.Sizeof(memStatus))
	ret, _, err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memStatus)))
	if ret == 0 || err != syscall.Errno(0) {
		return 0
	}
	return memStatus.ullTotalPhys
}

func getDefaultShell() string {
	return "powershell"
}

func getHomeDir() string {
	usr, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return usr
}

func getUptime() time.Duration {
	ret, _, _ := procGetTickCount64.Call()
	return time.Duration(ret) * time.Millisecond
}

func getRootFilesystemType() string {
	// Windows typically uses NTFS; detecting programmatically is complex.
	return "NTFS"
}
