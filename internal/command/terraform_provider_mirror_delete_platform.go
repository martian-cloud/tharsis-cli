package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type terraformProviderMirrorDeletePlatformCommand struct {
	*BaseCommand
}

// NewTerraformProviderMirrorDeletePlatformCommandFactory returns a terraformProviderMirrorDeletePlatformCommand struct.
func NewTerraformProviderMirrorDeletePlatformCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorDeletePlatformCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderMirrorDeletePlatformCommand) validate() error {
	const message = "platform-mirror-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *terraformProviderMirrorDeletePlatformCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider-mirror delete-platform"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.DeleteTerraformProviderPlatformMirrorRequest{
		Id: c.arguments[0],
	}

	c.Logger.Debug("terraform-provider-mirror delete-platform input", "input", input)

	if _, err := c.client.TerraformProviderMirrorsClient.DeleteTerraformProviderPlatformMirror(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform provider platform mirror")
		return 1
	}

	c.UI.Successf("Terraform provider platform mirror %s deleted successfully", c.arguments[0])
	return 0
}

func (*terraformProviderMirrorDeletePlatformCommand) Synopsis() string {
	return "Delete a terraform provider platform from mirror."
}

func (*terraformProviderMirrorDeletePlatformCommand) Description() string {
	return `
   The terraform-provider-mirror delete-platform command deletes a terraform provider
   platform from a group's mirror. The package will no longer be available for the
   associated provider's version and platform.
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

func (c *terraformProviderMirrorDeletePlatformCommand) Flags() *flag.FlagSet {
	return nil
}
