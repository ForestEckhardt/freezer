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

		tempDir     string
		cacheDBFile string

		cacheManager freezer.CacheManager
	)

	it.Before(func() {
		var err error
		tempDir, err = ioutil.TempDir("", "cache")
		Expect(err).ToNot(HaveOccurred())

		cacheDBFile = filepath.Join(tempDir, "buildpacks-cache.db")

		cacheManager = freezer.NewCacheManager(tempDir)
		Expect(cacheManager.CacheDir).To(Equal(tempDir))
	})

	it.After(func() {
		Expect(os.RemoveAll(tempDir)).To(Succeed())
	})

	context("Load", func() {
		context("when Load is called on the cache manager and there is a buildpacks-cache.db file present", func() {
			var inputMap freezer.CacheDB

			it.Before(func() {
				inputMap = freezer.CacheDB{"buildpack": freezer.CacheEntry{Version: "1.2.3", URI: "some-uri"}}

				b := bytes.NewBuffer(nil)
				err := gob.NewEncoder(b).Encode(&inputMap)
				Expect(err).ToNot(HaveOccurred())

				Expect(ioutil.WriteFile(cacheDBFile, b.Bytes(), os.ModePerm))
			})

			it("returns the cache map stored in the buildpacks-cache.db folder", func() {
				err := cacheManager.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(cacheManager.Cache).To(Equal(inputMap))
			})
		})

		context("when Load is called on the cache manager and there is no buildpacks-cache.db file present", func() {
			it("returns an empty cache map", func() {
				err := cacheManager.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(cacheManager.Cache).To(BeNil())
			})
		})

		context("failure cases", func() {
			context("unable to open the buildpack-cache.db", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(cacheDBFile, []byte{}, 0000))
				})
				it("returns an error", func() {
					err := cacheManager.Load()
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})

			context("unable to open the buildpack-cache.db", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(cacheDBFile, []byte(`%%%`), os.ModePerm))
				})
				it("returns an error", func() {
					err := cacheManager.Load()
					Expect(err).To(MatchError(ContainSubstring("unexpected EOF")))
				})
			})
		})
	})

	context("Save", func() {
		context("when Save is called on the cache manager", func() {
			it.Before(func() {
				cacheManager.Cache = freezer.CacheDB{"some-buildpack": freezer.CacheEntry{Version: "1.2.3", URI: "some-uri"}}
			})

			it("saves the cache map given", func() {
				err := cacheManager.Save()
				Expect(err).ToNot(HaveOccurred())

				var cacheCheck freezer.CacheDB
				file, err := os.Open(cacheDBFile)
				Expect(err).ToNot(HaveOccurred())

				err = gob.NewDecoder(file).Decode(&cacheCheck)
				Expect(err).ToNot(HaveOccurred())

				Expect(cacheCheck).To(Equal(cacheManager.Cache))
			})
		})

		context("failure cases", func() {
			context("the db file cannot be created or opened", func() {
				it.Before(func() {
					Expect(os.Chmod(tempDir, 0000)).To(Succeed())
					cacheManager.Cache = freezer.CacheDB{"some-buildpack": freezer.CacheEntry{Version: "1.2.3", URI: "some-uri"}}
				})

				it.After(func() {
					Expect(os.Chmod(tempDir, os.ModePerm)).To(Succeed())
				})

				it("returns an error", func() {
					err := cacheManager.Save()
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
