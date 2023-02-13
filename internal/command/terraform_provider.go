package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

// terraformProviderCommand is the top-level structure for the terraform-provider command.
type terraformProviderCommand struct {
	meta *Metadata
}

// NewTerraformProviderCommandFactory returns a (Terraform provider) Command struct.
func NewTerraformProviderCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderCommand{
			meta: meta,
		}, nil
	}
}

func (pc terraformProviderCommand) Run(args []string) int {
	pc.meta.Logger.Debugf("Starting the 'terraform-provider' command with %d arguments:", len(args))
	for ix, arg := range args {
		pc.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	// Show the help text.
	pc.meta.UI.Output(pc.HelpTerraformProvider(true))
	return 1
}

func (pc terraformProviderCommand) Synopsis() string {
	return "Do operations on a terraform provider."
}

func (pc terraformProviderCommand) Help() string {
	return pc.HelpTerraformProvider(false)
}

// HelpTerraformProvider produces the help string for the 'terraform-provider' command.
func (pc terraformProviderCommand) HelpTerraformProvider(subCommands bool) string {
	usage := fmt.Sprintf(`
Usage: %s [global options] terraform-provider ...

   The terraform-provider commands do operations on a Terraform provider.
`, pc.meta.BinaryName)
	sc := `

Subcommands:
    create            Create a new Terraform provider.
    upload-version    Upload a new Terraform provider version to the provider registry.`

	// Avoid duplicate subcommands when -h is used.
	if subCommands {
		return usage + sc
	}

	return usage
}

// The End.
