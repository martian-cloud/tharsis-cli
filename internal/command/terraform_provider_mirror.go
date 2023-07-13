package command

import (
	"fmt"

	"github.com/mitchellh/cli"
)

type terraformProviderMirrorCommand struct {
	meta *Metadata
}

// NewTerraformProviderMirrorCommandFactory returns a terraformProviderMirrorCommand struct.
func NewTerraformProviderMirrorCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderMirrorCommand{
			meta: meta,
		}, nil
	}
}

func (c terraformProviderMirrorCommand) Run(args []string) int {
	c.meta.Logger.Debugf("Starting the 'terraform-provider-mirror' command with %d arguments:", len(args))
	for ix, arg := range args {
		c.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	return cli.RunResultHelp
}

func (terraformProviderMirrorCommand) Synopsis() string {
	return "Mirror Terraform providers from any Terraform registry."
}

func (c terraformProviderMirrorCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider-mirror ...

   The terraform-provider-mirror command allows interacting
   with the Tharsis Terraform provider mirror which supports
   Terraform's Provider Network Mirror Protocol. The Tharsis
   provider mirror hosts a set of Terraform providers for
   use within a group's hierarchy and gives root group
   owners full control on which providers, platform packages
   and registries are available via their mirror.
   Subcommands help upload provider packages from any
   Terraform Provider Registry to the mirror. Uploaded
   packages will be verified for legitimacy against the
   provider's Terraform Registry API.
`, c.meta.BinaryName)
}
