// Inspired by github.com/cloudfoundry/occam/random_name.go

package freezer_test

import (
	"testing"

	"github.com/ForestEckhardt/freezer"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testRandomName(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		nameGenerator freezer.NameGenerator
	)

	it.Before(func() {
		nameGenerator = freezer.NewNameGenerator()
	})

	it("generates a random name with given prefix", func() {
		name, err := nameGenerator.RandomName("buildpack")
		Expect(err).NotTo(HaveOccurred())
		Expect(name).To(MatchRegexp(`^buildpack\-[0123456789abcdefghjkmnpqrstvwxyz]{26}$`))
	})
}
