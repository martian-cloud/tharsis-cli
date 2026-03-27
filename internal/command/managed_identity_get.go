package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// managedIdentityGetCommand is the top-level structure for the managed identity get command.
type managedIdentityGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*managedIdentityGetCommand)(nil)

func (c *managedIdentityGetCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewManagedIdentityGetCommandFactory returns a managedIdentityGetCommand struct.
func NewManagedIdentityGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &managedIdentityGetCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *managedIdentityGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("managed-identity get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetManagedIdentityByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeManagedIdentity, c.arguments[0]),
	}

	identity, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentityByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get managed identity")
		return 1
	}

	return c.Output(identity, c.toJSON)
}

func (*managedIdentityGetCommand) Synopsis() string {
	return "Get a single managed identity."
}

func (*managedIdentityGetCommand) Usage() string {
	return "tharsis [global options] managed-identity get [options] <id>"
}

func (*managedIdentityGetCommand) Description() string {
	return `
   The managed-identity get command prints information about one
   managed identity.
`
}

func (*managedIdentityGetCommand) Example() string {
	return `
tharsis managed-identity get trn:managed_identity:<group_path>/<managed_identity_name>
`
}

func (c *managedIdentityGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
