package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceAssignManagedIdentityCommand struct {
	*BaseCommand
}

var _ Command = (*workspaceAssignManagedIdentityCommand)(nil)

// NewWorkspaceAssignManagedIdentityCommandFactory returns a workspaceAssignManagedIdentityCommand struct.
func NewWorkspaceAssignManagedIdentityCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceAssignManagedIdentityCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceAssignManagedIdentityCommand) validate() error {
	if len(c.arguments) != 2 {
		return errors.New("expected exactly two arguments: workspace id and managed identity id")
	}

	return nil
}

func (c *workspaceAssignManagedIdentityCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("workspace assign-managed-identity"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.AssignManagedIdentityToWorkspaceRequest{
		WorkspaceId:       trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
		ManagedIdentityId: trn.ToTRN(trn.ResourceTypeManagedIdentity, c.arguments[1]),
	}

	if _, err := c.grpcClient.ManagedIdentitiesClient.AssignManagedIdentityToWorkspace(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to assign managed identity to workspace")
		return 1
	}

	c.UI.Successf("Managed identity assigned to workspace successfully!")
	return 0
}

func (*workspaceAssignManagedIdentityCommand) Synopsis() string {
	return "Assign a managed identity to a workspace."
}

func (*workspaceAssignManagedIdentityCommand) Description() string {
	return `
   Assigns a managed identity to a workspace for cloud
   provider authentication.
`
}

func (*workspaceAssignManagedIdentityCommand) Usage() string {
	return "tharsis [global options] workspace assign-managed-identity <workspace-id> <identity-id>"
}

func (*workspaceAssignManagedIdentityCommand) Example() string {
	return `
tharsis workspace assign-managed-identity \
  trn:workspace:<workspace_path> \
  trn:managed_identity:<group_path>/<identity_name>
`
}
