package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type terraformProviderUpdateCommand struct {
	*BaseCommand

	repositoryURL *string
	private       *bool
	version       *int64
	toJSON        *bool
}

var _ Command = (*terraformProviderUpdateCommand)(nil)

func (c *terraformProviderUpdateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewTerraformProviderUpdateCommandFactory returns a terraformProviderUpdateCommand struct.
func NewTerraformProviderUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderUpdateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *terraformProviderUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.TerraformProvidersClient.UpdateTerraformProvider(c.Context, &pb.UpdateTerraformProviderRequest{
		Id:            c.arguments[0],
		RepositoryUrl: c.repositoryURL,
		Private:       c.private,
		Version:       c.version,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update terraform provider")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*terraformProviderUpdateCommand) Synopsis() string {
	return "Update a terraform provider."
}

func (*terraformProviderUpdateCommand) Usage() string {
	return "tharsis [global options] terraform-provider update [options] <id>"
}

func (*terraformProviderUpdateCommand) Description() string {
	return `
   Updates a Terraform provider's repository URL or
   privacy setting.
`
}

func (*terraformProviderUpdateCommand) Example() string {
	return `
tharsis terraform-provider update \
  -repository-url "https://github.com/example/terraform-provider-example" \
  <terraform_provider_id>
`
}

func (c *terraformProviderUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.repositoryURL,
		"repository-url",
		"The repository URL for this terraform provider.",
	)
	f.BoolVar(
		&c.private,
		"private",
		"Set to false to allow all groups to view and use the terraform provider.",
	)
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
