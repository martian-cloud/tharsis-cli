package command

import (
	"errors"

	"github.com/aws/smithy-go/ptr"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

// roleCreateCommand is the top-level structure for the role create command.
type roleCreateCommand struct {
	*BaseCommand

	description *string
	permissions []string
	toJSON      *bool
}

var _ Command = (*roleCreateCommand)(nil)

func (c *roleCreateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: name")
	}

	return nil
}

// NewRoleCreateCommandFactory returns a roleCreateCommand struct.
func NewRoleCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &roleCreateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *roleCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("role create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	createdRole, err := c.grpcClient.RolesClient.CreateRole(c.Context, &pb.CreateRoleRequest{
		Name:        c.arguments[0],
		Description: ptr.ToString(c.description),
		Permissions: c.permissions,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create role")
		return 1
	}

	return c.Output(createdRole, c.toJSON)
}

func (*roleCreateCommand) Synopsis() string {
	return "Create a new role."
}

func (*roleCreateCommand) Usage() string {
	return "tharsis [global options] role create [options] <name>"
}

func (*roleCreateCommand) Description() string {
	return `
   Creates a new role with the specified
   permissions. Roles define a set of
   permissions assignable to users, service
   accounts, or teams via memberships.
`
}

func (*roleCreateCommand) Example() string {
	return `
tharsis role create \
  -description "<description>" \
  -permission "run:create" \
  -permission "workspace:view" \
  <name>
`
}

func (c *roleCreateCommand) Flags() *flag.Set {
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
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
