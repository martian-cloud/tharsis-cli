package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type gpgKeyGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*gpgKeyGetCommand)(nil)

func (c *gpgKeyGetCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewGPGKeyGetCommandFactory returns a gpgKeyGetCommand struct.
func NewGPGKeyGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &gpgKeyGetCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *gpgKeyGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("gpg-key get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	key, err := c.grpcClient.GPGKeysClient.GetGPGKeyByID(c.Context, &pb.GetGPGKeyByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get GPG key")
		return 1
	}

	return c.Output(key, c.toJSON)
}

func (*gpgKeyGetCommand) Synopsis() string {
	return "Get a GPG key."
}

func (*gpgKeyGetCommand) Usage() string {
	return "tharsis [global options] gpg-key get [options] <id>"
}

func (*gpgKeyGetCommand) Description() string {
	return `
   Retrieves details about a GPG key
   including its ASCII-armored public key,
   fingerprint, and associated group.
`
}

func (*gpgKeyGetCommand) Example() string {
	return `
tharsis gpg-key get <gpg_key_id>
`
}

func (c *gpgKeyGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
