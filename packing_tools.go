package freezer

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/pexec"
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
	args := []string{
		"pack",
		"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
		"--output", output,
		"--version", version,
	}

	if cached {
		args = append(args, "--offline")
	}

	return p.jam.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
