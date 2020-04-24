package freezer_test

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testCacheManager(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		cacheDir string

		cacheManager freezer.CacheManager
	)

	it.Before(func() {
		var err error
		cacheDir, err = ioutil.TempDir("", "cache")
		Expect(err).ToNot(HaveOccurred())

		cacheManager = freezer.NewCacheManager(cacheDir)
	})

	it.After(func() {
		Expect(os.RemoveAll(cacheDir)).To(Succeed())
	})

	context("Open", func() {
		context("when Open is called on the cache manager and there is a buildpacks-cache.db file present", func() {
			var inputMap freezer.CacheDB

			it.Before(func() {
				inputMap = freezer.CacheDB{"buildpack": freezer.CacheEntry{Version: "1.2.3", URI: "some-uri"}}

				b := bytes.NewBuffer(nil)
				err := gob.NewEncoder(b).Encode(&inputMap)
				Expect(err).ToNot(HaveOccurred())

				Expect(ioutil.WriteFile(filepath.Join(cacheDir, "buildpacks-cache.db"), b.Bytes(), os.ModePerm))
			})

			it("returns the cache map stored in the buildpacks-cache.db folder", func() {
				err := cacheManager.Open()
				Expect(err).ToNot(HaveOccurred())

				Expect(cacheManager.Cache).To(Equal(inputMap))
			})
		})

		context("when Open is called on the cache manager and there is no buildpacks-cache.db file present", func() {
			it("returns an empty cache map", func() {
				err := cacheManager.Open()
				Expect(err).ToNot(HaveOccurred())

				Expect(cacheManager.Cache).To(Equal(freezer.CacheDB{}))
			})
		})

		context("failure cases", func() {
			context("the buildpacks-cache.db file is unable to be created", func() {
				it.Before(func() {
					Expect(os.Chmod(cacheDir, 0222)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(cacheDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := cacheManager.Open()
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("unable to open the buildpack-cache.db", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(filepath.Join(cacheDir, "buildpacks-cache.db"), []byte{}, 0000))
				})
				it("returns an error", func() {
					err := cacheManager.Open()
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("unable to open the buildpack-cache.db", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(filepath.Join(cacheDir, "buildpacks-cache.db"), []byte(`%%%`), os.ModePerm))
				})
				it("returns an error", func() {
					err := cacheManager.Open()
					Expect(err).To(MatchError(ContainSubstring("unexpected EOF")))
				})
			})
		})
	})

	context("Close", func() {
		context("when Close is called on the cache manager", func() {
			it.Before(func() {
				err := cacheManager.Open()
				Expect(err).ToNot(HaveOccurred())
				cacheManager.Cache = freezer.CacheDB{"some-buildpack": freezer.CacheEntry{Version: "1.2.3", URI: "some-uri"}}
			})

			it("saves the cache map given", func() {
				err := cacheManager.Close()
				Expect(err).ToNot(HaveOccurred())

				var cacheCheck freezer.CacheDB
				file, err := os.Open(filepath.Join(cacheDir, "buildpacks-cache.db"))
				Expect(err).ToNot(HaveOccurred())

				err = gob.NewDecoder(file).Decode(&cacheCheck)
				Expect(err).ToNot(HaveOccurred())

				Expect(cacheCheck).To(Equal(cacheManager.Cache))
			})
		})
	})

	context("Get", func() {
		var uri string

		it.Before(func() {
			err := cacheManager.Open()
			Expect(err).ToNot(HaveOccurred())

			uri = filepath.Join(cacheDir, "some-uri")
			cacheManager.Cache = freezer.CacheDB{"some-buildpack": freezer.CacheEntry{Version: "1.2.3", URI: uri}}
		})

		context("when the key exists", func() {
			context("and the file in uri exists", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(uri, []byte(`some-content`), 0644)).To(Succeed())
				})

				it("returns the entry and ok", func() {
					entry, ok, err := cacheManager.Get("some-buildpack")
					Expect(err).NotTo(HaveOccurred())
					Expect(ok).To(BeTrue())
					Expect(entry).To(Equal(freezer.CacheEntry{Version: "1.2.3", URI: uri}))
				})
			})

			context("and the file in uri does not exists", func() {
				it("returns the entry and not ok", func() {
					entry, ok, err := cacheManager.Get("some-buildpack")
					Expect(err).NotTo(HaveOccurred())
					Expect(ok).To(BeFalse())
					Expect(entry).To(Equal(freezer.CacheEntry{Version: "1.2.3", URI: uri}))
				})
			})
		})

		context("when the does not key exist", func() {
			it("returns with an empty entry and not ok", func() {
				entry, ok, err := cacheManager.Get("some-buildpack-other")
				Expect(err).NotTo(HaveOccurred())
				Expect(ok).To(BeFalse())
				Expect(entry).To(Equal(freezer.CacheEntry{}))
			})
		})

		context("failure cases", func() {
			context("the cached file is cannot be stated", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(uri, []byte(`some-content`), 0644)).To(Succeed())

					Expect(os.Chmod(cacheDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(cacheDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					_, _, err := cacheManager.Get("some-buildpack")
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})

	context("Set", func() {
		var uri string

		it.Before(func() {
			err := cacheManager.Open()
			Expect(err).ToNot(HaveOccurred())

			uri = filepath.Join(cacheDir, "some-file")

			Expect(ioutil.WriteFile(uri, []byte(`some content`), 0644)).To(Succeed())

			cacheManager.Cache = freezer.CacheDB{"some-buildpack": freezer.CacheEntry{Version: "1.2.3", URI: uri}}
		})

		context("when there is an already existing entry", func() {
			it("deletes the previous file and sets the new information", func() {
				err := cacheManager.Set("some-buildpack", freezer.CacheEntry{Version: "1.2.4", URI: "some-uri"})
				Expect(err).NotTo(HaveOccurred())

				Expect(uri).NotTo(BeAnExistingFile())
				Expect(cacheManager.Cache["some-buildpack"]).To(Equal(freezer.CacheEntry{Version: "1.2.4", URI: "some-uri"}))
			})
		})

		context("when there is not an already existing entry", func() {
			it("deletes the previous file and sets the new information", func() {
				err := cacheManager.Set("some-buildpack-other", freezer.CacheEntry{Version: "1.2.4", URI: "some-uri"})
				Expect(err).NotTo(HaveOccurred())

				Expect(uri).To(BeAnExistingFile())
				Expect(cacheManager.Cache["some-buildpack-other"]).To(Equal(freezer.CacheEntry{Version: "1.2.4", URI: "some-uri"}))
			})
		})

		context("failure cases", func() {
			context("when the previous entry file cannot be removed", func() {
				it.Before(func() {
					Expect(os.Chmod(cacheDir, 0000)).To(Succeed())
				})

				it.After(func() {
					Expect(os.Chmod(cacheDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := cacheManager.Set("some-buildpack", freezer.CacheEntry{Version: "1.2.4", URI: "some-uri"})
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("when the cache is nil meaning that the database has not been opened", func() {
				it.Before(func() {
					cacheManager.Cache = nil
				})

				it("returns an error", func() {
					err := cacheManager.Set("some-buildpack", freezer.CacheEntry{Version: "1.2.4", URI: "some-uri"})
					Expect(err).To(MatchError("the cache manager is not loaded properly"))

				})
			})
		})
	})
}
