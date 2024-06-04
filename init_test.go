package freezer_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var gitToken string

func TestFreezer(t *testing.T) {
	suite := spec.New("freezer", spec.Report(report.Terminal{}))
	suite("CacheManager", testCacheManager)
	suite("LocalFetcher", testLocalFetcher)
	suite("PackingTools", testPackingTools)
	suite("RandomName", testRandomName)
	suite("RemoteFetcher", testRemoteFetcher)
	suite.Run(t)
}
