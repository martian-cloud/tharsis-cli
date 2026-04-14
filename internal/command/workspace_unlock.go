package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type workspaceUnlockCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*workspaceUnlockCommand)(nil)

func (c *workspaceUnlockCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	return nil
}

// NewWorkspaceUnlockCommandFactory returns a workspaceUnlockCommand struct.
func NewWorkspaceUnlockCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceUnlockCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *workspaceUnlockCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace unlock"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	result, err := c.grpcClient.WorkspacesClient.UnlockWorkspace(c.Context, &pb.UnlockWorkspaceRequest{
		WorkspaceId: c.arguments[0],
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to unlock workspace")
		return 1
	}

	return c.Output(result, c.toJSON)
}

func (*workspaceUnlockCommand) Synopsis() string {
	return "Unlock a workspace."
}

func (*workspaceUnlockCommand) Usage() string {
	return "tharsis [global options] workspace unlock [options] <workspace-id>"
}

func (*workspaceUnlockCommand) Description() string {
	return `
   Unlocks a workspace so that new runs can be queued
   and created again. A workspace that is locked by an
   active run cannot be manually unlocked — the lock
   is released automatically when the run completes.
   Only manually applied locks can be removed with
   this command.
`
}

func (*workspaceUnlockCommand) Example() string {
	return `
tharsis workspace unlock <workspace_id>
`
}

func (c *workspaceUnlockCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
