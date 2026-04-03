package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// moduleDeleteCommand is the top-level structure for the module delete command.
type moduleDeleteCommand struct {
	*BaseCommand
}

var _ Command = (*moduleDeleteCommand)(nil)

func (c *moduleDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
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
		Id: trn.ToTRN(trn.ResourceTypeTerraformModule, c.arguments[0]),
	}

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
   Permanently removes a module and all its versions
   from the registry.
`
}

func (*moduleDeleteCommand) Example() string {
	return `
tharsis module delete trn:terraform_module:<group_path>/<module_name>/<system>
`
}

func (*moduleDeleteCommand) Flags() *flag.Set {
	return nil
}
