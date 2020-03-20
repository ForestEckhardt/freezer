package freezer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ForestEckhardt/freezer/github"
	"github.com/cloudfoundry/packit/vacation"
)

//go:generate faux --interface GitReleaseFetcher --output fakes/git_release_fetcher.go
type GitReleaseFetcher interface {
	Get(org, repo string) (github.Release, error)
}

//go:generate faux --interface Transport --output fakes/transport.go
type Transport interface {
	Drop(root, uri string) (io.ReadCloser, error)
}

//go:generate faux --interface Packager --output fakes/packager.go
type Packager interface {
	Execute(buildpackDir, output, version string) error
}

//go:generate faux --interface BuildpackCache --output fakes/buildpack_cache.go
type BuildpackCache interface {
	Get(key string) (CacheEntry, bool)
	Set(key string, cachedEntry CacheEntry) error
	Dir() string
}

type RemoteFetcher struct {
	buildpackCache    BuildpackCache
	gitReleaseFetcher GitReleaseFetcher
	transport         Transport
	packager          Packager
	fileSystem        FileSystem
}

func NewRemoteFetcher(buildpackCache BuildpackCache, gitReleaseFetcher GitReleaseFetcher, transport Transport, packager Packager, fileSystem FileSystem) RemoteFetcher {
	return RemoteFetcher{
		buildpackCache:    buildpackCache,
		gitReleaseFetcher: gitReleaseFetcher,
		transport:         transport,
		packager:          packager,
		fileSystem:        fileSystem,
	}
}

func (r RemoteFetcher) Get(remoteBuildpack RemoteBuildpack) (string, error) {
	var path string

	release, err := r.gitReleaseFetcher.Get(remoteBuildpack.org, remoteBuildpack.repo)
	if err != nil {
		return "", err
	}

	buildpackCacheDir := filepath.Join(r.buildpackCache.Dir(), remoteBuildpack.org, remoteBuildpack.repo)

	cachedEntry, exist := r.buildpackCache.Get(remoteBuildpack.uncachedKey)
	if !exist {
		err = os.MkdirAll(buildpackCacheDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	path = cachedEntry.URI

	missingReleaseArtifacts := !(len(release.Assets) > 0)
	if release.TagName != cachedEntry.Version || !exist {
		var url string
		if missingReleaseArtifacts {
			url = release.TarballURL
		} else {
			url = release.Assets[0].BrowserDownloadURL
		}

		bundle, err := r.transport.Drop("", url)
		if err != nil {
			return "", err
		}

		path = filepath.Join(buildpackCacheDir, fmt.Sprintf("%s.tgz", release.TagName))

		if missingReleaseArtifacts {
			downloadDir, err := r.fileSystem.TempDir("", remoteBuildpack.repo)
			if err != nil {
				return "", err
			}
			defer os.RemoveAll(downloadDir)

			err = vacation.NewTarGzipArchive(bundle).Decompress(downloadDir)
			if err != nil {
				return "", err
			}

			// This strips one layer of the directories off to compensate for the file format
			// given to use by github maybe try and find a more elegant solution to this if it
			// matters.
			files, err := filepath.Glob(filepath.Join(downloadDir, "*", "*"))
			if err != nil {
				return "", err
			}

			for _, f := range files {
				err := os.Rename(f, filepath.Join(downloadDir, filepath.Base(f)))
				if err != nil {
					return "", err
				}
			}

			err = r.packager.Execute(downloadDir, path, release.TagName)
			if err != nil {
				return "", err
			}

		} else {
			file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				return "", err
			}
			defer file.Close()

			_, err = io.Copy(file, bundle)
			if err != nil {
				return "", err
			}
		}

		err = r.buildpackCache.Set(remoteBuildpack.uncachedKey, CacheEntry{
			Version: release.TagName,
			URI:     path,
		})

		if err != nil {
			return "", err
		}

	}

	return path, nil
}
