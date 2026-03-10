package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// managedIdentityAliasDeleteCommand is the top-level structure for the managed identity alias delete command.
type managedIdentityAliasDeleteCommand struct {
	*BaseCommand

	force bool
}

var _ Command = (*managedIdentityAliasDeleteCommand)(nil)

func (c *managedIdentityAliasDeleteCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewManagedIdentityAliasDeleteCommandFactory returns a managedIdentityAliasDeleteCommand struct.
func NewManagedIdentityAliasDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityAliasDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityAliasDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity-alias delete"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithForcePrompt("Are you sure you want to delete this managed identity alias?"),
	); code != 0 {
		return code
	}

	input := &pb.DeleteManagedIdentityAliasRequest{
		Id:    c.arguments[0],
		Force: &c.force,
	}

	c.Logger.Debug("managed identity alias delete input", "input", input)

	if _, err := c.grpcClient.ManagedIdentitiesClient.DeleteManagedIdentityAlias(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete managed identity alias")
		return 1
	}

	c.UI.Output("Managed identity alias deleted successfully!")

	return 0
}

func (*managedIdentityAliasDeleteCommand) Synopsis() string {
	return "Delete a managed identity alias."
}

func (*managedIdentityAliasDeleteCommand) Usage() string {
	return "tharsis [global options] managed-identity-alias delete [options] <id>"
}

func (*managedIdentityAliasDeleteCommand) Description() string {
	return `
   The managed-identity-alias delete command deletes a managed identity alias.
`
}

func (*managedIdentityAliasDeleteCommand) Example() string {
	return `
tharsis managed-identity-alias delete trn:managed_identity:ops/my-group/prod-identity-alias
`
}

func (c *managedIdentityAliasDeleteCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.force,
		"force",
		false,
		"Force delete the managed identity alias.",
	)

	return f
}
