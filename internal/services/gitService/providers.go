package gitservice

import "fmt"

type GitProvider string

const (
	Github   GitProvider = "github"
	Gitlab   GitProvider = "gitlab"
	Codeberg GitProvider = "codeberg"
)

var providerHostMap = map[GitProvider]string{
	Github:   "github.com",
	Gitlab:   "gitlab.com",
	Codeberg: "codeberg.org",
}

func GetHostByProvider(provider string) string {
	if host, ok := providerHostMap[GitProvider(provider)]; ok {
		return host
	}

	return ""
}

func ValidateGitProvider(provider string) bool {
	_, ok := providerHostMap[GitProvider(provider)]
	return ok
}

func BuildRepoURL(protocol, host, user, repo string) string {
	if protocol == "ssh" {
		return fmt.Sprintf("git@%s:%s/%s.git", host, user, repo)
	}
	return fmt.Sprintf("https://%s/%s/%s.git", host, user, repo)
}
