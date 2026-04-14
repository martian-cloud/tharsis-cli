package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderDeleteCommand struct {
	*BaseCommand
}

var _ Command = (*terraformProviderDeleteCommand)(nil)

func (c *terraformProviderDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewTerraformProviderDeleteCommandFactory returns a terraformProviderDeleteCommand struct.
func NewTerraformProviderDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider delete"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithWarningPrompt("This will permanently delete the Terraform provider and all its versions."),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.TerraformProvidersClient.DeleteTerraformProvider(c.Context, &pb.DeleteTerraformProviderRequest{
		Id: c.arguments[0],
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform provider")
		return 1
	}

	c.UI.Successf("Terraform provider deleted successfully!")
	return 0
}

func (*terraformProviderDeleteCommand) Synopsis() string {
	return "Delete a terraform provider."
}

func (*terraformProviderDeleteCommand) Usage() string {
	return "tharsis [global options] terraform-provider delete [options] <id>"
}

func (*terraformProviderDeleteCommand) Description() string {
	return `
   Permanently removes a Terraform provider and all
   its versions. This operation is irreversible.
`
}

func (*terraformProviderDeleteCommand) Example() string {
	return `
tharsis terraform-provider delete trn:terraform_provider:<group_path>/<provider_name>
`
}

func (c *terraformProviderDeleteCommand) Flags() *flag.Set {
	return flag.NewSet("Command options")
}
