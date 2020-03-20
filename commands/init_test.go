package commands_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestFreezer(t *testing.T) {
	suite := spec.New("commands", spec.Report(report.Terminal{}))
	suite("Stock", testStock)
	suite.Run(t)
}
