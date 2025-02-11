package freezer

import (
	"fmt"
)

type RemoteBuildpack struct {
	Org         string
	Repo        string
	Platform    string
	Arch        string
	UncachedKey string
	CachedKey   string
	Offline     bool
	Version     string
}

func NewRemoteBuildpack(org, repo, platform, arch string) RemoteBuildpack {
	return RemoteBuildpack{
		Org:         org,
		Repo:        repo,
		Platform:    platform,
		Arch:        arch,
		UncachedKey: fmt.Sprintf("%s:%s:%s:%s", org, repo, platform, arch),
		CachedKey:   fmt.Sprintf("%s:%s:%s:%s:cached", org, repo, platform, arch),
	}
}
