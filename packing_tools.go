package freezer

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/packit/cargo"
	"github.com/cloudfoundry/packit/cargo/jam/commands"
	"github.com/cloudfoundry/packit/pexec"
	"github.com/cloudfoundry/packit/scribe"
)

type PackingTools struct {
	jam commands.Pack
}

func NewPackingTools() PackingTools {
	logger := scribe.NewLogger(os.Stdout)
	bash := pexec.NewExecutable("bash")

	transport := cargo.NewTransport()
	directoryDuplicator := cargo.NewDirectoryDuplicator()
	buildpackParser := cargo.NewBuildpackParser()
	fileBundler := cargo.NewFileBundler()
	tarBuilder := cargo.NewTarBuilder(logger)
	prePackager := cargo.NewPrePackager(bash, logger, scribe.NewWriter(os.Stdout, scribe.WithIndent(2)))
	dependencyCacher := cargo.NewDependencyCacher(transport, logger)

	return PackingTools{
		jam: commands.NewPack(directoryDuplicator, buildpackParser, prePackager, dependencyCacher, fileBundler, tarBuilder, os.Stdout),
	}
}

func (p PackingTools) Execute(buildpackDir, output, version string, cached bool) error {
	_, err := os.Stat(filepath.Join(buildpackDir, ".packit"))
	if err != nil {
		return err
	}

	args := []string{
		"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
		"--output", output,
		"--version", version,
	}

	if cached {
		args = append(args, "--offline")
	}

	return p.jam.Execute(args)
}
