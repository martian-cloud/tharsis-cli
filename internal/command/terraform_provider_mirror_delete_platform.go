package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type terraformProviderMirrorDeletePlatformCommand struct {
	*BaseCommand
}

var _ Command = (*terraformProviderMirrorDeletePlatformCommand)(nil)

// NewTerraformProviderMirrorDeletePlatformCommandFactory returns a terraformProviderMirrorDeletePlatformCommand struct.
func NewTerraformProviderMirrorDeletePlatformCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorDeletePlatformCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderMirrorDeletePlatformCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: platform mirror id")
	}

	return nil
}

func (c *terraformProviderMirrorDeletePlatformCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("terraform-provider-mirror delete-platform"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.DeleteTerraformProviderPlatformMirrorRequest{
		Id: c.arguments[0],
	}

	if _, err := c.grpcClient.TerraformProviderMirrorsClient.DeleteTerraformProviderPlatformMirror(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform provider platform mirror")
		return 1
	}

	c.UI.Successf("Terraform provider platform mirror deleted successfully!")
	return 0
}

func (*terraformProviderMirrorDeletePlatformCommand) Synopsis() string {
	return "Delete a terraform provider platform from mirror."
}

func (*terraformProviderMirrorDeletePlatformCommand) Description() string {
	return `
   Removes a platform binary from the provider mirror.
`
}

func (*terraformProviderMirrorDeletePlatformCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror delete-platform [options] <platform-mirror-id>"
}

func (*terraformProviderMirrorDeletePlatformCommand) Example() string {
	return `
tharsis terraform-provider-mirror delete-platform <platform-mirror-id>
`
}
