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

				Expect(cacheManager.Cache).To(BeNil())
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
}
