package freezer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/paketo-buildpacks/packit/v2/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(pexec.Execution) error
}

type PackingTools struct {
	jam        Executable
	pack       Executable
	tempOutput func(dir string, pattern string) (string, error)
}

func NewPackingTools() PackingTools {
	return PackingTools{
		jam:        pexec.NewExecutable("jam"),
		pack:       pexec.NewExecutable("pack"),
		tempOutput: os.MkdirTemp,
	}
}

func (p PackingTools) WithExecutable(executable Executable) PackingTools {
	p.jam = executable
	return p
}

func (p PackingTools) WithPack(pack Executable) PackingTools {
	p.pack = pack
	return p
}

func (p PackingTools) WithTempOutput(tempOutput func(string, string) (string, error)) PackingTools {
	p.tempOutput = tempOutput
	return p
}

func (p PackingTools) Execute(buildpackDir, output, version string, cached bool) error {
	jamOutput, err := p.tempOutput("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(jamOutput)

	args := []string{
		"pack",
		"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
		"--output", filepath.Join(jamOutput, fmt.Sprintf("%s.tgz", version)),
		"--version", version,
	}

	if cached {
		args = append(args, "--offline")
	}

	err = p.jam.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return err
	}

	args = []string{
		"buildpack", "package",
		output,
		"--path", filepath.Join(jamOutput, fmt.Sprintf("%s.tgz", version)),
		"--format", "file",
		"--target", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	return p.pack.Execute(pexec.Execution{
		Args:   args,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
