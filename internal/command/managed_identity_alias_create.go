package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// managedIdentityAliasCreateCommand is the top-level structure for the managed identity alias create command.
type managedIdentityAliasCreateCommand struct {
	*BaseCommand

	name          *string
	groupID       string
	aliasSourceID string
	toJSON        bool
}

var _ Command = (*managedIdentityAliasCreateCommand)(nil)

func (c *managedIdentityAliasCreateCommand) validate() error {
	const message = "name is required either as an argument or a flag"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.When(c.name == nil).Error(message),
		),
		validation.Field(&c.name, validation.Required.When(len(c.arguments) == 0).Error(message)),
		validation.Field(&c.groupID, validation.Required),
		validation.Field(&c.aliasSourceID, validation.Required),
	)
}

// NewManagedIdentityAliasCreateCommandFactory returns a managedIdentityAliasCreateCommand struct.
func NewManagedIdentityAliasCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityAliasCreateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityAliasCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity-alias create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if c.name != nil {
		c.arguments = append(c.arguments, *c.name)
	}

	input := &pb.CreateManagedIdentityAliasRequest{
		Name:          c.arguments[0],
		AliasSourceId: c.aliasSourceID,
		GroupId:       c.groupID,
	}

	createdAlias, err := c.grpcClient.ManagedIdentitiesClient.CreateManagedIdentityAlias(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a managed identity alias")
		return 1
	}

	return outputManagedIdentity(c.UI, c.toJSON, createdAlias)
}

func (*managedIdentityAliasCreateCommand) Synopsis() string {
	return "Create a new managed identity alias."
}

func (*managedIdentityAliasCreateCommand) Usage() string {
	return "tharsis [global options] managed-identity-alias create [options] <name>"
}

func (*managedIdentityAliasCreateCommand) Description() string {
	return `
   The managed-identity-alias create command creates a new managed identity alias.
`
}

func (*managedIdentityAliasCreateCommand) Example() string {
	return `
tharsis managed-identity-alias create \
  --group-id trn:group:<group_path> \
  --alias-source-id trn:managed_identity:<group_path>/<source_identity_name> \
  prod-identity-alias
`
}

func (c *managedIdentityAliasCreateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.groupID,
		"group-id",
		"",
		"Group ID or TRN where the managed identity alias will be created.",
	)
	f.StringVar(
		&c.aliasSourceID,
		"alias-source-id",
		"",
		"The ID or TRN of the source managed identity.",
	)
	f.Func(
		"alias-source-path",
		"The alias source path. Deprecated.",
		func(s string) error {
			c.aliasSourceID = trn.NewResourceTRN(trn.ResourceTypeManagedIdentity, s)
			return nil
		},
	)
	f.Func(
		"group-path",
		"Full path of the group where the managed identity alias will be created. Deprecated",
		func(s string) error {
			c.groupID = trn.NewResourceTRN(trn.ResourceTypeGroup, s)
			return nil
		},
	)
	f.Func(
		"name",
		"The name of the managed identity alias. Deprecated",
		func(s string) error {
			c.name = &s
			return nil
		},
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
