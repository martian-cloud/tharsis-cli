package command

import (
	"errors"

	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"google.golang.org/protobuf/types/known/emptypb"
)

// roleGetAvailablePermissionsCommand is the top-level structure for the role get-available-permissions command.
type roleGetAvailablePermissionsCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*roleGetAvailablePermissionsCommand)(nil)

func (c *roleGetAvailablePermissionsCommand) validate() error {
	if len(c.arguments) != 0 {
		return errors.New("no arguments expected")
	}

	return nil
}

// NewRoleGetAvailablePermissionsCommandFactory returns a roleGetAvailablePermissionsCommand struct.
func NewRoleGetAvailablePermissionsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &roleGetAvailablePermissionsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *roleGetAvailablePermissionsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("role get-available-permissions"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.RolesClient.GetAvailablePermissions(c.Context, &emptypb.Empty{})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get available permissions")
		return 1
	}

	if c.toJSON != nil && *c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}

		return 0
	}

	for _, p := range result.Permissions {
		c.UI.Output(p)
	}

	return 0
}

func (*roleGetAvailablePermissionsCommand) Synopsis() string {
	return "Get available permissions for roles."
}

func (*roleGetAvailablePermissionsCommand) Usage() string {
	return "tharsis [global options] role get-available-permissions [options]"
}

func (*roleGetAvailablePermissionsCommand) Description() string {
	return `
   Returns the list of available permissions that can be
   assigned to roles.
`
}

func (*roleGetAvailablePermissionsCommand) Example() string {
	return `
tharsis role get-available-permissions
`
}

func (c *roleGetAvailablePermissionsCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
