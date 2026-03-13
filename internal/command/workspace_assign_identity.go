package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceAssignManagedIdentityCommand struct {
	*BaseCommand
}

// NewWorkspaceAssignManagedIdentityCommandFactory returns a workspaceAssignManagedIdentityCommand struct.
func NewWorkspaceAssignManagedIdentityCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceAssignManagedIdentityCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceAssignManagedIdentityCommand) validate() error {
	const message = "workspace id and managed identity id are required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(2, 2).Error(message),
		),
	)
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
		WorkspaceId:       toTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
		ManagedIdentityId: toTRN(trn.ResourceTypeManagedIdentity, c.arguments[1]),
	}

	c.Logger.Debug("workspace assign-managed-identity input", "input", input)

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
   The workspace assign-managed-identity command assigns a managed identity to a workspace.
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

func (c *workspaceAssignManagedIdentityCommand) Flags() *flag.FlagSet {
	return nil
}
