package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type gpgKeyCreateCommand struct {
	*BaseCommand

	groupID    *string
	asciiArmor *string
	toJSON     *bool
}

var _ Command = (*gpgKeyCreateCommand)(nil)

func (c *gpgKeyCreateCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

// NewGPGKeyCreateCommandFactory returns a gpgKeyCreateCommand struct.
func NewGPGKeyCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &gpgKeyCreateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *gpgKeyCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("gpg-key create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	createdKey, err := c.grpcClient.GPGKeysClient.CreateGPGKey(c.Context, &pb.CreateGPGKeyRequest{
		GroupId:    *c.groupID,
		AsciiArmor: *c.asciiArmor,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create GPG key")
		return 1
	}

	return c.Output(createdKey, c.toJSON)
}

func (*gpgKeyCreateCommand) Synopsis() string {
	return "Create a new GPG key."
}

func (*gpgKeyCreateCommand) Usage() string {
	return "tharsis [global options] gpg-key create [options]"
}

func (*gpgKeyCreateCommand) Description() string {
	return `
   Creates a new GPG key within a group.
   GPG keys are used to verify Terraform
   module attestations. The key is used to
   sign or verify module versions.
`
}

func (*gpgKeyCreateCommand) Example() string {
	return `
tharsis gpg-key create \
  -group-id "trn:group:<group_path>" \
  -ascii-armor "-----BEGIN PGP PUBLIC KEY BLOCK-----..."
`
}

func (c *gpgKeyCreateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.groupID,
		"group-id",
		"Group ID or TRN where the GPG key will be created.",
		flag.Required(),
	)
	f.StringVar(
		&c.asciiArmor,
		"ascii-armor",
		"ASCII-armored GPG public key.",
		flag.Required(),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
