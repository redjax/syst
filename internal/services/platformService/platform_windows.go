//go:build windows
// +build windows

package platformservice

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func detectOSRelease() string {
	out, err := exec.Command("cmd", "/C", "ver").Output()

	if err == nil {
		return strings.TrimSpace(string(out))
	}

	return ""
}

func detectUptime() time.Duration {
	// Use GetTickCount64 from kernel32.dll
	mod := syscall.NewLazyDLL("kernel32.dll")
	proc := mod.NewProc("GetTickCount64")

	ret, _, _ := proc.Call()

	return time.Duration(ret) * time.Millisecond
}

func detectTotalRAM() uint64 {
	// Use GlobalMemoryStatusEx from kernel32.dll
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

	var m memoryStatusEx

	m.dwLength = uint32(unsafe.Sizeof(m))
	mod := syscall.NewLazyDLL("kernel32.dll")

	proc := mod.NewProc("GlobalMemoryStatusEx")
	proc.Call(uintptr(unsafe.Pointer(&m)))

	return m.ullTotalPhys
}

// detectDefaultShell checks for PowerShell 7, then PowerShell 5, then cmd.exe
func detectDefaultShell() string {
	// Check if pwsh (PowerShell 7+) is available in PATH
	if shellPath, err := exec.LookPath("pwsh.exe"); err == nil {
		// Return just the name (can also return full path if desired)
		return filepath.Base(shellPath) // "pwsh.exe"
	}

	// Check if powershell (v5) is available
	if shellPath, err := exec.LookPath("powershell.exe"); err == nil {
		return filepath.Base(shellPath) // "powershell.exe"
	}

	// Fallback: use ComSpec (usually "cmd.exe")
	return filepath.Base(os.Getenv("ComSpec"))
}

func detectDNSServers() []string {
	out, err := exec.Command("nslookup").Output()
	if err != nil {
		return nil
	}

	var servers []string
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "address:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				servers = append(servers, fields[1])
			}
		}
	}

	return servers
}
