package freezer_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/ForestEckhardt/freezer/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testLocalFetcher(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cacheDir string

		buildpackCache *fakes.BuildpackCache
		packager       *fakes.Packager
		namer          *fakes.Namer

		localBuildpack freezer.LocalBuildpack
		localFetcher   freezer.LocalFetcher
	)

	it.Before(func() {
		var err error

		cacheDir, err = os.MkdirTemp("", "cache")
		Expect(err).NotTo(HaveOccurred())

		packager = &fakes.Packager{}

		buildpackCache = &fakes.BuildpackCache{}
		buildpackCache.DirCall.Stub = func() string {
			return cacheDir
		}
		buildpackCache.GetCall.Returns.Bool = true

		namer = &fakes.Namer{}
		namer.RandomNameCall.Stub = func(name string) (string, error) {
			return fmt.Sprintf("%s-random-string", name), nil
		}

		localBuildpack = freezer.NewLocalBuildpack("path/to/buildpack", "some-buildpack")
		localBuildpack.Offline = false
		localBuildpack.Version = "some-version"

		localFetcher = freezer.NewLocalFetcher(buildpackCache, packager, namer)

	})

	it.After(func() {
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
	})

	context("Get", func() {
		context("when there is not already an existing file", func() {
			it("builds an uncached version of the buildpack and puts it in the cache", func() {
				uri, err := localFetcher.Get(localBuildpack)
				Expect(err).ToNot(HaveOccurred())

				Expect(buildpackCache.GetCall.CallCount).To(Equal(1))

				Expect(namer.RandomNameCall.Receives.Name).To(Equal("some-buildpack"))

				Expect(packager.ExecuteCall.Receives.BuildpackDir).To(Equal("path/to/buildpack"))
				Expect(packager.ExecuteCall.Receives.Output).To(Equal(filepath.Join(cacheDir, "some-buildpack", "some-buildpack-random-string.cnb")))
				Expect(packager.ExecuteCall.Receives.Version).To(Equal("some-version"))
				Expect(packager.ExecuteCall.Receives.Cached).To(BeFalse())

				Expect(buildpackCache.SetCall.CallCount).To(Equal(1))

				Expect(uri).To(Equal(filepath.Join(cacheDir, "some-buildpack", "some-buildpack-random-string.cnb")))
			})
		})

		context("failure cases", func() {
			context("when the namer fails to generate a random name", func() {
				it.Before(func() {
					namer.RandomNameCall.Stub = nil
					namer.RandomNameCall.Returns.Error = fmt.Errorf("namer failed")
				})

				it("returns an error", func() {
					_, err := localFetcher.Get(localBuildpack)
					Expect(err).To(MatchError("random name generation failed: namer failed"))
				})
			})

			context("cache get fails", func() {
				it.Before(func() {
					buildpackCache.GetCall.Returns.Error = errors.New("failed get")
				})

				it("returns an error", func() {
					_, err := localFetcher.Get(localBuildpack)
					Expect(err).To(MatchError("failed get"))
				})
			})

			context("unable to create new directory in cache directory", func() {
				it.Before(func() {
					buildpackCache.GetCall.Returns.Bool = false

					Expect(os.Chmod(cacheDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(cacheDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := localFetcher.Get(localBuildpack)
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("the packager fails to package the buildpack", func() {
				it.Before(func() {
					packager.ExecuteCall.Returns.Error = errors.New("execution failed")
				})

				it("returns an error", func() {
					_, err := localFetcher.Get(localBuildpack)
					Expect(err).To(MatchError("failed to package buildpack: execution failed"))
				})
			})

			context("when setting the new buildpack information failes", func() {
				it.Before(func() {
					buildpackCache.SetCall.Returns.Error = errors.New("failed to set new cache entry")

					Expect(os.MkdirAll(filepath.Join(cacheDir, "some-buildpack"), os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := localFetcher.Get(localBuildpack)
					Expect(err).To(MatchError("failed to set new cache entry"))
				})
			})
		})
	})
}
