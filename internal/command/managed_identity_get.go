package command

import (
	"encoding/base64"
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// managedIdentityGetCommand is the top-level structure for the managed identity get command.
type managedIdentityGetCommand struct {
	*BaseCommand

	toJSON bool
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
		Id: toTRN(trn.ResourceTypeManagedIdentity, c.arguments[0]),
	}

	c.Logger.Debug("managed identity get input", "input", input)

	identity, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentityByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get managed identity")
		return 1
	}

	return outputManagedIdentity(c.UI, c.toJSON, identity)
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

func (c *managedIdentityGetCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}

func outputManagedIdentity(ui terminal.UI, toJSON bool, identity *pb.ManagedIdentity) int {
	if toJSON {
		if err := ui.JSON(identity); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		// Decode base64 data
		decoded, err := base64.StdEncoding.DecodeString(identity.Data)
		if err != nil {
			ui.ErrorWithSummary(err, "failed to decode identity data")
			return 1
		}

		t := terminal.NewTable("id", "name", "type", "group_id", "is_alias", "data")
		t.Rich([]string{
			identity.Metadata.Id,
			identity.Name,
			identity.Type,
			identity.GroupId,
			strconv.FormatBool(identity.AliasSourceId != nil),
			string(decoded),
		}, nil)

		ui.Table(t)
	}

	return 0
}
