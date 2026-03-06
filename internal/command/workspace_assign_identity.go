package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceAssignManagedIdentityCommand struct {
	*BaseCommand

	workspaceID       string
	managedIdentityID string
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
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
		validation.Field(&c.workspaceID, validation.Required),
		validation.Field(&c.managedIdentityID, validation.Required),
	)
}

func (c *workspaceAssignManagedIdentityCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace assign-managed-identity"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.AssignManagedIdentityToWorkspaceRequest{
		WorkspaceId:       c.workspaceID,
		ManagedIdentityId: c.managedIdentityID,
	}

	c.Logger.Debug("workspace assign-managed-identity input", "input", input)

	if _, err := c.client.ManagedIdentitiesClient.AssignManagedIdentityToWorkspace(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to assign managed identity to workspace")
		return 1
	}

	c.UI.Successf("Managed identity %s assigned to workspace successfully", c.managedIdentityID)
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
	return "tharsis [global options] workspace assign-managed-identity [options]"
}

func (*workspaceAssignManagedIdentityCommand) Example() string {
	return `
tharsis workspace assign-managed-identity \
  --workspace-id trn:workspace:ops/my-workspace \
  --managed-identity-id trn:managed_identity:ops/my-identity
`
}

func (c *workspaceAssignManagedIdentityCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.workspaceID,
		"workspace-id",
		"",
		"The ID of the workspace.",
	)
	f.StringVar(
		&c.managedIdentityID,
		"managed-identity-id",
		"",
		"The ID of the managed identity to assign.",
	)

	return f
}
