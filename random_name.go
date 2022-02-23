// Inspired by github.com/paketo-buildpacks/occam/random_name.go

package freezer

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid"
)

type NameGenerator struct{}

func NewNameGenerator() NameGenerator {
	return NameGenerator{}
}

func (n NameGenerator) RandomName(name string) (string, error) {
	now := time.Now()
	timestamp := ulid.Timestamp(now)
	entropy := ulid.Monotonic(rand.Reader, 0)

	guid, err := ulid.New(timestamp, entropy)
	if err != nil {
		return "", err
	}

	return strings.ToLower(fmt.Sprintf("%s-%s", name, guid)), nil
}
