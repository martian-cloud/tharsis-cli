package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderDeletePlatformCommand struct {
	*BaseCommand

	version *int64
}

var _ Command = (*terraformProviderDeletePlatformCommand)(nil)

// NewTerraformProviderDeletePlatformCommandFactory returns a terraformProviderDeletePlatformCommand struct.
func NewTerraformProviderDeletePlatformCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderDeletePlatformCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderDeletePlatformCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

func (c *terraformProviderDeletePlatformCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider delete-platform"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.TerraformProvidersClient.DeleteTerraformProviderPlatform(c.Context, &pb.DeleteTerraformProviderPlatformRequest{
		Id:      c.arguments[0],
		Version: c.version,
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform provider platform")
		return 1
	}

	c.UI.Successf("Terraform provider platform deleted successfully!")
	return 0
}

func (*terraformProviderDeletePlatformCommand) Synopsis() string {
	return "Delete a Terraform provider platform."
}

func (*terraformProviderDeletePlatformCommand) Usage() string {
	return "tharsis [global options] terraform-provider delete-platform [options] <id>"
}

func (*terraformProviderDeletePlatformCommand) Description() string {
	return `
   Permanently removes a Terraform provider platform
   binary. This operation is irreversible.
`
}

func (*terraformProviderDeletePlatformCommand) Example() string {
	return `
tharsis terraform-provider delete-platform <id>
`
}

func (c *terraformProviderDeletePlatformCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)

	return f
}
