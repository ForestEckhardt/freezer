package freezer_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/ForestEckhardt/freezer/fakes"
	"github.com/ForestEckhardt/freezer/github"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testUncachedRemoteFetcher(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cacheDir string

		gitReleaseFetcher      *fakes.GitReleaseFetcher
		transport              *fakes.Transport
		cacheManager           freezer.CacheManager
		remoteBuildpack        freezer.RemoteBuildpack
		jamPackager            *fakes.JamPackager
		remoteBuildpackFetcher freezer.UncachedRemoteFetcher
	)

	it.Before(func() {
		var err error

		cacheDir, err = ioutil.TempDir("", "cache")
		Expect(err).NotTo(HaveOccurred())

		gitReleaseFetcher = &fakes.GitReleaseFetcher{}
		gitReleaseFetcher.GetCall.Returns.Release = github.Release{
			TagName: "some-tag",
			Assets: []github.ReleaseAsset{
				{
					BrowserDownloadURL: "some-browser-download-url",
				},
			},
			TarballURL: "some-tarball-url",
		}

		transport = &fakes.Transport{}
		buffer := bytes.NewBufferString("some content")
		transport.DropCall.Returns.ReadCloser = ioutil.NopCloser(buffer)

		jamPackager = &fakes.JamPackager{}

		cacheManager = freezer.NewCacheManager(cacheDir)

		remoteBuildpack = freezer.NewRemoteBuildpack("some-org", "some-repo")

		remoteBuildpackFetcher = freezer.NewUncachedRemoteFetcher(&cacheManager, gitReleaseFetcher, transport, jamPackager)

	})

	it.After(func() {
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
	})

	context("Get", func() {
		context("when the remote buildpack's version is out of sync with github", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(cacheDir, "some-org", "some-repo"), os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(cacheDir, "some-org", "some-repo", "some-other-tag.tgz"),
					[]byte(`some other content`),
					os.ModePerm)).To(Succeed())

				cacheManager.Cache = freezer.CacheDB{
					"some-org:some-repo": freezer.CacheEntry{
						Version: "some-other-tag",
						URI:     filepath.Join(cacheDir, "some-org", "some-repo", "some-other-tag.tgz"),
					},
				}
			})

			it("gets the latest buildpack", func() {
				err := remoteBuildpackFetcher.Get(remoteBuildpack)
				Expect(err).ToNot(HaveOccurred())

				Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
				Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

				Expect(transport.DropCall.Receives.Root).To(Equal(""))
				Expect(transport.DropCall.Receives.Uri).To(Equal("some-browser-download-url"))

				Expect(filepath.Join(cacheDir, "some-org", "some-repo", "some-other-tag.tgz")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz")).To(BeAnExistingFile())

				content, err := ioutil.ReadFile(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("some content"))

				Expect(cacheManager.Cache["some-org:some-repo"]).To(Equal(freezer.CacheEntry{
					Version: "some-tag",
					URI:     filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"),
				},
				))
			})
		})

		context("when the remote buildpack's version is in sync with github ", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(cacheDir, "some-org", "some-repo"), os.ModePerm)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"),
					[]byte(`some content`),
					os.ModePerm)).To(Succeed())

				cacheManager.Cache = freezer.CacheDB{
					"some-org:some-repo": freezer.CacheEntry{
						Version: "some-tag",
						URI:     filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"),
					},
				}
			})

			it("keeps the latest buildpack", func() {
				err := remoteBuildpackFetcher.Get(remoteBuildpack)
				Expect(err).ToNot(HaveOccurred())

				Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
				Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

				Expect(transport.DropCall.CallCount).To(Equal(0))

				Expect(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz")).To(BeAnExistingFile())

				content, err := ioutil.ReadFile(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("some content"))

				Expect(cacheManager.Cache["some-org:some-repo"]).To(Equal(freezer.CacheEntry{
					Version: "some-tag",
					URI:     filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"),
				},
				))
			})
		})

		context("when there is no cache entry", func() {
			it.Before(func() {
				Expect(cacheManager.Open()).To(Succeed())
			})

			it("keeps the latest buildpack", func() {
				err := remoteBuildpackFetcher.Get(remoteBuildpack)
				Expect(err).ToNot(HaveOccurred())

				Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
				Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

				Expect(transport.DropCall.Receives.Root).To(Equal(""))
				Expect(transport.DropCall.Receives.Uri).To(Equal("some-browser-download-url"))

				Expect(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz")).To(BeAnExistingFile())

				content, err := ioutil.ReadFile(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("some content"))

				Expect(cacheManager.Cache["some-org:some-repo"]).To(Equal(freezer.CacheEntry{
					Version: "some-tag",
					URI:     filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"),
				},
				))
			})
		})

		context("failure cases", func() {
			context("when there is a failure in the gitReleaseFetcher get", func() {
				it.Before(func() {
					gitReleaseFetcher.GetCall.Returns.Error = errors.New("unable to get release")
				})

				it("returns an error", func() {
					err := remoteBuildpackFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError("unable to get release"))
				})
			})

			context("transport drop fails", func() {
				it.Before(func() {
					transport.DropCall.Returns.Error = errors.New("drop failed")
				})

				it("returns an error", func() {
					err := remoteBuildpackFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError("drop failed"))
				})
			})
		})
	})
}
