//go:build linux

package scan

import (
	"fmt"
	"os"
	"os/user"
	"syscall"
	"time"
)

func getMeta(info os.FileInfo, fullPath string) (string, string) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "N/A", "N/A"
	}
	ctime := time.Unix(stat.Ctim.Sec, stat.Ctim.Nsec).Format("2006-01-02 15:04:05")

	uid := fmt.Sprint(stat.Uid)
	gid := fmt.Sprint(stat.Gid)

	u, _ := user.LookupId(uid)
	g, _ := user.LookupGroupId(gid)

	username := uid
	groupname := gid
	if u != nil {
		username = u.Username
	}
	if g != nil {
		groupname = g.Name
	}
	return ctime, fmt.Sprintf("%s:%s", username, groupname)
}
