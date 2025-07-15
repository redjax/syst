//go:build windows

package scan

import (
	"os"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// getMeta returns a file's metadata i.e. creation time & owner
func getMeta(info os.FileInfo, fullPath string) (string, string) {
	ctime := "N/A"

	if stat, ok := info.Sys().(*syscall.Win32FileAttributeData); ok {
		t := time.Unix(0, stat.CreationTime.Nanoseconds())
		ctime = t.Format("2006-01-02 15:04:05")
	}

	sd, err := windows.GetNamedSecurityInfo(
		fullPath,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION,
	)
	if err != nil {
		return ctime, "N/A"
	}

	owner, _, err := sd.Owner()
	if err != nil {
		return ctime, "N/A"
	}

	account, domain, _, err := owner.LookupAccount("")
	if err != nil {
		return ctime, "N/A"
	}

	return ctime, domain + `\` + account
}
