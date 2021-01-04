package freezer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

type PackingTools struct {
	jam Executable
}

func NewPackingTools() PackingTools {
	return PackingTools{
		jam: pexec.NewExecutable("jam"),
	}
}

func (p PackingTools) WithExecutable(executable Executable) PackingTools {
	p.jam = executable
	return p
}

func (p PackingTools) Execute(buildpackDir, output, version string, cached bool) error {
	_, err := os.Stat(filepath.Join(buildpackDir, ".packit"))
	if err != nil {
		return fmt.Errorf("unable to find .packit in buildpack directory: %w", err)
	}

	args := []string{
		"pack",
		"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
		"--output", output,
		"--version", version,
	}

	if cached {
		args = append(args, "--offline")
	}

	return p.jam.Execute(pexec.Execution{Args: args})
}
