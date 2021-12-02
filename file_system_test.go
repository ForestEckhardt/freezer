package freezer_test

import (
	"os"
	"testing"

	"github.com/ForestEckhardt/freezer"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testFileSystem(t *testing.T, context spec.G, it spec.S) {
	var Expect = NewWithT(t).Expect

	context("TempDir", func() {
		it("returns the value given from the TempDir function", func() {
			filename, err := os.MkdirTemp("", "tempDir")
			defer os.RemoveAll(filename)

			fileSystem := freezer.NewFileSystem(func(string, string) (string, error) {
				return filename, err
			})

			tName, tErr := fileSystem.TempDir("", "tempDir")
			Expect(tName).To(Equal(filename))
			Expect(tErr).To(BeNil())
		})
	})
}
