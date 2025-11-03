package command

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/optparser"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/output"
	tharsis "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

type terraformProviderMirrorDeleteVersionCommand struct {
	meta *Metadata
}

// NewTerraformProviderMirrorDeleteVersionCommandFactory returns a terraformProviderMirrorDeleteVersionCommand struct.
func NewTerraformProviderMirrorDeleteVersionCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderMirrorDeleteVersionCommand{
			meta: meta,
		}, nil
	}
}

func (c terraformProviderMirrorDeleteVersionCommand) Run(args []string) int {
	c.meta.Logger.Debugf("Starting the 'terraform-provider-mirror delete-version' command with %d arguments:", len(args))
	for ix, arg := range args {
		c.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := c.meta.GetSDKClient()
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return c.doTerraformProviderMirrorDeleteVersion(ctx, client, args)
}

func (c terraformProviderMirrorDeleteVersionCommand) doTerraformProviderMirrorDeleteVersion(ctx context.Context, client *tharsis.Client, opts []string) int {
	c.meta.Logger.Debugf("will do terraform-provider delete-version, %d opts", len(opts))

	defs := c.defs()
	cmdOpts, cmdArgs, err := optparser.ParseCommandOptions(c.meta.BinaryName+" terraform-provider-mirror delete-version", defs, opts)
	if err != nil {
		c.meta.Logger.Error(output.FormatError("failed to parse terraform-provider-mirror delete-version options", err))
		return cli.RunResultHelp
	}
	if len(cmdArgs) < 1 {
		c.meta.Logger.Error(output.FormatError("missing terraform-provider-mirror delete-version id", nil))
		return cli.RunResultHelp
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive terraform-provider-mirror delete-version arguments: %s", cmdArgs)
		c.meta.Logger.Error(output.FormatError(msg, nil))
		return cli.RunResultHelp
	}

	force, err := getBoolOptionValue("force", "false", cmdOpts)
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to parse boolean value", err))
		return 1
	}

	toDelete := &sdktypes.DeleteTerraformProviderVersionMirrorInput{
		ID:    cmdArgs[0],
		Force: force,
	}

	c.meta.Logger.Debugf("terraform-provider-mirror delete-version input: %#v", toDelete)

	err = client.TerraformProviderVersionMirror.DeleteProviderVersionMirror(ctx, toDelete)
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to delete terraform provider version mirror", err))
		return 1
	}

	c.meta.UI.Output("Terraform provider version successfully deleted from mirror.")
	return 0
}

func (terraformProviderMirrorDeleteVersionCommand) defs() optparser.OptionDefinitions {
	return optparser.OptionDefinitions{
		"force": {
			Arguments: []string{},
			Synopsis:  "Force the deletion of a provider version from mirror.",
		},
	}
}

func (terraformProviderMirrorDeleteVersionCommand) Synopsis() string {
	return "Delete a Terraform provider version from mirror."
}

func (c terraformProviderMirrorDeleteVersionCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider-mirror delete-version <id>

   The terraform-provider-mirror delete-version command deletes
   a terraform provider version and any associated platform
   binaries from a group's mirror. The --force option must
   be used when deleting a provider version which actively
   hosts platform binaries.

%s

`, c.meta.BinaryName, buildHelpText(c.defs()))
}
