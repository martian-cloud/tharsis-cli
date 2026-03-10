package command

import (
	"flag"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
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
		Id: c.arguments[0],
	}

	c.Logger.Debug("module get input", "input", input)

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
   The module get command prints information about one
   Terraform module.
`
}

func (*moduleGetCommand) Example() string {
	return `
tharsis module get trn:terraform_module:ops/my-group/vpc
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
		t := terminal.NewTable("id", "name", "system", "group_id", "private")
		t.Rich([]string{
			module.Metadata.Id,
			module.Name,
			module.System,
			module.GroupId,
			fmt.Sprintf("%t", module.Private),
		}, nil)

		ui.Table(t)
	}

	return 0
}
