package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// managedIdentityAliasDeleteCommand is the top-level structure for the managed identity alias delete command.
type managedIdentityAliasDeleteCommand struct {
	*BaseCommand

	force *bool
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
		WithWarningPrompt("This will permanently delete the managed identity alias."),
	); code != 0 {
		return code
	}

	input := &pb.DeleteManagedIdentityAliasRequest{
		Id:    trn.ToTRN(trn.ResourceTypeManagedIdentity, c.arguments[0]),
		Force: c.force,
	}

	if _, err := c.grpcClient.ManagedIdentitiesClient.DeleteManagedIdentityAlias(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete managed identity alias")
		return 1
	}

	c.UI.Successf("Managed identity alias deleted successfully!")

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
tharsis managed-identity-alias delete trn:managed_identity:<group_path>/<managed_identity_name>
`
}

func (c *managedIdentityAliasDeleteCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.force,
		"force",
		"Force delete the managed identity alias.",
		flag.Aliases("f"),
	)

	return f
}
