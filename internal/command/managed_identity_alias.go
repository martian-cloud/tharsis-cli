package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// managedIdentityAliasCommand is the top-level structure for the managed-identity-alias command.
type managedIdentityAliasCommand struct {
	meta *Metadata
}

// NewManagedIdentityAliasCommandFactory returns a managedIdentityAliasCommand struct.
func NewManagedIdentityAliasCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return managedIdentityAliasCommand{
			meta: meta,
		}, nil
	}
}

func (m managedIdentityAliasCommand) Run(args []string) int {
	m.meta.Logger.Debugf("Starting the 'managed-identity-alias' command with %d arguments:", len(args))
	for ix, arg := range args {
		m.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Show the help text.
	m.meta.UI.Output(m.HelpManagedIdentityAlias(true))
	return 1
}

func (m managedIdentityAliasCommand) Synopsis() string {
	return "Do operations on a managed identity alias."
}

func (m managedIdentityAliasCommand) Help() string {
	return m.HelpManagedIdentityAlias(false)
}

// HelpManagedIdentityAlias produces the help string for the 'managed-identity-alias' command.
func (m managedIdentityAliasCommand) HelpManagedIdentityAlias(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] managed-identity-alias ...

   The managed-identity-alias commands do operations on a managed identity alias.
`, m.meta.BinaryName)
	sc := `

Subcommands:
    create                       Create a new managed identity alias.
    delete                       Delete a managed identity alias.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}
