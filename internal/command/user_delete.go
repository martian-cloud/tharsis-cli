package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type userDeleteCommand struct {
	*BaseCommand
}

var _ Command = (*userDeleteCommand)(nil)

func (c *userDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewUserDeleteCommandFactory returns a userDeleteCommand struct.
func NewUserDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &userDeleteCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *userDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("user delete"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.UsersClient.DeleteUser(c.Context, &pb.DeleteUserRequest{Id: c.arguments[0]}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete user")
		return 1
	}

	c.UI.Successf("User deleted successfully!")
	return 0
}

func (*userDeleteCommand) Synopsis() string {
	return "Delete a user."
}

func (*userDeleteCommand) Usage() string {
	return "tharsis [global options] user delete <id>"
}

func (*userDeleteCommand) Description() string {
	return `
   Permanently deletes a user. This is
   irreversible and removes all memberships
   and access for the user.
`
}

func (*userDeleteCommand) Example() string {
	return `
tharsis user delete <user_id>
`
}
