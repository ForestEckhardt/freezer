package freezer

import (
	"fmt"
	"os"
	"path/filepath"
)

//go:generate faux --interface Namer --output fakes/namer.go
type Namer interface {
	RandomName(name string) (string, error)
}

type LocalFetcher struct {
	buildpackCache BuildpackCache
	packager       Packager
	namer          Namer
}

func NewLocalFetcher(buildpackCache BuildpackCache, packager Packager, namer Namer) LocalFetcher {
	return LocalFetcher{
		buildpackCache: buildpackCache,
		packager:       packager,
		namer:          namer,
	}
}

func (l LocalFetcher) Get(localBuildpack LocalBuildpack, cached bool) (string, error) {
	buildpackCacheDir := filepath.Join(l.buildpackCache.Dir(), localBuildpack.name)
	if cached {
		buildpackCacheDir = filepath.Join(buildpackCacheDir, "cached")
	}

	key := localBuildpack.uncachedKey
	if cached {
		key = localBuildpack.cachedKey
	}

	name, err := l.namer.RandomName(localBuildpack.name)
	if err != nil {
		return "", fmt.Errorf("random name generation failed: %w", err)
	}

	path := filepath.Join(buildpackCacheDir, fmt.Sprintf("%s.tgz", name))

	cachedEntry, exist := l.buildpackCache.Get(key)
	if !exist {
		err := os.MkdirAll(buildpackCacheDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	} else {
		//Add locking logic or override logic
		err := os.RemoveAll(cachedEntry.URI)
		if err != nil {
			return "", err
		}
	}

	err = l.packager.Execute(localBuildpack.path, path, "testing", cached)
	if err != nil {
		return "", fmt.Errorf("failed to package buildpack: %w", err)
	}

	err = l.buildpackCache.Set(key, CacheEntry{
		Version: "testing",
		URI:     path,
	})

	if err != nil {
		return "", err
	}

	return path, nil
}
