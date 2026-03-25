package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type moduleDeleteVersionCommand struct {
	*BaseCommand

	version *int64
}

// NewModuleDeleteVersionCommandFactory returns a moduleDeleteVersionCommand struct.
func NewModuleDeleteVersionCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &moduleDeleteVersionCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *moduleDeleteVersionCommand) validate() error {
	const message = "version-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *moduleDeleteVersionCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("module delete-version"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.DeleteTerraformModuleVersionRequest{
		Id:      c.arguments[0],
		Version: c.version,
	}

	if _, err := c.grpcClient.TerraformModulesClient.DeleteTerraformModuleVersion(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete module version")
		return 1
	}

	c.UI.Successf("Module version deleted successfully!")
	return 0
}

func (*moduleDeleteVersionCommand) Synopsis() string {
	return "Delete a module version."
}

func (*moduleDeleteVersionCommand) Description() string {
	return `
   The module delete-version command deletes a module version.
`
}

func (*moduleDeleteVersionCommand) Usage() string {
	return "tharsis [global options] module delete-version [options] <version-id>"
}

func (*moduleDeleteVersionCommand) Example() string {
	return `
tharsis module delete-version trn:terraform_module_version:<group_path>/<module_name>/<system>/<semantic_version>
`
}

func (c *moduleDeleteVersionCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Metadata version of the resource to be deleted. In most cases, this is not required.",
	)

	return f
}
