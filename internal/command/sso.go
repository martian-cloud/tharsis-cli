package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// ssoCommand is the top-level structure for the sso command.
// It exists only for the purpose of filling a hole in the help output.
type ssoCommand struct {
	meta *Metadata
}

// NewSSOCommandFactory returns a ssoCommand struct.
func NewSSOCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return ssoCommand{
			meta: meta,
		}, nil
	}
}

func (sc ssoCommand) Run(args []string) int {
	sc.meta.Logger.Debugf("Starting the 'sso' command with %d arguments:", len(args))
	for ix, arg := range args {
		sc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	sc.meta.UI.Output(sc.HelpSSO(true))
	return 1
}

func (sc ssoCommand) Synopsis() string {
	return "Log in to the OAuth2 provider and return an authentication token"
}

func (sc ssoCommand) Help() string {
	return sc.HelpSSO(false)
}

// HelpSSO produces the help string for the 'sso' command.
func (sc ssoCommand) HelpSSO(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] sso login
`, sc.meta.BinaryName)

	subs := `

Subcommands:
    login    Log in to the OAuth2 provider and return an authentication token`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + subs
	}

	return usage
}

// The End.
