package github_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestGithub(t *testing.T) {
	suite := spec.New("github", spec.Report(report.Terminal{}))
	suite("ReleaseService", testReleaseService)

	suite.Before(func(t *testing.T) {
		RegisterTestingT(t)
	})

	suite.Run(t)
}

func Fail(message string) {
	panic(message)
}
