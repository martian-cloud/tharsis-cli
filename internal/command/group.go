package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// groupCommand is the top-level structure for the group command.
type groupCommand struct {
	meta *Metadata
}

// NewGroupCommandFactory returns a groupCommand struct.
func NewGroupCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return groupCommand{
			meta: meta,
		}, nil
	}
}

func (gc groupCommand) Run(args []string) int {
	gc.meta.Logger.Debugf("Starting the 'group' command with %d arguments:", len(args))
	for ix, arg := range args {
		gc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	return cli.RunResultHelp
}

func (gc groupCommand) Synopsis() string {
	return "Do operations on groups."
}

func (gc groupCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] group ...

   The group commands do operations on groups. Subcommands
   allow creating / updating / deleting groups, setting
   Terraform / environment variables, listing all groups and
   more.
`, gc.meta.BinaryName)
}
