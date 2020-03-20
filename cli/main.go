package main

import (
	"fmt"
	"os"

	"github.com/ForestEckhardt/freezer"
	"github.com/ForestEckhardt/freezer/commands"
	"github.com/cloudfoundry/packit/cargo"
)

func main() {
	if len(os.Args) < 2 {
		fail("missing command")
	}

	switch os.Args[1] {
	case "stock":
		transport := cargo.NewTransport()
		packager := freezer.NewPackingTools()
		command := commands.NewStock(transport, packager)

		if err := command.Execute(os.Args[2:]); err != nil {
			fail("failed to execute stocker command: %s", err)
		}

	default:
		fail("unknown command: %q", os.Args[1])
	}
}

func fail(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format, v...)
	os.Exit(1)
}
