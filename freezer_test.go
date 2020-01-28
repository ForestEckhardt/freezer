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
				b := bytes.NewBuffer(nil)
				encoder := gob.NewEncoder(b)

				inputMap = freezer.CacheDB{"buildpack": freezer.CacheEntry{Version: "1.2.3", URI: "some-uri"}}

				err := encoder.Encode(&inputMap)
				Expect(err).ToNot(HaveOccurred())

				Expect(ioutil.WriteFile(cacheDBFile, b.Bytes(), os.ModePerm))
			})

			it("returns the cache map stored in the buildpacks-cache.db folder", func() {
				cacheMap, err := cacheManager.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(cacheMap).To(Equal(inputMap))
			})
		})

		context("when Load is called on the cache manager and there is no buildpacks-cache.db file present", func() {
			it("returns an empty cache map", func() {
				cacheMap, err := cacheManager.Load()
				Expect(err).ToNot(HaveOccurred())

				Expect(cacheMap).To(Equal(freezer.CacheDB{}))
			})
		})

		context("failure cases", func() {
			context("unable to open the buildpack-cache.db", func() {
				it.Before(func() {
					Expect(ioutil.WriteFile(cacheDBFile, []byte{}, 0000))
				})
				it("returns an error", func() {
					_, err := cacheManager.Load()
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
}
