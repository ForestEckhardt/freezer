package freezer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ForestEckhardt/freezer/github"
	"github.com/paketo-buildpacks/packit/vacation"
)

//go:generate faux --interface GitReleaseFetcher --output fakes/git_release_fetcher.go
type GitReleaseFetcher interface {
	Get(org, repo string) (github.Release, error)
	GetReleaseAsset(asset github.ReleaseAsset) (io.ReadCloser, error)
	GetReleaseTarball(url string) (io.ReadCloser, error)
}

//go:generate faux --interface Packager --output fakes/packager.go
type Packager interface {
	Execute(buildpackDir, output, version string, cached bool) error
}

//go:generate faux --interface BuildpackCache --output fakes/buildpack_cache.go
type BuildpackCache interface {
	Get(key string) (CacheEntry, bool, error)
	Set(key string, cachedEntry CacheEntry) error
	Dir() string
}

type RemoteFetcher struct {
	buildpackCache    BuildpackCache
	gitReleaseFetcher GitReleaseFetcher
	packager          Packager
	fileSystem        FileSystem
}

func NewRemoteFetcher(buildpackCache BuildpackCache, gitReleaseFetcher GitReleaseFetcher, packager Packager, fileSystem FileSystem) RemoteFetcher {
	return RemoteFetcher{
		buildpackCache:    buildpackCache,
		gitReleaseFetcher: gitReleaseFetcher,
		packager:          packager,
		fileSystem:        fileSystem,
	}
}

func (r RemoteFetcher) Get(buildpack RemoteBuildpack) (string, error) {
	release, err := r.gitReleaseFetcher.Get(buildpack.Org, buildpack.Repo)
	if err != nil {
		return "", err
	}

	buildpackCacheDir := filepath.Join(r.buildpackCache.Dir(), buildpack.Org, buildpack.Repo)
	if buildpack.Offline {
		buildpackCacheDir = filepath.Join(buildpackCacheDir, "cached")
	}

	key := buildpack.UncachedKey
	if buildpack.Offline {
		key = buildpack.CachedKey
	}

	cachedEntry, exist, err := r.buildpackCache.Get(key)
	if err != nil {
		return "", err
	}

	if !exist {
		err = os.MkdirAll(buildpackCacheDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	path := cachedEntry.URI

	if release.TagName != cachedEntry.Version || !exist {
		missingReleaseArtifacts := !(len(release.Assets) > 0)
		var bundle io.ReadCloser
		if missingReleaseArtifacts || buildpack.Offline {
			bundle, err = r.gitReleaseFetcher.GetReleaseTarball(release.TarballURL)
			if err != nil {
				return "", err
			}
		} else {
			bundle, err = r.gitReleaseFetcher.GetReleaseAsset(release.Assets[0])
			if err != nil {
				return "", err
			}
		}

		path = filepath.Join(buildpackCacheDir, fmt.Sprintf("%s.tgz", release.TagName))

		if missingReleaseArtifacts || buildpack.Offline {
			downloadDir, err := r.fileSystem.TempDir("", buildpack.Repo)
			if err != nil {
				return "", err
			}
			defer os.RemoveAll(downloadDir)

			err = vacation.NewTarGzipArchive(bundle).StripComponents(1).Decompress(downloadDir)
			if err != nil {
				return "", err
			}

			err = r.packager.Execute(downloadDir, path, release.TagName, buildpack.Offline)
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

		err = r.buildpackCache.Set(key, CacheEntry{
			Version: release.TagName,
			URI:     path,
		})

		if err != nil {
			return "", err
		}

	}

	return path, nil
}
