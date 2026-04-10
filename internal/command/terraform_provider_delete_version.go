package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderDeleteVersionCommand struct {
	*BaseCommand

	version *int64
}

var _ Command = (*terraformProviderDeleteVersionCommand)(nil)

// NewTerraformProviderDeleteVersionCommandFactory returns a terraformProviderDeleteVersionCommand struct.
func NewTerraformProviderDeleteVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderDeleteVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderDeleteVersionCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

func (c *terraformProviderDeleteVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider delete-version"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.TerraformProvidersClient.DeleteTerraformProviderVersion(c.Context, &pb.DeleteTerraformProviderVersionRequest{
		Id:      c.arguments[0],
		Version: c.version,
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete terraform provider version")
		return 1
	}

	c.UI.Successf("Terraform provider version deleted successfully!")
	return 0
}

func (*terraformProviderDeleteVersionCommand) Synopsis() string {
	return "Delete a terraform provider version."
}

func (*terraformProviderDeleteVersionCommand) Description() string {
	return `
   Permanently removes a Terraform provider version
   and all its platforms. This operation is
   irreversible.
`
}

func (*terraformProviderDeleteVersionCommand) Usage() string {
	return "tharsis [global options] terraform-provider delete-version [options] <id>"
}

func (*terraformProviderDeleteVersionCommand) Example() string {
	return `
tharsis terraform-provider delete-version <provider-version-id>
`
}

func (c *terraformProviderDeleteVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)

	return f
}
