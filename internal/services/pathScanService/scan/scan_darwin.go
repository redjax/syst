//go:build darwin

package scan

import (
	"fmt"
	"os"
	"os/user"
	"syscall"
	"time"
)

// getMeta returns a file's metadata i.e. creation time & owner
func getMeta(info os.FileInfo, fullPath string) (ctime string, owner string) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		ctime = "N/A"
		owner = "N/A"
		return
	}

	t := time.Unix(stat.Birthtimespec.Sec, stat.Birthtimespec.Nsec)
	ctime = t.Format("2006-01-02 15:04:05")

	uid := fmt.Sprint(stat.Uid)
	gid := fmt.Sprint(stat.Gid)

	u, err := user.LookupId(uid)
	if err == nil {
		uid = u.Username
	}

	g, err := user.LookupGroupId(gid)
	if err == nil {
		gid = g.Name
	}

	owner = fmt.Sprintf("%s:%s", uid, gid)

	return
}
