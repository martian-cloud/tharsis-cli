package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// managedIdentityCommand is the top-level structure for the managed-identity command.
type managedIdentityCommand struct {
	meta *Metadata
}

// NewManagedIdentityCommandFactory returns a managedIdentityCommand struct.
func NewManagedIdentityCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Show the help text.
	m.meta.UI.Output(m.HelpManagedIdentity(true))
	return 1
}

func (m managedIdentityCommand) Synopsis() string {
	return "Do operations on a managed identity."
}

func (m managedIdentityCommand) Help() string {
	return m.HelpManagedIdentity(false)
}

// HelpManagedIdentity produces the help string for the 'managed-identity' command.
func (m managedIdentityCommand) HelpManagedIdentity(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] managed-identity ...

   The managed-identity commands do operations on a managed identity.
`, m.meta.BinaryName)
	sc := `

Subcommands:
    create                       Create a new managed identity.
    delete                       Delete a managed identity.
    get                          Get a single managed identity.
    update                       Update a managed identity.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}
