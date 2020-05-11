package freezer_test

import (
	"testing"

	"github.com/ForestEckhardt/freezer"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testPackingTools(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		packingTools freezer.PackingTools
	)

	it.Before(func() {
		packingTools = freezer.NewPackingTools()
	})

	context("Execute", func() {
		context("failure cases", func() {
			context("when the buildpack is not a packit buildpack", func() {
				it("returns an error", func() {
					err := packingTools.Execute("fake-dir/", "", "", false)
					Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				})
			})
		})
	})
}
