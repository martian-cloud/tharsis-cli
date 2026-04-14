package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderGetPlatformCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*terraformProviderGetPlatformCommand)(nil)

// NewTerraformProviderGetPlatformCommandFactory returns a terraformProviderGetPlatformCommand struct.
func NewTerraformProviderGetPlatformCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderGetPlatformCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderGetPlatformCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

func (c *terraformProviderGetPlatformCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider get-platform"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.TerraformProvidersClient.GetTerraformProviderPlatformByID(c.Context, &pb.GetTerraformProviderPlatformByIDRequest{
		Id: c.arguments[0],
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get terraform provider platform")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*terraformProviderGetPlatformCommand) Synopsis() string {
	return "Get a terraform provider platform by ID or TRN."
}

func (*terraformProviderGetPlatformCommand) Description() string {
	return `
   Retrieves details about a Terraform provider
   platform including its OS, architecture, and
   binary upload status.
`
}

func (*terraformProviderGetPlatformCommand) Usage() string {
	return "tharsis [global options] terraform-provider get-platform [options] <id>"
}

func (*terraformProviderGetPlatformCommand) Example() string {
	return `
tharsis terraform-provider get-platform <provider-platform-id>
`
}

func (c *terraformProviderGetPlatformCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
