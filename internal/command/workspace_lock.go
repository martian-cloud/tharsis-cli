package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type workspaceLockCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*workspaceLockCommand)(nil)

func (c *workspaceLockCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
}

// NewWorkspaceLockCommandFactory returns a workspaceLockCommand struct.
func NewWorkspaceLockCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceLockCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *workspaceLockCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace lock"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.WorkspacesClient.LockWorkspace(c.Context, &pb.LockWorkspaceRequest{
		WorkspaceId: c.arguments[0],
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to lock workspace")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*workspaceLockCommand) Synopsis() string {
	return "Lock a workspace."
}

func (*workspaceLockCommand) Usage() string {
	return "tharsis [global options] workspace lock [options] <workspace-id>"
}

func (*workspaceLockCommand) Description() string {
	return `
   Locks a workspace to prevent new runs from being
   queued or created. Useful during maintenance windows
   or when coordinating infrastructure changes across
   teams. A workspace is also automatically locked while
   a run is actively executing. VCS-triggered and
   manually created runs will be rejected until the
   workspace is unlocked.
`
}

func (*workspaceLockCommand) Example() string {
	return `
tharsis workspace lock <workspace_id>
`
}

func (c *workspaceLockCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
