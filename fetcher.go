package freezer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ForestEckhardt/freezer/github"
	"github.com/cloudfoundry/packit/cargo"
)

type Caching int

const (
	Uncached Caching = iota
	Cached
)

type Fetcher struct {
	cacheManager         CacheManager
	githubReleaseService github.ReleaseService
	packager             PackingTools
	transport            cargo.Transport
	fileSystem           FileSystem
	localFetcher         LocalFetcher
	remoteFetcher        RemoteFetcher
	namer                NameGenerator
}

func NewFetcher() Fetcher {
	return Fetcher{
		cacheManager:         NewCacheManager(filepath.Join(os.Getenv("HOME"), ".freezer-cache")),
		githubReleaseService: github.NewReleaseService(github.NewConfig("https://api.github.com", os.Getenv("GIT_TOKEN"))),
		packager:             NewPackingTools(),
		transport:            cargo.NewTransport(),
		fileSystem:           NewFileSystem(ioutil.TempDir),
		namer:                NewNameGenerator(),
	}
}

func (f *Fetcher) Open() error {
	f.localFetcher = NewLocalFetcher(
		&f.cacheManager,
		f.packager,
		f.namer,
	)

	f.remoteFetcher = NewRemoteFetcher(
		&f.cacheManager,
		f.githubReleaseService,
		f.transport,
		f.packager,
		f.fileSystem,
	)

	return f.cacheManager.Open()
}

func (f *Fetcher) Close() error {
	return f.cacheManager.Close()
}

func (f Fetcher) Get(buildpack string, cached Caching) (string, error) {
	if strings.HasPrefix(buildpack, "github.com") {
		request := strings.SplitN(buildpack, "/", 3)
		return f.remoteFetcher.Get(NewRemoteBuildpack(request[1], request[2]), cached == Cached)
	}

	return f.localFetcher.Get(NewLocalBuildpack(buildpack, filepath.Base(buildpack)), cached == Cached)
}
