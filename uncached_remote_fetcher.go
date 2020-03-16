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

//go:generate faux --interface JamPackager --output fakes/jam_packager.go
type JamPackager interface {
	Execute(args []string) error
}

type UncachedRemoteFetcher struct {
	cacheManager      *CacheManager
	gitReleaseFetcher GitReleaseFetcher
	transport         Transport
	jamPackager       JamPackager
}

func NewUncachedRemoteFetcher(cacheManager *CacheManager, gitReleaseFetcher GitReleaseFetcher, transport Transport, jamPackager JamPackager) UncachedRemoteFetcher {
	return UncachedRemoteFetcher{
		cacheManager:      cacheManager,
		gitReleaseFetcher: gitReleaseFetcher,
		transport:         transport,
		jamPackager:       jamPackager,
	}
}

func (r UncachedRemoteFetcher) Get(remoteBuildpack RemoteBuildpack) error {
	release, err := r.gitReleaseFetcher.Get(remoteBuildpack.org, remoteBuildpack.repo)
	if err != nil {
		return err
	}

	if len(release.Assets) != 1 {
		//TODO: this will download the repo and build from source
		panic("special error")
	}

	cachedEntry, exist := r.cacheManager.Get(remoteBuildpack.uncachedKey)

	if release.TagName != cachedEntry.Version || !exist {
		bundle, err := r.transport.Drop("", release.Assets[0].BrowserDownloadURL)
		if err != nil {
			return err
		}

		buildpackCacheDir := filepath.Join(r.cacheManager.cacheDir, remoteBuildpack.org, remoteBuildpack.repo)
		err = os.MkdirAll(buildpackCacheDir, os.ModePerm)
		if err != nil {
			return err
		}

		path := filepath.Join(buildpackCacheDir, fmt.Sprintf("%s.tgz", release.TagName))

		file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, bundle)
		if err != nil {
			return err
		}

		err = r.cacheManager.Set(remoteBuildpack.uncachedKey, CacheEntry{
			Version: release.TagName,
			URI:     path,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
