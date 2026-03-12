package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// moduleDeleteCommand is the top-level structure for the module delete command.
type moduleDeleteCommand struct {
	*BaseCommand
}

var _ Command = (*moduleDeleteCommand)(nil)

func (c *moduleDeleteCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewModuleDeleteCommandFactory returns a moduleDeleteCommand struct.
func NewModuleDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("module delete"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.DeleteTerraformModuleRequest{
		Id: toTRN(trn.ResourceTypeTerraformModule, c.arguments[0]),
	}

	c.Logger.Debug("module delete input", "input", input)

	if _, err := c.grpcClient.TerraformModulesClient.DeleteTerraformModule(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete a module")
		return 1
	}

	c.UI.Successf("Module deleted successfully!")
	return 0
}

func (*moduleDeleteCommand) Synopsis() string {
	return "Delete a Terraform module."
}

func (*moduleDeleteCommand) Usage() string {
	return "tharsis [global options] module delete [options] <id>"
}

func (*moduleDeleteCommand) Description() string {
	return `
   The module delete command deletes a Terraform module.
`
}

func (*moduleDeleteCommand) Example() string {
	return `
tharsis module delete trn:terraform_module:<group_path>/<module_name>/<system>
`
}

func (*moduleDeleteCommand) Flags() *flag.FlagSet {
	return nil
}
