package freezer_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestFreezer(t *testing.T) {
	suite := spec.New("freezer", spec.Report(report.Terminal{}))
	suite("CacheManager", testCacheManager)
	suite("RemoteBuildpack", testRemoteBuildpack)
	suite.Run(t)
}
