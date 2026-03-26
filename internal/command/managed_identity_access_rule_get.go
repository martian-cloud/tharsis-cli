package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// managedIdentityAccessRuleGetCommand is the top-level structure for the managed identity access rule get command.
type managedIdentityAccessRuleGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*managedIdentityAccessRuleGetCommand)(nil)

func (c *managedIdentityAccessRuleGetCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewManagedIdentityAccessRuleGetCommandFactory returns a managedIdentityAccessRuleGetCommand struct.
func NewManagedIdentityAccessRuleGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityAccessRuleGetCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityAccessRuleGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity-access-rule get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetManagedIdentityAccessRuleByIDRequest{Id: c.arguments[0]}

	rule, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentityAccessRuleByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get managed identity access rule")
		return 1
	}

	return c.OutputProto(rule, c.toJSON)
}

func (*managedIdentityAccessRuleGetCommand) Synopsis() string {
	return "Get a managed identity access rule."
}

func (*managedIdentityAccessRuleGetCommand) Usage() string {
	return "tharsis [global options] managed-identity-access-rule get [options] <id>"
}

func (*managedIdentityAccessRuleGetCommand) Description() string {
	return `
   The managed-identity-access-rule get command gets a managed identity access rule by ID.
`
}

func (*managedIdentityAccessRuleGetCommand) Example() string {
	return `
tharsis managed-identity-access-rule get <id>
`
}

func (c *managedIdentityAccessRuleGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
