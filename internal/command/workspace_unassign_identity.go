package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceUnassignManagedIdentityCommand struct {
	*BaseCommand
}

var _ Command = (*workspaceUnassignManagedIdentityCommand)(nil)

// NewWorkspaceUnassignManagedIdentityCommandFactory returns a workspaceUnassignManagedIdentityCommand struct.
func NewWorkspaceUnassignManagedIdentityCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceUnassignManagedIdentityCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceUnassignManagedIdentityCommand) validate() error {
	const message = "workspace id and managed identity id are required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(2, 2).Error(message),
		),
	)
}

func (c *workspaceUnassignManagedIdentityCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithCommandName("workspace unassign-managed-identity"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.RemoveManagedIdentityFromWorkspaceRequest{
		WorkspaceId:       trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
		ManagedIdentityId: trn.ToTRN(trn.ResourceTypeManagedIdentity, c.arguments[1]),
	}

	if _, err := c.grpcClient.ManagedIdentitiesClient.RemoveManagedIdentityFromWorkspace(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to unassign managed identity from workspace")
		return 1
	}

	c.UI.Successf("Managed identity unassigned from workspace successfully!")
	return 0
}

func (*workspaceUnassignManagedIdentityCommand) Synopsis() string {
	return "Unassign a managed identity from a workspace."
}

func (*workspaceUnassignManagedIdentityCommand) Description() string {
	return `
   The workspace unassign-managed-identity command removes a managed identity from a workspace.
`
}

func (*workspaceUnassignManagedIdentityCommand) Usage() string {
	return "tharsis [global options] workspace unassign-managed-identity <workspace-id> <identity-id>"
}

func (*workspaceUnassignManagedIdentityCommand) Example() string {
	return `
tharsis workspace unassign-managed-identity \
  trn:workspace:<workspace_path> \
  trn:managed_identity:<group_path>/<identity_name>
`
}
