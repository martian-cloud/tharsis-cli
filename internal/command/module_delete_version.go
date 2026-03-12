package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type moduleDeleteVersionCommand struct {
	*BaseCommand

	version *int64
	force   bool
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

	c.Logger.Debug("module delete version input", "input", input)

	_, err := c.grpcClient.TerraformModulesClient.DeleteTerraformModuleVersion(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete module version")
		return 1
	}

	c.UI.Output("Module version deleted successfully!", terminal.WithSuccessStyle())
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

func (c *moduleDeleteVersionCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"version",
		"Metadata version of the resource to be deleted. "+
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
		&c.force,
		"force",
		false,
		"Force deletion without confirmation.",
	)
	return f
}
