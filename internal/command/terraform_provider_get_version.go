package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderGetVersionCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*terraformProviderGetVersionCommand)(nil)

// NewTerraformProviderGetVersionCommandFactory returns a terraformProviderGetVersionCommand struct.
func NewTerraformProviderGetVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderGetVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderGetVersionCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

func (c *terraformProviderGetVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider get-version"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.TerraformProvidersClient.GetTerraformProviderVersionByID(c.Context, &pb.GetTerraformProviderVersionByIDRequest{
		Id: c.arguments[0],
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get terraform provider version")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*terraformProviderGetVersionCommand) Synopsis() string {
	return "Get a terraform provider version by ID or TRN."
}

func (*terraformProviderGetVersionCommand) Description() string {
	return `
   Retrieves details about a Terraform provider
   version including its semantic version and upload
   status.
`
}

func (*terraformProviderGetVersionCommand) Usage() string {
	return "tharsis [global options] terraform-provider get-version [options] <id>"
}

func (*terraformProviderGetVersionCommand) Example() string {
	return `
tharsis terraform-provider get-version <provider-version-id>
`
}

func (c *terraformProviderGetVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
