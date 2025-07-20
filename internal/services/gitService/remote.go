package gitservice

import (
	"os"
	"os/exec"
	"strings"
)

func pruneRemotes() error {
	cmd := exec.Command("git", "remote", "update", "origin", "--prune")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	return cmd.Run()
}

func getRemotes() ([]RemoteInfo, error) {
	cmd := exec.Command("git", "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	remotesMap := make(map[string]RemoteInfo)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			name := fields[0]
			url := fields[1]
			typ := fields[2]
			info := remotesMap[name]
			info.Name = name
			if strings.Contains(typ, "(fetch)") {
				info.FetchURL = url
			}
			if strings.Contains(typ, "(push)") {
				info.PushURL = url
			}
			remotesMap[name] = info
		}
	}
	var remotes []RemoteInfo
	for _, r := range remotesMap {
		remotes = append(remotes, r)
	}
	return remotes, nil
}
