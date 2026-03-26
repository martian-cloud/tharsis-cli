package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// managedIdentityAccessRuleDeleteCommand is the top-level structure for the managed identity access rule delete command.
type managedIdentityAccessRuleDeleteCommand struct {
	*BaseCommand
}

var _ Command = (*managedIdentityAccessRuleDeleteCommand)(nil)

func (c *managedIdentityAccessRuleDeleteCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewManagedIdentityAccessRuleDeleteCommandFactory returns a managedIdentityAccessRuleDeleteCommand struct.
func NewManagedIdentityAccessRuleDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityAccessRuleDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityAccessRuleDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("managed-identity-access-rule delete"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.DeleteManagedIdentityAccessRuleRequest{
		Id: c.arguments[0],
	}

	if _, err := c.grpcClient.ManagedIdentitiesClient.DeleteManagedIdentityAccessRule(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete managed identity access rule")
		return 1
	}

	c.UI.Successf("Managed identity access rule deleted successfully!")

	return 0
}

func (*managedIdentityAccessRuleDeleteCommand) Synopsis() string {
	return "Delete a managed identity access rule."
}

func (*managedIdentityAccessRuleDeleteCommand) Usage() string {
	return "tharsis [global options] managed-identity-access-rule delete [options] <id>"
}

func (*managedIdentityAccessRuleDeleteCommand) Description() string {
	return `
   The managed-identity-access-rule delete command deletes a managed identity access rule.
`
}

func (*managedIdentityAccessRuleDeleteCommand) Example() string {
	return `
tharsis managed-identity-access-rule delete <id>
`
}
