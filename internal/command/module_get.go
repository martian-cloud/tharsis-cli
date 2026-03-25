package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// moduleGetCommand is the top-level structure for the module get command.
type moduleGetCommand struct {
	*BaseCommand

	toJSON bool
}

var _ Command = (*moduleGetCommand)(nil)

func (c *moduleGetCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewModuleGetCommandFactory returns a moduleGetCommand struct.
func NewModuleGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleGetCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetTerraformModuleByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeTerraformModule, c.arguments[0]),
	}

	module, err := c.grpcClient.TerraformModulesClient.GetTerraformModuleByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get module")
		return 1
	}

	return outputModule(c.UI, c.toJSON, module)
}

func (*moduleGetCommand) Synopsis() string {
	return "Get a single Terraform module."
}

func (*moduleGetCommand) Usage() string {
	return "tharsis [global options] module get [options] <id>"
}

func (*moduleGetCommand) Description() string {
	return `
   The module get command prints information about one Terraform module.
`
}

func (*moduleGetCommand) Example() string {
	return `
tharsis module get trn:terraform_module:<group_path>/<module_name>/<system>
`
}

func (c *moduleGetCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

func outputModule(ui terminal.UI, toJSON bool, module *pb.TerraformModule) int {
	if toJSON {
		if err := ui.JSON(module); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		ui.NamedValues([]terminal.NamedValue{
			{Name: "ID", Value: module.Metadata.Id},
			{Name: "TRN", Value: module.Metadata.Trn},
			{Name: "Name", Value: module.Name},
			{Name: "System", Value: module.System},
			{Name: "Private", Value: module.Private},
			{Name: "Repository URL", Value: module.RepositoryUrl},
			{Name: "Created By", Value: module.CreatedBy},
			{Name: "Created At", Value: module.Metadata.CreatedAt.AsTime().Local().Format(humanTimeFormat)},
			{Name: "Updated At", Value: module.Metadata.UpdatedAt.AsTime().Local().Format(humanTimeFormat)},
		})
	}

	return 0
}
