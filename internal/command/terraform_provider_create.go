package command

import (
	"flag"
	"fmt"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type terraformProviderCreateCommand struct {
	*BaseCommand

	groupID       string
	repositoryURL string
	private       bool
	toJSON        bool
}

// NewTerraformProviderCreateCommandFactory returns a terraformProviderCreateCommand struct.
func NewTerraformProviderCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &terraformProviderCreateCommand{
			BaseCommand: baseCommand,
			private:     true,
		}, nil
	}
}

func (c *terraformProviderCreateCommand) validate() error {
	const message = "provider-name is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *terraformProviderCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("terraform-provider create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	providerName := c.arguments[0]

	parts := strings.Split(providerName, "/")
	if len(parts) == 1 {
		// Ensure a group is supplied when using just provider name argument.
		if c.groupID == "" {
			c.UI.Errorf("group-id is required when supplying the name in the argument")
			return 1
		}
	} else {
		if c.groupID != "" {
			c.UI.Errorf("group-id should not be supplied when using provider path")
			return 1
		}

		// Handle deprecated syntax by extracting name and group path.
		providerName = parts[len(parts)-1]
		c.groupID = trn.NewResourceTRN(trn.ResourceTypeGroup, extractParentPath(providerName))
	}

	input := &pb.CreateTerraformProviderRequest{
		Name:          providerName,
		GroupId:       c.groupID,
		RepositoryUrl: c.repositoryURL,
		Private:       c.private,
	}

	c.Logger.Debug("terraform-provider create input", "input", input)

	provider, err := c.grpcClient.TerraformProvidersClient.CreateTerraformProvider(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create terraform provider")
		return 1
	}

	if c.toJSON {
		if err := c.UI.JSON(provider); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
		return 0
	}

	t := terminal.NewTable("id", "name", "private")
	t.Rich([]string{
		provider.Metadata.Id,
		provider.Name,
		fmt.Sprintf("%t", provider.Private),
	}, nil)

	c.UI.Table(t)
	return 0
}

func (*terraformProviderCreateCommand) Synopsis() string {
	return "Create a new terraform provider."
}

func (*terraformProviderCreateCommand) Description() string {
	return `
   The terraform-provider create command creates a new terraform provider.
`
}

func (*terraformProviderCreateCommand) Usage() string {
	return "tharsis [global options] terraform-provider create [options] <provider-name>"
}

func (*terraformProviderCreateCommand) Example() string {
	return `
tharsis terraform-provider create \
  --group-id trn:group:<group_path> \
  --repository-url https://github.com/example/terraform-provider-example \
  my-provider
`
}

func (c *terraformProviderCreateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.groupID,
		"group-id",
		"",
		"The ID of the group to create the provider in.",
	)
	f.StringVar(
		&c.repositoryURL,
		"repository-url",
		"",
		"The repository URL for this terraform provider.",
	)
	f.BoolVar(
		&c.private,
		"private",
		true,
		"Set to false to allow all groups to view and use the terraform provider.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}
