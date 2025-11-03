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

type terraformProviderMirrorDeletePlatformCommand struct {
	meta *Metadata
}

// NewTerraformProviderMirrorDeletePlatformCommandFactory returns a terraformProviderMirrorDeletePlatformCommand struct.
func NewTerraformProviderMirrorDeletePlatformCommandFactory(meta *Metadata) func() (cli.Command, error) {
	return func() (cli.Command, error) {
		return terraformProviderMirrorDeletePlatformCommand{
			meta: meta,
		}, nil
	}
}

func (c terraformProviderMirrorDeletePlatformCommand) Run(args []string) int {
	c.meta.Logger.Debugf("Starting the 'terraform-provider-mirror delete-platform' command with %d arguments:", len(args))
	for ix, arg := range args {
		c.meta.Logger.Debugf("    argument %d: %s", ix, arg)
	}

	client, err := c.meta.GetSDKClient()
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to get SDK client", err))
		return 1
	}

	ctx := context.Background()

	return c.doTerraformProviderMirrorDeletePlatform(ctx, client, args)
}

func (c terraformProviderMirrorDeletePlatformCommand) doTerraformProviderMirrorDeletePlatform(ctx context.Context, client *tharsis.Client, opts []string) int {
	c.meta.Logger.Debugf("will do terraform-provider delete-platform, %d opts", len(opts))

	_, cmdArgs, err := optparser.ParseCommandOptions(c.meta.BinaryName+" terraform-provider-mirror delete-platform", optparser.OptionDefinitions{}, opts)
	if err != nil {
		c.meta.Logger.Error(output.FormatError("failed to parse terraform-provider-mirror delete-platform options", err))
		return cli.RunResultHelp
	}
	if len(cmdArgs) < 1 {
		c.meta.Logger.Error(output.FormatError("missing terraform-provider-mirror delete-platform id", nil))
		return cli.RunResultHelp
	}
	if len(cmdArgs) > 1 {
		msg := fmt.Sprintf("excessive terraform-provider-mirror delete-platform arguments: %s", cmdArgs)
		c.meta.Logger.Error(output.FormatError(msg, nil))
		return cli.RunResultHelp
	}

	toDelete := &sdktypes.DeleteTerraformProviderPlatformMirrorInput{
		ID: cmdArgs[0],
	}

	c.meta.Logger.Debugf("terraform-provider-mirror delete-platform input: %#v", toDelete)

	err = client.TerraformProviderPlatformMirror.DeleteProviderPlatformMirror(ctx, toDelete)
	if err != nil {
		c.meta.UI.Error(output.FormatError("failed to delete terraform provider platform mirror", err))
		return 1
	}

	c.meta.UI.Output("Terraform provider platform successfully deleted from mirror.")
	return 0
}

func (terraformProviderMirrorDeletePlatformCommand) Synopsis() string {
	return "Delete a Terraform provider platform from mirror."
}

func (c terraformProviderMirrorDeletePlatformCommand) Help() string {
	return fmt.Sprintf(`
Usage: %s [global options] terraform-provider-mirror delete-platform <id>

   The terraform-provider-mirror delete-platform command deletes
   a Terraform provider platform from a group's mirror. If
   successful, the package will no longer be available for the
   associated provider's version and platform. Useful when a
   package may no longer be needed (testing) or it becomes
   corrupted.
`, c.meta.BinaryName)
}
