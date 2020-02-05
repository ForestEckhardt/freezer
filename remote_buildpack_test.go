package freezer_test

import (
	"io/ioutil"
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/ForestEckhardt/freezer/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testRemoteBuildpack(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cacheDir string

		gitClient       *fakes.GitClient
		cacheManager    freezer.CacheManager
		remoteBuildpack freezer.RemoteBuildpack
	)

	it.Before(func() {
		var err error

		cacheDir, err = ioutil.TempDir("", "cache")
		Expect(err).NotTo(HaveOccurred())

		gitClient = &fakes.GitClient{}
		cacheManager = freezer.NewCacheManager(cacheDir)
		cacheManager.Open()

		remoteBuildpack = freezer.NewRemoteBuildpack("some-org", "some-repo", &cacheManager, gitClient)
	})
	context("Get", func() {
		context("when the remote buildpack's version is out of sync for github", func() {
			it("gets the latest buildpack", func() {
				err := remoteBuildpack.Get()
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
}
