package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type terraformProviderMirrorDeleteVersionCommand struct {
	*BaseCommand

	force bool
}

// NewTerraformProviderMirrorDeleteVersionCommandFactory returns a terraformProviderMirrorDeleteVersionCommand struct.
func NewTerraformProviderMirrorDeleteVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderMirrorDeleteVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderMirrorDeleteVersionCommand) validate() error {
	const message = "version-mirror-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *terraformProviderMirrorDeleteVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider-mirror delete-version"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithForcePrompt("Are you sure you want to delete this provider version mirror?"),
	); code != 0 {
		return code
	}

	input := &pb.DeleteTerraformProviderVersionMirrorRequest{
		Id:    c.arguments[0],
		Force: c.force,
	}

	c.Logger.Debug("terraform-provider-mirror delete-version input", "input", input)

	if _, err := c.client.TerraformProviderMirrorsClient.DeleteTerraformProviderVersionMirror(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform provider version mirror")
		return 1
	}

	c.UI.Successf("Terraform provider version mirror %s deleted successfully", c.arguments[0])
	return 0
}

func (*terraformProviderMirrorDeleteVersionCommand) Synopsis() string {
	return "Delete a terraform provider version from mirror."
}

func (*terraformProviderMirrorDeleteVersionCommand) Description() string {
	return `
   The terraform-provider-mirror delete-version command deletes a terraform provider
   version and any associated platform binaries from a group's mirror.
`
}

func (*terraformProviderMirrorDeleteVersionCommand) Usage() string {
	return "tharsis [global options] terraform-provider-mirror delete-version [options] <version-mirror-id>"
}

func (*terraformProviderMirrorDeleteVersionCommand) Example() string {
	return `
tharsis terraform-provider-mirror delete-version --force <version-mirror-id>
`
}

func (c *terraformProviderMirrorDeleteVersionCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.force,
		"force",
		false,
		"Skip confirmation prompt.",
	)

	return f
}
