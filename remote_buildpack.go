package freezer

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
)

//go:generate faux --interface GitClient --output fakes/git_client.go
type GitClient interface {
	Do(ctx context.Context, req *http.Request, v interface{}) (*github.Response, error)
}

type RemoteBuildpack struct {
	org          string
	repo         string
	cacheManager *CacheManager
	gitClient    GitClient
}

func NewRemoteBuildpack(org, repo string, cacheManager *CacheManager, gitClient GitClient) RemoteBuildpack {
	return RemoteBuildpack{
		org:          org,
		repo:         repo,
		cacheManager: cacheManager,
		gitClient:    gitClient,
	}
}

func (r RemoteBuildpack) Get() error {
	return nil
}
