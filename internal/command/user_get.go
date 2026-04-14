package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type userGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*userGetCommand)(nil)

func (c *userGetCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewUserGetCommandFactory returns a userGetCommand struct.
func NewUserGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &userGetCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *userGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("user get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	user, err := c.grpcClient.UsersClient.GetUserByID(c.Context, &pb.GetUserByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get user")
		return 1
	}

	return c.Output(user, c.toJSON)
}

func (*userGetCommand) Synopsis() string {
	return "Get a user."
}

func (*userGetCommand) Usage() string {
	return "tharsis [global options] user get [options] <id>"
}

func (*userGetCommand) Description() string {
	return `
   Returns user details including username,
   email, and admin status.
`
}

func (*userGetCommand) Example() string {
	return `
tharsis user get <user_id>
`
}

func (c *userGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
