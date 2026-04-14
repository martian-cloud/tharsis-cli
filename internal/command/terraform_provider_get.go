package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*terraformProviderGetCommand)(nil)

func (c *terraformProviderGetCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewTerraformProviderGetCommandFactory returns a terraformProviderGetCommand struct.
func NewTerraformProviderGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderGetCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.TerraformProvidersClient.GetTerraformProviderByID(c.Context, &pb.GetTerraformProviderByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get terraform provider")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*terraformProviderGetCommand) Synopsis() string {
	return "Get a terraform provider."
}

func (*terraformProviderGetCommand) Usage() string {
	return "tharsis [global options] terraform-provider get [options] <id>"
}

func (*terraformProviderGetCommand) Description() string {
	return `
   Retrieves details about a Terraform provider
   including its name, group, repository URL, and
   privacy setting.
`
}

func (*terraformProviderGetCommand) Example() string {
	return `
tharsis terraform-provider get trn:terraform_provider:<group_path>/<provider_name>
`
}

func (c *terraformProviderGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
