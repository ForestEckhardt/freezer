package freezer_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/paketo-buildpacks/occam/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPackingTools(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buildpackDir string

		executable   *fakes.Executable
		packingTools freezer.PackingTools
	)

	it.Before(func() {
		var err error
		buildpackDir, err = os.MkdirTemp("", "buildpack-dir")
		Expect(err).ToNot(HaveOccurred())

		executable = &fakes.Executable{}

		packingTools = freezer.NewPackingTools().WithExecutable(executable)

	})

	it.After(func() {
		Expect(os.RemoveAll(buildpackDir)).To(Succeed())
	})

	context("Execute", func() {
		it("creates a correct pexec.Execution", func() {
			err := packingTools.Execute(buildpackDir, "some-output", "some-version", false)
			Expect(err).NotTo(HaveOccurred())

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"pack",
				"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
				"--output", "some-output",
				"--version", "some-version",
			}))
		})

		context("when cache is set to true", func() {
			it("creates a correct pexec.Execution", func() {
				err := packingTools.Execute(buildpackDir, "some-output", "some-version", true)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"pack",
					"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
					"--output", "some-output",
					"--version", "some-version",
					"--offline",
				}))
			})
		})

		context("failure cases", func() {
			context("when the execution returns an error", func() {
				it.Before(func() {
					executable.ExecuteCall.Returns.Error = errors.New("some error")
				})
				it("returns an error", func() {
					err := packingTools.Execute(buildpackDir, "some-output", "some-version", true)
					Expect(err).To(MatchError("some error"))
				})
			})
		})
	})
}
