package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type userCreateCommand struct {
	*BaseCommand

	email    *string
	password *string
	admin    *bool
	toJSON   *bool
}

var _ Command = (*userCreateCommand)(nil)

func (c *userCreateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: username")
	}

	return nil
}

// NewUserCreateCommandFactory returns a userCreateCommand struct.
func NewUserCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &userCreateCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *userCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("user create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	user, err := c.grpcClient.UsersClient.CreateUser(c.Context, &pb.CreateUserRequest{
		Username: c.arguments[0],
		Email:    *c.email,
		Password: c.password,
		Admin:    *c.admin,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create user")
		return 1
	}

	return c.Output(user, c.toJSON)
}

func (*userCreateCommand) Synopsis() string {
	return "Create a new user."
}

func (*userCreateCommand) Usage() string {
	return "tharsis [global options] user create [options] <username>"
}

func (*userCreateCommand) Description() string {
	return `
   Creates a new user account with the given
   email address. Use -admin to grant
   administrator privileges.
`
}

func (*userCreateCommand) Example() string {
	return `
tharsis user create -email "<email>" -admin <username>
`
}

func (c *userCreateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.email,
		"email",
		"Email address for the user.",
		flag.Required(),
	)
	f.StringVar(
		&c.password,
		"password",
		"Password for the user.",
	)
	f.BoolVar(
		&c.admin,
		"admin",
		"Whether the user is an admin.",
		flag.Default(false),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
