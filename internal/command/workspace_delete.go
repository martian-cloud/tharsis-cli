package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// workspaceDeleteCommand is the top-level structure for the workspace delete command.
type workspaceDeleteCommand struct {
	*BaseCommand

	version *int64
	force   *bool
}

var _ Command = (*workspaceDeleteCommand)(nil)

func (c *workspaceDeleteCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewWorkspaceDeleteCommandFactory returns a workspaceDeleteCommand struct.
func NewWorkspaceDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace delete"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithWarningPrompt("This will delete the workspace even if resources are still deployed."),
	); code != 0 {
		return code
	}

	input := &pb.DeleteWorkspaceRequest{
		Id:      trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
		Version: c.version,
		Force:   c.force,
	}

	if _, err := c.grpcClient.WorkspacesClient.DeleteWorkspace(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete a workspace")
		return 1
	}

	c.UI.Successf("Workspace deleted successfully!")
	return 0
}

func (*workspaceDeleteCommand) Synopsis() string {
	return "Delete a workspace."
}

func (*workspaceDeleteCommand) Usage() string {
	return "tharsis [global options] workspace delete [options] <id>"
}

func (*workspaceDeleteCommand) Description() string {
	return `
   Permanently removes a workspace. Use -force to delete
   even if resources are deployed.
`
}

func (*workspaceDeleteCommand) Example() string {
	return `
tharsis workspace delete -force trn:workspace:<workspace_path>
`
}

func (c *workspaceDeleteCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Optimistic locking version. Usually not required.",
	)
	f.BoolVar(
		&c.force,
		"force",
		"Force the workspace to delete even if resources are deployed.",
		flag.Aliases("f"),
	)

	return f
}
