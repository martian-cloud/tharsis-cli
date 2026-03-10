package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// managedIdentityDeleteCommand is the top-level structure for the managed identity delete command.
type managedIdentityDeleteCommand struct {
	*BaseCommand

	force bool
}

var _ Command = (*managedIdentityDeleteCommand)(nil)

func (c *managedIdentityDeleteCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewManagedIdentityDeleteCommandFactory returns a managedIdentityDeleteCommand struct.
func NewManagedIdentityDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity delete"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithForcePrompt("Are you sure you want to delete this managed identity?"),
	); code != 0 {
		return code
	}

	input := &pb.DeleteManagedIdentityRequest{
		Id:    c.arguments[0],
		Force: &c.force,
	}

	c.Logger.Debug("managed identity delete input", "input", input)

	if _, err := c.grpcClient.ManagedIdentitiesClient.DeleteManagedIdentity(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete a managed identity")
		return 1
	}

	c.UI.Successf("Managed identity deleted successfully!")
	return 0
}

func (*managedIdentityDeleteCommand) Synopsis() string {
	return "Delete a managed identity."
}

func (*managedIdentityDeleteCommand) Usage() string {
	return "tharsis [global options] managed-identity delete [options] <id>"
}

func (*managedIdentityDeleteCommand) Description() string {
	return `
   The managed-identity delete command deletes a managed identity.

   Use with caution as deleting a managed identity is irreversible!
`
}

func (*managedIdentityDeleteCommand) Example() string {
	return `
tharsis managed-identity delete --force trn:managed_identity:ops/my-group/aws-prod
`
}

func (c *managedIdentityDeleteCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.force,
		"force",
		false,
		"Force delete the managed identity.",
	)

	return f
}
