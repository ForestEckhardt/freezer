package freezer

import "fmt"

type RemoteBuildpack struct {
	Org         string
	Repo        string
	UncachedKey string
	CachedKey   string
	Offline     bool
	Version     string
}

func NewRemoteBuildpack(org, repo string) RemoteBuildpack {
	return RemoteBuildpack{
		Org:         org,
		Repo:        repo,
		UncachedKey: fmt.Sprintf("%s:%s", org, repo),
		CachedKey:   fmt.Sprintf("%s:%s:cached", org, repo),
	}
}
