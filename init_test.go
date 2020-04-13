package freezer_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestFreezer(t *testing.T) {
	suite := spec.New("freezer", spec.Report(report.Terminal{}))
	suite("CacheManager", testCacheManager)
	suite("FileSystem", testFileSystem)
	suite("LocalFetcher", testLocalFetcher)
	suite("RandomName", testRandomName)
	suite("RemoteFetcher", testRemoteFetcher)
	suite.Run(t)
}
