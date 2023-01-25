package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// moduleCommand is the top-level structure for the module command.
type moduleCommand struct {
	meta *Metadata
}

// NewModuleCommandFactory returns a moduleCommand struct.
func NewModuleCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return moduleCommand{
			meta: meta,
		}, nil
	}
}

func (mc moduleCommand) Run(args []string) int {
	// Show the help text.
	mc.meta.UI.Output(mc.HelpModule(true))
	return 1
}

func (mc moduleCommand) Synopsis() string {
	return "Do operations on a terraform module."
}

func (mc moduleCommand) Help() string {
	return mc.HelpModule(false)
}

// HelpModule produces the help string for the 'module' command.
func (mc moduleCommand) HelpModule(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] module ...

   The module commands do operations on a Terraform module.
`, mc.meta.BinaryName)
	sc := `

Subcommands:
    create            Create a new module.
    upload-version    Upload a new module version to the module registry.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}

// The End.
