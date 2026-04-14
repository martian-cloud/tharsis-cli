package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// roleGetCommand is the top-level structure for the role get command.
type roleGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*roleGetCommand)(nil)

func (c *roleGetCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewRoleGetCommandFactory returns a roleGetCommand struct.
func NewRoleGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &roleGetCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *roleGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("role get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	role, err := c.grpcClient.RolesClient.GetRoleByID(c.Context, &pb.GetRoleByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get role")
		return 1
	}

	return c.Output(role, c.toJSON)
}

func (*roleGetCommand) Synopsis() string {
	return "Get a role."
}

func (*roleGetCommand) Usage() string {
	return "tharsis [global options] role get [options] <id>"
}

func (*roleGetCommand) Description() string {
	return `
   Retrieves details about a role including
   its name, description, and assigned
   permissions.
`
}

func (*roleGetCommand) Example() string {
	return `
tharsis role get <role_id>
`
}

func (c *roleGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
