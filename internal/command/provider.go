package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// providerCommand is the top-level structure for the provider command.
type providerCommand struct {
	meta *Metadata
}

// NewProviderCommandFactory returns a providerCommand struct.
func NewProviderCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return providerCommand{
			meta: meta,
		}, nil
	}
}

func (pc providerCommand) Run(args []string) int {
	pc.meta.Logger.Debugf("Starting the 'provider' command with %d arguments:", len(args))
	for ix, arg := range args {
		pc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Show the help text.
	pc.meta.UI.Output(pc.HelpProvider(true))
	return 1
}

func (pc providerCommand) Synopsis() string {
	return "Do operations on a terraform provider."
}

func (pc providerCommand) Help() string {
	return pc.HelpProvider(false)
}

// HelpProvider produces the help string for the 'provider' command.
func (pc providerCommand) HelpProvider(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] provider ...

   The provider commands do operations on a provider.
`, pc.meta.BinaryName)
	sc := `

Subcommands:
    create            Create a new provider.
    upload-version    Upload a new provider version to the provider registry.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}

// The End.
