package freezer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ForestEckhardt/freezer/github"
)

//Fetecher

//go:generate faux --interface GitReleaseFetcher --output fakes/git_release_fetcher.go
type GitReleaseFetcher interface {
	Get(org, repo string) (github.Release, error)
}

//go:generate faux --interface Transport --output fakes/transport.go
type Transport interface {
	Drop(root, uri string) (io.ReadCloser, error)
}

type RemoteBuildpack struct {
	org               string
	repo              string
	cacheManager      *CacheManager
	cacheKey          string
	gitReleaseFetcher GitReleaseFetcher
	transport         Transport
}

func NewRemoteBuildpack(org, repo string, cacheManager *CacheManager, gitReleaseFetcher GitReleaseFetcher, transport Transport) RemoteBuildpack {
	return RemoteBuildpack{
		org:               org,
		repo:              repo,
		cacheManager:      cacheManager,
		cacheKey:          fmt.Sprintf("%s:%s", org, repo),
		gitReleaseFetcher: gitReleaseFetcher,
		transport:         transport,
	}
}

func (r RemoteBuildpack) Get() error {
	release, err := r.gitReleaseFetcher.Get(r.org, r.repo)
	if err != nil {
		return err
	}

	if len(release.Assets) != 1 {
		//TODO: this will download the repo and build from source
		panic("special error")
	}

	cachedEntry, exist := r.cacheManager.Get(r.cacheKey)

	if release.TagName != cachedEntry.Version || !exist {
		bundle, err := r.transport.Drop("", release.Assets[0].BrowserDownloadURL)
		if err != nil {
			return err
		}

		path := filepath.Join(r.cacheManager.cacheDir, r.org, r.repo, fmt.Sprintf("%s.tgz", release.TagName))

		err = os.MkdirAll(filepath.Join(r.cacheManager.cacheDir, r.org, r.repo), os.ModePerm)
		if err != nil {
			return err
		}

		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, bundle)
		if err != nil {
			return err
		}

		err = r.cacheManager.Set(r.cacheKey, CacheEntry{
			Version: release.TagName,
			URI:     path,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
