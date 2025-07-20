package platformservice

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
)

type DiskInfo struct {
	MountPoint  string
	FSType      string
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
	Label       string
}

func (p PlatformInfo) PrintDiskFormat() string {
	var b strings.Builder
	b.WriteString("Disk Information:\n")

	for _, d := range p.Disks {
		b.WriteString(fmt.Sprintf("  - Mount: %s\n", d.MountPoint))
		b.WriteString(fmt.Sprintf("    Type:  %s\n", d.FSType))
		b.WriteString(fmt.Sprintf("    Total: %.2f GB\n", float64(d.Total)/(1024*1024*1024)))
		b.WriteString(fmt.Sprintf("    Used:  %.2f GB (%.1f%%)\n", float64(d.Used)/(1024*1024*1024), d.UsedPercent))
		b.WriteString(fmt.Sprintf("    Free:  %.2f GB\n\n", float64(d.Free)/(1024*1024*1024)))
	}

	return b.String()
}

func detectDisks(verbose bool) []DiskInfo {
	var result []DiskInfo

	// System/dynamic/ignored FS types (non-verbose)
	ignoreFSTypes := map[string]bool{
		"autofs": true, "binfmt_misc": true, "cgroup": true, "cgroup2": true,
		"debugfs": true, "devpts": true, "devtmpfs": true, "efivarfs": true,
		"fusectl": true, "mqueue": true, "proc": true, "pstore": true,
		"securityfs": true, "sysfs": true, "tmpfs": true, "overlay": true,
		"tracefs": true, "nsfs": true, "ramfs": true, "squashfs": true,
		"aufs": true, "snap": true,
	}

	partitions, err := disk.Partitions(true)

	if err != nil {
		return result
	}

	for _, part := range partitions {
		if !verbose {
			if ignoreFSTypes[part.Fstype] {
				continue
			}

			// Fallback: skip pseudo mountpoints
			if strings.HasPrefix(part.Mountpoint, "/snap") ||
				strings.HasPrefix(part.Mountpoint, "/boot/efi") ||
				strings.HasPrefix(part.Mountpoint, "/var/lib/docker") ||
				strings.HasPrefix(part.Mountpoint, "/dev/") ||
				strings.HasPrefix(part.Mountpoint, "/proc") ||
				strings.HasPrefix(part.Mountpoint, "/sys") {
				continue
			}
		}

		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue
		}

		result = append(result, DiskInfo{
			MountPoint:  part.Mountpoint,
			FSType:      part.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
			Label:       "", // optional
		})
	}

	return result
}
