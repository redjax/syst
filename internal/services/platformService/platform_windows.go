//go:build windows
// +build windows

package platformservice

import (
	"os/exec"
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
