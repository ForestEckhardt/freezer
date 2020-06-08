package commands

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"

	"github.com/ForestEckhardt/freezer"
	"github.com/ForestEckhardt/freezer/github"
	"github.com/paketo-buildpacks/packit/cargo"
)

type Stock struct {
	transport  cargo.Transport
	packager   freezer.PackingTools
	fileSystem freezer.FileSystem
}

func NewStock(transport cargo.Transport, packager freezer.PackingTools) Stock {
	return Stock{
		transport:  transport,
		packager:   packager,
		fileSystem: freezer.NewFileSystem(ioutil.TempDir),
	}
}

func (s Stock) Execute(args []string) error {
	var (
		org         string
		repo        string
		cacheDir    string
		gitEndpoint string
		githubToken string
		cached      bool
	)

	usr, err := user.Current()
	if err != nil {
		return err
	}

	fset := flag.NewFlagSet("stock", flag.ContinueOnError)
	fset.StringVar(&org, "org", "", "the name of the org in the form of org/repo (eg. cloudfoundry/nodejs-cnb) (required)")
	fset.StringVar(&repo, "repo", "", "the name of the repository in the form of org/repo (eg. cloudfoundry/nodejs-cnb) (required)")
	fset.StringVar(&cacheDir, "cache-directory", filepath.Join(usr.HomeDir, ".freezer-cache"), "the location of the cache directory on disk (this will default to $HOME/.freezer-cache)")
	fset.StringVar(&gitEndpoint, "git-endpoint", "https://api.github.com", "Git endpoint url")
	fset.StringVar(&githubToken, "github-token", "", "Personal github token to prevent rate limiting (required)")
	fset.BoolVar(&cached, "cached", false, "builds a cached version of a buildpack is set to true")

	err = fset.Parse(args)
	if err != nil {
		return err
	}

	if org == "" {
		return errors.New("missing required flag --org")
	}

	if repo == "" {
		return errors.New("missing required flag --repo")
	}

	if githubToken == "" {
		return errors.New("missing required flag --github-token")
	}

	cacheManager := freezer.NewCacheManager(cacheDir)
	if err = cacheManager.Open(); err != nil {
		panic(err)
	}
	defer cacheManager.Close()

	githubReleaseService := github.NewReleaseService(github.NewConfig(gitEndpoint, githubToken))

	fetcher := freezer.NewRemoteFetcher(&cacheManager, githubReleaseService, s.transport, s.packager, s.fileSystem)

	buildpack := freezer.NewRemoteBuildpack(org, repo)
	buildpack.Offline = cached

	uri, err := fetcher.Get(buildpack)
	if err != nil {
		panic(err)
	}

	fmt.Println(uri)

	return nil
}
