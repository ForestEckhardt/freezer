package freezer_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestFreezer(t *testing.T) {
	suite := spec.New("vacation", spec.Report(report.Terminal{}))
	suite("CacheManager", testCacheManager)
	suite.Run(t)
}
