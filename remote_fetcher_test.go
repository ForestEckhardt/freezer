package freezer_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/ForestEckhardt/freezer/fakes"
	"github.com/ForestEckhardt/freezer/github"
	"github.com/paketo-buildpacks/packit/v2/vacation"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testRemoteFetcher(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cacheDir    string
		downloadDir string
		tmpDir      string

		gitReleaseFetcher *fakes.GitReleaseFetcher
		buildpackCache    *fakes.BuildpackCache
		remoteBuildpack   freezer.RemoteBuildpack
		packager          *fakes.Packager
		fileSystem        freezer.FileSystem
		remoteFetcher     freezer.RemoteFetcher
	)

	it.Before(func() {
		var err error

		cacheDir, err = os.MkdirTemp("", "cache")
		Expect(err).NotTo(HaveOccurred())

		gitReleaseFetcher = &fakes.GitReleaseFetcher{}
		gitReleaseFetcher.GetCall.Returns.Release = github.Release{
			TagName: "some-tag",
			Assets: []github.ReleaseAsset{
				{
					URL: "some-url",
				},
			},
			TarballURL: "some-tarball-url",
		}

		buffer := bytes.NewBuffer(nil)
		gw := gzip.NewWriter(buffer)
		tw := tar.NewWriter(gw)

		Expect(tw.WriteHeader(&tar.Header{Name: "some-file", Mode: 0755, Size: int64(len("some content"))})).To(Succeed())
		_, err = tw.Write([]byte(`some content`))
		Expect(err).NotTo(HaveOccurred())

		Expect(tw.Close()).To(Succeed())
		Expect(gw.Close()).To(Succeed())

		gitReleaseFetcher.GetReleaseAssetCall.Returns.ReadCloser = io.NopCloser(buffer)
		gitReleaseFetcher.GetReleaseTarballCall.Returns.ReadCloser = io.NopCloser(buffer)

		packager = &fakes.Packager{}
		buildpackCache = &fakes.BuildpackCache{}
		buildpackCache.DirCall.Stub = func() string {
			return cacheDir
		}
		buildpackCache.GetCall.Returns.Bool = true

		remoteBuildpack = freezer.NewRemoteBuildpack("some-org", "some-repo")
		remoteBuildpack.Offline = false
		remoteBuildpack.Version = "some-version"

		tmpDir, err = os.MkdirTemp("", "tmpDir")
		Expect(err).NotTo(HaveOccurred())

		downloadDir, err = os.MkdirTemp(tmpDir, "downloadDir")
		Expect(err).NotTo(HaveOccurred())

		fileSystem = freezer.NewFileSystem(func(string, string) (string, error) {
			return downloadDir, nil
		})

		remoteFetcher = freezer.NewRemoteFetcher(buildpackCache, gitReleaseFetcher, packager, fileSystem)

	})

	it.After(func() {
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	context("Get", func() {
		context("when the remote buildpack's version is in sync with github ", func() {
			it.Before(func() {
				buildpackCache.GetCall.Returns.CacheEntry = freezer.CacheEntry{
					Version: "some-tag",
					URI:     "keep-this-uri",
				}
			})

			it("keeps the latest buildpack", func() {
				uri, err := remoteFetcher.Get(remoteBuildpack)
				Expect(err).ToNot(HaveOccurred())

				Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
				Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

				Expect(buildpackCache.GetCall.Receives.Key).To(Equal("some-org:some-repo"))

				Expect(buildpackCache.SetCall.CallCount).To(Equal(0))

				Expect(uri).To(Equal("keep-this-uri"))
			})
		})

		context("when the remote buildpack's version is out of sync with github", func() {
			it.Before(func() {
				buildpackCache.GetCall.Returns.CacheEntry = freezer.CacheEntry{
					Version: "some-other-tag",
				}

				Expect(os.MkdirAll(filepath.Join(cacheDir, "some-org", "some-repo"), os.ModePerm)).To(Succeed())
			})

			context("when there is a release artifact present", func() {
				context("when the resulting buildpack should be uncached", func() {
					it("fetches the latest buildpack", func() {
						uri, err := remoteFetcher.Get(remoteBuildpack)
						Expect(err).ToNot(HaveOccurred())

						Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
						Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

						Expect(buildpackCache.GetCall.Receives.Key).To(Equal("some-org:some-repo"))

						Expect(gitReleaseFetcher.GetReleaseAssetCall.Receives.Asset).To(Equal(github.ReleaseAsset{
							URL: "some-url",
						}))

						Expect(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz")).To(BeAnExistingFile())
						file, err := os.Open(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"))
						Expect(err).ToNot(HaveOccurred())

						err = vacation.NewArchive(file).Decompress(filepath.Join(cacheDir, "some-org", "some-repo"))
						Expect(err).ToNot(HaveOccurred())

						content, err := os.ReadFile(filepath.Join(cacheDir, "some-org", "some-repo", "some-file"))
						Expect(err).NotTo(HaveOccurred())
						Expect(string(content)).To(Equal("some content"))

						Expect(buildpackCache.SetCall.CallCount).To(Equal(1))

						Expect(uri).To(Equal(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz")))
					})
				})

				context("when the resulting buildpack should be cached", func() {
					it("fetches and builds a cached version of the latest buildpack", func() {
						remoteBuildpack.Offline = true
						uri, err := remoteFetcher.Get(remoteBuildpack)
						Expect(err).ToNot(HaveOccurred())

						Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
						Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

						Expect(buildpackCache.GetCall.Receives.Key).To(Equal("some-org:some-repo:cached"))

						Expect(gitReleaseFetcher.GetReleaseTarballCall.Receives.Url).To(Equal("some-tarball-url"))

						Expect(packager.ExecuteCall.Receives.BuildpackDir).To(Equal(downloadDir))
						Expect(packager.ExecuteCall.Receives.Output).To(Equal(filepath.Join(cacheDir, "some-org", "some-repo", "cached", "some-tag.tgz")))
						Expect(packager.ExecuteCall.Receives.Version).To(Equal("some-tag"))
						Expect(packager.ExecuteCall.Receives.Cached).To(BeTrue())

						Expect(buildpackCache.SetCall.CallCount).To(Equal(1))

						Expect(uri).To(Equal(filepath.Join(cacheDir, "some-org", "some-repo", "cached", "some-tag.tgz")))
					})
				})
			})

			context("when there is not release artifact present", func() {
				it.Before(func() {
					var err error

					gitReleaseFetcher.GetCall.Returns.Release = github.Release{
						TagName:    "some-tag",
						TarballURL: "some-tarball-url",
					}

					buildpackCache.GetCall.Returns.CacheEntry = freezer.CacheEntry{
						Version: "some-other-tag",
					}

					buffer := bytes.NewBuffer(nil)
					gw := gzip.NewWriter(buffer)
					tw := tar.NewWriter(gw)

					Expect(tw.WriteHeader(&tar.Header{Name: "some-dir", Mode: 0755, Typeflag: tar.TypeDir})).To(Succeed())
					_, err = tw.Write((nil))
					Expect(err).NotTo(HaveOccurred())

					Expect(tw.WriteHeader(&tar.Header{Name: "some-dir/some-file", Mode: 0755, Size: int64(len("some content"))})).To(Succeed())
					_, err = tw.Write([]byte(`some content`))
					Expect(err).NotTo(HaveOccurred())

					Expect(tw.Close()).To(Succeed())
					Expect(gw.Close()).To(Succeed())

					gitReleaseFetcher.GetReleaseTarballCall.Returns.ReadCloser = io.NopCloser(buffer)

					Expect(os.MkdirAll(filepath.Join(cacheDir, "some-org", "some-repo"), os.ModePerm)).To(Succeed())

					packager.ExecuteCall.Stub = func(string, string, string, bool) error {
						content, err := os.ReadFile(filepath.Join(downloadDir, "some-file"))
						if err != nil {
							return err
						}

						if string(content) != "some content" {
							return errors.New("error during decompression something is broken")
						}

						return nil
					}
				})

				context("when the resulting buildpack should be uncached", func() {
					it("fetches and builds the latest uncached buildpack", func() {
						uri, err := remoteFetcher.Get(remoteBuildpack)
						Expect(err).ToNot(HaveOccurred())

						Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
						Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

						Expect(buildpackCache.GetCall.Receives.Key).To(Equal("some-org:some-repo"))

						Expect(gitReleaseFetcher.GetReleaseTarballCall.Receives.Url).To(Equal("some-tarball-url"))

						Expect(packager.ExecuteCall.Receives.BuildpackDir).To(Equal(downloadDir))
						Expect(packager.ExecuteCall.Receives.Output).To(Equal(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz")))
						Expect(packager.ExecuteCall.Receives.Version).To(Equal("some-tag"))
						Expect(packager.ExecuteCall.Receives.Cached).To(BeFalse())

						Expect(packager.ExecuteCall.Returns.Error).To(BeNil())

						Expect(buildpackCache.SetCall.CallCount).To(Equal(1))

						Expect(uri).To(Equal(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz")))
					})
				})

				context("when the resulting buildpack should be cached", func() {
					it("fetches and builds the latest cached buildpack", func() {
						remoteBuildpack.Offline = true
						uri, err := remoteFetcher.Get(remoteBuildpack)
						Expect(err).ToNot(HaveOccurred())

						Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
						Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

						Expect(buildpackCache.GetCall.Receives.Key).To(Equal("some-org:some-repo:cached"))

						Expect(gitReleaseFetcher.GetReleaseTarballCall.Receives.Url).To(Equal("some-tarball-url"))

						Expect(packager.ExecuteCall.Receives.BuildpackDir).To(Equal(downloadDir))
						Expect(packager.ExecuteCall.Receives.Output).To(Equal(filepath.Join(cacheDir, "some-org", "some-repo", "cached", "some-tag.tgz")))
						Expect(packager.ExecuteCall.Receives.Version).To(Equal("some-tag"))
						Expect(packager.ExecuteCall.Receives.Cached).To(BeTrue())

						Expect(packager.ExecuteCall.Returns.Error).To(BeNil())

						Expect(buildpackCache.SetCall.CallCount).To(Equal(1))

						Expect(uri).To(Equal(filepath.Join(cacheDir, "some-org", "some-repo", "cached", "some-tag.tgz")))
					})
				})
			})
		})

		context("when there is no cache entry", func() {
			it.Before(func() {
				buildpackCache.GetCall.Returns.Bool = false
			})

			it("fetches the latest buildpack", func() {
				uri, err := remoteFetcher.Get(remoteBuildpack)
				Expect(err).ToNot(HaveOccurred())

				Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
				Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

				Expect(buildpackCache.GetCall.CallCount).To(Equal(1))

				Expect(gitReleaseFetcher.GetReleaseAssetCall.Receives.Asset).To(Equal(github.ReleaseAsset{
					URL: "some-url",
				}))

				Expect(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz")).To(BeAnExistingFile())
				file, err := os.Open(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz"))
				Expect(err).ToNot(HaveOccurred())

				err = vacation.NewArchive(file).Decompress(filepath.Join(cacheDir, "some-org", "some-repo"))
				Expect(err).ToNot(HaveOccurred())

				content, err := os.ReadFile(filepath.Join(cacheDir, "some-org", "some-repo", "some-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("some content"))

				Expect(buildpackCache.SetCall.CallCount).To(Equal(1))

				Expect(uri).To(Equal(filepath.Join(cacheDir, "some-org", "some-repo", "some-tag.tgz")))
			})
		})

		context("when there is a v prepending the release tag", func() {
			it.Before(func() {
				gitReleaseFetcher.GetCall.Returns.Release = github.Release{
					TagName: "v1.2.3",
					Assets: []github.ReleaseAsset{
						{
							URL: "some-url",
						},
					},
					TarballURL: "some-tarball-url",
				}

				buildpackCache.GetCall.Returns.CacheEntry = freezer.CacheEntry{
					Version: "some-other-tag",
				}

				Expect(os.MkdirAll(filepath.Join(cacheDir, "some-org", "some-repo"), os.ModePerm)).To(Succeed())
			})

			it("removes the v from the tag", func() {
				uri, err := remoteFetcher.Get(remoteBuildpack)
				Expect(err).ToNot(HaveOccurred())

				Expect(gitReleaseFetcher.GetCall.Receives.Org).To(Equal("some-org"))
				Expect(gitReleaseFetcher.GetCall.Receives.Repo).To(Equal("some-repo"))

				Expect(buildpackCache.GetCall.Receives.Key).To(Equal("some-org:some-repo"))

				Expect(gitReleaseFetcher.GetReleaseAssetCall.Receives.Asset).To(Equal(github.ReleaseAsset{
					URL: "some-url",
				}))

				Expect(filepath.Join(cacheDir, "some-org", "some-repo", "1.2.3.tgz")).To(BeAnExistingFile())
				file, err := os.Open(filepath.Join(cacheDir, "some-org", "some-repo", "1.2.3.tgz"))
				Expect(err).ToNot(HaveOccurred())

				err = vacation.NewArchive(file).Decompress(filepath.Join(cacheDir, "some-org", "some-repo"))
				Expect(err).ToNot(HaveOccurred())

				content, err := os.ReadFile(filepath.Join(cacheDir, "some-org", "some-repo", "some-file"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(content)).To(Equal("some content"))

				Expect(buildpackCache.SetCall.CallCount).To(Equal(1))

				Expect(uri).To(Equal(filepath.Join(cacheDir, "some-org", "some-repo", "1.2.3.tgz")))
			})
		})

		context("failure cases", func() {
			context("when there is a failure in the gitReleaseFetcher get", func() {
				it.Before(func() {
					gitReleaseFetcher.GetCall.Returns.Error = errors.New("unable to get release")
				})

				it("returns an error", func() {
					_, err := remoteFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError("unable to get release"))
				})
			})

			context("cache get fails", func() {
				it.Before(func() {
					buildpackCache.GetCall.Returns.Error = errors.New("failed get")
				})

				it("returns an error", func() {
					_, err := remoteFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError("failed get"))
				})
			})

			context("when getting the release tarball fails", func() {
				it.Before(func() {
					remoteBuildpack.Offline = true
					gitReleaseFetcher.GetReleaseTarballCall.Returns.Error = errors.New("unable to get release tarball")
				})

				it("returns an error", func() {
					_, err := remoteFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError("unable to get release tarball"))
				})
			})

			context("when getting the release asset fails", func() {
				it.Before(func() {
					gitReleaseFetcher.GetReleaseAssetCall.Returns.Error = errors.New("unable to get release asset")
				})

				it("returns an error", func() {
					_, err := remoteFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError("unable to get release asset"))
				})
			})

			context("when creating a temp directory fails", func() {
				it.Before(func() {
					gitReleaseFetcher.GetCall.Returns.Release = github.Release{
						TagName:    "some-tag",
						TarballURL: "some-tarball-url",
					}

					buildpackCache.GetCall.Returns.CacheEntry = freezer.CacheEntry{
						Version: "some-other-tag",
					}

					fileSystem = freezer.NewFileSystem(func(string, string) (string, error) {
						return "", errors.New("failed to create temp directory")
					})

					remoteFetcher = freezer.NewRemoteFetcher(buildpackCache, gitReleaseFetcher, packager, fileSystem)
				})

				it("returns an error", func() {
					_, err := remoteFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError("failed to create temp directory"))
				})
			})

			context("when decompression fails", func() {
				it.Before(func() {
					gitReleaseFetcher.GetCall.Returns.Release = github.Release{
						TagName:    "some-tag",
						TarballURL: "some-tarball-url",
					}

					buildpackCache.GetCall.Returns.CacheEntry = freezer.CacheEntry{
						Version: "some-other-tag",
					}
					gitReleaseFetcher.GetReleaseTarballCall.Returns.ReadCloser = io.NopCloser(bytes.NewBuffer(nil))
				})

				it("returns an error", func() {
					_, err := remoteFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError(ContainSubstring("unsupported archive type: text/plain")))
				})
			})

			context("when packing fails", func() {
				it.Before(func() {
					gitReleaseFetcher.GetCall.Returns.Release = github.Release{
						TagName:    "some-tag",
						TarballURL: "some-tarball-url",
					}

					buildpackCache.GetCall.Returns.CacheEntry = freezer.CacheEntry{
						Version: "some-other-tag",
					}
					packager.ExecuteCall.Returns.Error = errors.New("failed to package buildpack")
				})

				it("returns an error", func() {
					_, err := remoteFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError("failed to package buildpack"))
				})
			})

			context("when setting the new buildpack information failes", func() {
				it.Before(func() {
					buildpackCache.SetCall.Returns.Error = errors.New("failed to set new cache entry")

					Expect(os.MkdirAll(filepath.Join(cacheDir, "some-org", "some-repo"), os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, err := remoteFetcher.Get(remoteBuildpack)
					Expect(err).To(MatchError("failed to set new cache entry"))
				})
			})
		})
	})
}
