package freezer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ForestEckhardt/freezer/github"
)

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
	gitReleaseFetcher GitReleaseFetcher
	transport         Transport
}

func NewRemoteBuildpack(org, repo string, cacheManager *CacheManager, gitReleaseFetcher GitReleaseFetcher, transport Transport) RemoteBuildpack {
	return RemoteBuildpack{
		org:               org,
		repo:              repo,
		cacheManager:      cacheManager,
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
		panic("special error")
	}

	cacheKey := fmt.Sprintf("%s:%s", r.org, r.repo)

	cachedEntry, exist := r.cacheManager.Cache[cacheKey]

	if !exist {
		err := os.MkdirAll(filepath.Join(r.cacheManager.CacheDir, r.org, r.repo), os.ModePerm)
		if err != nil {
			panic(err)
		}

		r.cacheManager.Cache = CacheDB{
			cacheKey: CacheEntry{},
		}
	}

	if release.TagName != cachedEntry.Version {
		bundle, err := r.transport.Drop("", release.Assets[0].BrowserDownloadURL)
		if err != nil {
			panic(err)
		}

		path := filepath.Join(r.cacheManager.CacheDir, r.org, r.repo, fmt.Sprintf("%s.tgz", release.TagName))

		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		_, err = io.Copy(file, bundle)
		if err != nil {
			return err
		}

		r.cacheManager.Cache[cacheKey] = CacheEntry{
			Version: release.TagName,
			URI:     path,
		}

		os.RemoveAll(cachedEntry.URI)
	}

	return nil
}
