package freezer_test

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/ForestEckhardt/freezer/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testPackingTools(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		executable   *fakes.Executable
		pack         *fakes.Executable
		tempOutput   func(string, string) (string, error)
		packingTools freezer.PackingTools
	)

	it.Before(func() {
		executable = &fakes.Executable{}
		pack = &fakes.Executable{}

		tempOutput = func(string, string) (string, error) {
			return "some-jam-output", nil
		}

		packingTools = freezer.NewPackingTools().WithExecutable(executable).WithPack(pack).WithTempOutput(tempOutput)

	})

	context("Execute", func() {
		it("creates a correct pexec.Execution", func() {
			err := packingTools.Execute("some-buildpack-dir", "some-output", "some-version", false)
			Expect(err).NotTo(HaveOccurred())

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"pack",
				"--buildpack", filepath.Join("some-buildpack-dir", "buildpack.toml"),
				"--output", filepath.Join("some-jam-output", "some-version.tgz"),
				"--version", "some-version",
			}))

			Expect(pack.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				"buildpack", "package",
				"some-output",
				"--path", filepath.Join("some-jam-output", "some-version.tgz"),
				"--format", "file",
				"--target", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			}))
		})

		context("when cache is set to true", func() {
			it("creates a correct pexec.Execution", func() {
				err := packingTools.Execute("some-buildpack-dir", "some-output", "some-version", true)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"pack",
					"--buildpack", filepath.Join("some-buildpack-dir", "buildpack.toml"),
					"--output", filepath.Join("some-jam-output", "some-version.tgz"),
					"--version", "some-version",
					"--offline",
				}))

				Expect(pack.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
					"buildpack", "package",
					"some-output",
					"--path", filepath.Join("some-jam-output", "some-version.tgz"),
					"--format", "file",
					"--target", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				}))
			})
		})

		context("failure cases", func() {
			context("when the tempDir creation fails returns an error", func() {
				it.Before(func() {
					tempOutput = func(string, string) (string, error) {
						return "", errors.New("some tempDir error")
					}

					packingTools = packingTools.WithTempOutput(tempOutput)
				})
				it("returns an error", func() {
					err := packingTools.Execute("some-buildpack-dir", "some-output", "some-version", true)
					Expect(err).To(MatchError("some tempDir error"))
				})
			})

			context("when the jam execution returns an error", func() {
				it.Before(func() {
					executable.ExecuteCall.Returns.Error = errors.New("some jam error")
				})
				it("returns an error", func() {
					err := packingTools.Execute("some-buildpack-dir", "some-output", "some-version", true)
					Expect(err).To(MatchError("some jam error"))
				})
			})

			context("when the pack execution returns an error", func() {
				it.Before(func() {
					pack.ExecuteCall.Returns.Error = errors.New("some pack error")
				})
				it("returns an error", func() {
					err := packingTools.Execute("some-buildpack-dir", "some-output", "some-version", true)
					Expect(err).To(MatchError("some pack error"))
				})
			})
		})
	})
}
