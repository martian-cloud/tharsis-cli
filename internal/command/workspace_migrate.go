package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type workspaceMigrateCommand struct {
	*BaseCommand

	newGroupID *string
	toJSON     *bool
}

var _ Command = (*workspaceMigrateCommand)(nil)

// NewWorkspaceMigrateCommandFactory returns a workspaceMigrateCommand struct.
func NewWorkspaceMigrateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceMigrateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceMigrateCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
}

func (c *workspaceMigrateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace migrate"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.MigrateWorkspaceRequest{
		WorkspaceId: c.arguments[0],
		NewGroupId:  *c.newGroupID,
	}

	workspace, err := c.grpcClient.WorkspacesClient.MigrateWorkspace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to migrate workspace")
		return 1
	}

	return c.Output(workspace, c.toJSON)
}

func (*workspaceMigrateCommand) Synopsis() string {
	return "Migrate a workspace to a new group."
}

func (*workspaceMigrateCommand) Description() string {
	return `
   Moves a workspace to a different group.
`
}

func (*workspaceMigrateCommand) Usage() string {
	return "tharsis [global options] workspace migrate [options] <workspace-id>"
}

func (*workspaceMigrateCommand) Example() string {
	return `
tharsis workspace migrate \
  -new-group-id "trn:group:<group_path>" \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceMigrateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.newGroupID,
		"new-group-id",
		"New parent group ID.",
		flag.Required(),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
