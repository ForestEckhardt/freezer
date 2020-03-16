package freezer

import "fmt"

type RemoteBuildpack struct {
	org         string
	repo        string
	uncachedKey string
	cachedKey   string
}

func NewRemoteBuildpack(org, repo string) RemoteBuildpack {
	return RemoteBuildpack{
		org:         org,
		repo:        repo,
		uncachedKey: fmt.Sprintf("%s:%s", org, repo),
		cachedKey:   fmt.Sprintf("%s:%s:cached", org, repo),
	}
}
