package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type gpgKeyDeleteCommand struct {
	*BaseCommand
}

var _ Command = (*gpgKeyDeleteCommand)(nil)

func (c *gpgKeyDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewGPGKeyDeleteCommandFactory returns a gpgKeyDeleteCommand struct.
func NewGPGKeyDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &gpgKeyDeleteCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *gpgKeyDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("gpg-key delete"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.GPGKeysClient.DeleteGPGKey(c.Context, &pb.DeleteGPGKeyRequest{Id: c.arguments[0]}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete GPG key")
		return 1
	}

	c.UI.Successf("GPG key deleted successfully!")
	return 0
}

func (*gpgKeyDeleteCommand) Synopsis() string {
	return "Delete a GPG key."
}

func (*gpgKeyDeleteCommand) Usage() string {
	return "tharsis [global options] gpg-key delete <id>"
}

func (*gpgKeyDeleteCommand) Description() string {
	return `
   Permanently removes a GPG key. This
   action is irreversible. Any module
   attestations signed with this key can
   no longer be verified.
`
}

func (*gpgKeyDeleteCommand) Example() string {
	return `
tharsis gpg-key delete <gpg_key_id>
`
}
