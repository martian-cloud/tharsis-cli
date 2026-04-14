package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// roleUpdateCommand is the top-level structure for the role update command.
type roleUpdateCommand struct {
	*BaseCommand

	description *string
	version     *int64
	permissions []string
	toJSON      *bool
}

var _ Command = (*roleUpdateCommand)(nil)

func (c *roleUpdateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewRoleUpdateCommandFactory returns a roleUpdateCommand struct.
func NewRoleUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &roleUpdateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *roleUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("role update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	updatedRole, err := c.grpcClient.RolesClient.UpdateRole(c.Context, &pb.UpdateRoleRequest{
		Id:          c.arguments[0],
		Description: c.description,
		Version:     c.version,
		Permissions: c.permissions,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update role")
		return 1
	}

	return c.Output(updatedRole, c.toJSON)
}

func (*roleUpdateCommand) Synopsis() string {
	return "Update a role."
}

func (*roleUpdateCommand) Usage() string {
	return "tharsis [global options] role update [options] <id>"
}

func (*roleUpdateCommand) Description() string {
	return `
   Updates a role's description or permissions.
   When permissions are specified, they fully
   replace the existing set.
`
}

func (*roleUpdateCommand) Example() string {
	return `
tharsis role update \
  -description "<description>" \
  -permission "run:create" \
  -permission "workspace:view" \
  <role_id>
`
}

func (c *roleUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the role.",
	)
	f.StringSliceVar(
		&c.permissions,
		"permission",
		"Permission to assign to the role.",
	)
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
