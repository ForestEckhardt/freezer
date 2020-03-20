package commands_test

import (
	"testing"

	"github.com/ForestEckhardt/freezer/commands"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testStock(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		command commands.Stock
	)

	it.Before(func() {
		command = commands.NewStock()
	})

	context("Execute", func() {
		it("runs the fetcher", func() {
			err := command.Execute([]string{
				"--org", "some-org",
				"--repo", "some-repo",
			})

			Expect(err).NotTo(HaveOccurred())
		})
	})

	context("failure cases", func() {
		context("when given an unknown flag", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{"--unknown"})
				Expect(err).To(MatchError(ContainSubstring("flag provided but not defined: -unknown")))
			})
		})

		context("when the --org flag is empty", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{
					"--repo", "some-repo",
				})
				Expect(err).To(MatchError("missing required flag --org"))
			})
		})

		context("when the --repo flag is empty", func() {
			it("prints an error message", func() {
				err := command.Execute([]string{
					"--org", "some-org",
				})
				Expect(err).To(MatchError("missing required flag --repo"))
			})
		})

	})
}
