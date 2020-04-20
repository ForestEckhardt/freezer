package freezer_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFetcher(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cacheDir string
		home     string

		fetcher freezer.Fetcher
	)

	it.Before(func() {
		var err error
		cacheDir, err = ioutil.TempDir("", "cacheDir")
		Expect(err).NotTo(HaveOccurred())

		home = os.Getenv("HOME")
		Expect(os.Setenv("HOME", cacheDir)).To(Succeed())

		fetcher = freezer.NewFetcher()
		Expect(fetcher.Open()).To(Succeed())
	})

	it.After(func() {
		Expect(os.Setenv("HOME", home)).To(Succeed())
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
		Expect(fetcher.Close()).To(Succeed())
	})

	context("Get", func() {
		context("when trying to get a remote buildpack uncached", func() {
			it("downloads the latest uncached buildpack", func() {
				uri, err := fetcher.Get("github.com/paketo-buildpacks/npm", freezer.Uncached)
				Expect(err).NotTo(HaveOccurred())
				Expect(uri).To(BeAnExistingFile())
			})
		})

		context("when trying to get a remote buildpack cached", func() {
			it("downloads the latest uncached buildpack source and builds a cached buildpack", func() {
				uri, err := fetcher.Get("github.com/paketo-buildpacks/npm", freezer.Cached)
				Expect(err).NotTo(HaveOccurred())
				Expect(uri).To(BeAnExistingFile())
			})
		})

		context("when trying to get a local buildpack uncached", func() {
			it("builds an uncached buildpack", func() {
				uri, err := fetcher.Get(filepath.Join("testdata", "example-cnb"), freezer.Uncached)
				Expect(err).NotTo(HaveOccurred())
				Expect(uri).To(BeAnExistingFile())
			})
		})

		context("when trying to get a local buildpack cached", func() {
			it("builds a cached buildpack", func() {
				uri, err := fetcher.Get(filepath.Join("testdata", "example-cnb"), freezer.Cached)
				Expect(err).NotTo(HaveOccurred())
				Expect(uri).To(BeAnExistingFile())
			})
		})
	})
}
