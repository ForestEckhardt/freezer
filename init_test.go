package freezer_test

import (
	"os"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var gitToken string

func TestFreezer(t *testing.T) {
	var ok bool
	gitToken, ok = os.LookupEnv("GIT_TOKEN")
	if !ok {
		t.Fatal("$GIT_TOKEN environment variable must be set")
	}

	suite := spec.New("freezer", spec.Report(report.Terminal{}))
	suite("CacheManager", testCacheManager)
	suite("FileSystem", testFileSystem)
	suite("LocalFetcher", testLocalFetcher)
	suite("PackingTools", testPackingTools)
	suite("RandomName", testRandomName)
	suite("RemoteFetcher", testRemoteFetcher)
	suite.Run(t)
}
