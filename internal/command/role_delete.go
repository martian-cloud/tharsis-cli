package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// roleDeleteCommand is the top-level structure for the role delete command.
type roleDeleteCommand struct {
	*BaseCommand

	version *int64
}

var _ Command = (*roleDeleteCommand)(nil)

func (c *roleDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewRoleDeleteCommandFactory returns a roleDeleteCommand struct.
func NewRoleDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &roleDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *roleDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("role delete"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	if _, err := c.grpcClient.RolesClient.DeleteRole(c.Context, &pb.DeleteRoleRequest{
		Id:      c.arguments[0],
		Version: c.version,
	}); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete role")
		return 1
	}

	c.UI.Successf("Role deleted successfully!")
	return 0
}

func (*roleDeleteCommand) Synopsis() string {
	return "Delete a role."
}

func (*roleDeleteCommand) Usage() string {
	return "tharsis [global options] role delete [options] <id>"
}

func (*roleDeleteCommand) Description() string {
	return `
   Permanently removes a role. This action
   is irreversible. Any memberships using
   this role will lose the associated
   permissions.
`
}

func (*roleDeleteCommand) Example() string {
	return `
tharsis role delete <role_id>
`
}

func (c *roleDeleteCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)

	return f
}
