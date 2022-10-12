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

	// Show the help text.
	gc.meta.UI.Output(gc.HelpGroup(true))
	return 1
}

func (gc groupCommand) Synopsis() string {
	return "Do operations on groups."
}

func (gc groupCommand) Help() string {
	return gc.HelpGroup(false)
}

// HelpGroup produces the help string for the 'group' command.
func (gc groupCommand) HelpGroup(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] group ...

   The group commands do operations on groups. Subcommands
   allow creating / updating / deleting groups, setting
   Terraform / environment variables, listing all groups and
   more.
`, gc.meta.BinaryName)
	sc := `

Subcommands:
    create                  Create a new group.
    delete                  Delete a group.
    get                     Get a single group.
    list                    List groups.
    set-environment-vars    Set environment variables for a group.
    set-terraform-vars      Set terraform variables for a group.
    update                  Update a group.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}

// The End.
