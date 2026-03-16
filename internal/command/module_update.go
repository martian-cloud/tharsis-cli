package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// moduleUpdateCommand is the top-level structure for the module update command.
type moduleUpdateCommand struct {
	*BaseCommand

	repositoryURL *string
	private       *bool
	version       *int64
	toJSON        bool
}

var _ Command = (*moduleUpdateCommand)(nil)

func (c *moduleUpdateCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewModuleUpdateCommandFactory returns a moduleUpdateCommand struct.
func NewModuleUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleUpdateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.UpdateTerraformModuleRequest{
		Id:            toTRN(trn.ResourceTypeTerraformModule, c.arguments[0]),
		RepositoryUrl: c.repositoryURL,
		Private:       c.private,
		Version:       c.version,
	}

	updatedModule, err := c.grpcClient.TerraformModulesClient.UpdateTerraformModule(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update a module")
		return 1
	}

	return outputModule(c.UI, c.toJSON, updatedModule)
}

func (*moduleUpdateCommand) Synopsis() string {
	return "Update a Terraform module."
}

func (*moduleUpdateCommand) Usage() string {
	return "tharsis [global options] module update [options] <id>"
}

func (*moduleUpdateCommand) Description() string {
	return `
   The module update command updates a Terraform module.
   Currently, it supports updating the repository URL and
   private flag. Shows final output as JSON, if specified.
`
}

func (*moduleUpdateCommand) Example() string {
	return `
tharsis module update \
  --repository-url https://github.com/example/terraform-aws-vpc-v2 \
  --private true \
  trn:terraform_module:<group_path>/<module_name>/<system>
`
}

func (c *moduleUpdateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"repository-url",
		"The repository URL for the module.",
		func(s string) error {
			c.repositoryURL = &s
			return nil
		},
	)
	f.Func(
		"private",
		"Whether the module is private.",
		func(s string) error {
			v, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}
			c.private = &v
			return nil
		},
	)
	f.Func(
		"version",
		"Metadata version of the resource to be updated. "+
			"In most cases, this is not required.",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			c.version = &v
			return nil
		},
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
