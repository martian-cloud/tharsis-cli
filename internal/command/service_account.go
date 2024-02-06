package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// serviceAccountCommand is the top-level structure for the service account command.
// It exists only for the purpose of filling a hole in the help output.
type serviceAccountCommand struct {
	meta *Metadata
}

// NewServiceAccountCommandFactory returns a serviceAccountCommand struct.
func NewServiceAccountCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return serviceAccountCommand{
			meta: meta,
		}, nil
	}
}

func (sc serviceAccountCommand) Run(args []string) int {
	sc.meta.Logger.Debugf("Starting the 'service-account' command with %d arguments:", len(args))
	for ix, arg := range args {
		sc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	sc.meta.UI.Output(sc.HelpServiceAccount(true))
	return 1
}

func (sc serviceAccountCommand) Synopsis() string {
	return "Create an authentication token for a service account."
}

func (sc serviceAccountCommand) Help() string {
	return sc.HelpServiceAccount(false)
}

// HelpServiceAccount produces the help string for the 'service-account' command.
func (sc serviceAccountCommand) HelpServiceAccount(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] service-account create-token
`, sc.meta.BinaryName)

	subs := `

Subcommands:
    create-token    Create a token for a service account.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + subs
	}

	return usage
}
