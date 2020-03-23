package freezer

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/packager/cnbpackager"
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
	args := []string{
		"--buildpack", filepath.Join(buildpackDir, "buildpack.toml"),
		"--output", output,
		"--version", version,
	}

	if cached {
		_, err := os.Stat(filepath.Join(buildpackDir, ".packit"))
		if err != nil {
			// Run packager this stop gap until either more things are packit compliant or
			// the format for cached buildpacks change
			usr, err := user.Current()
			if err != nil {
				return err
			}

			globalCacheDir := filepath.Join(usr.HomeDir, cnbpackager.DefaultCacheBase)

			packager, err := cnbpackager.New(buildpackDir, output, version, globalCacheDir)
			if err != nil {
				return err
			}

			err = packager.Create(cached)
			if err != nil {
				return err
			}

			// Return out of if statement so that we keep Jam as the default and only
			// add arguments
			return packager.Archive()
		} else {
			args = append(args, "--offline")
		}
	}

	return p.jam.Execute(args)

}
