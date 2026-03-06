package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceUnassignManagedIdentityCommand struct {
	*BaseCommand

	workspaceID       string
	managedIdentityID string
}

// NewWorkspaceUnassignManagedIdentityCommandFactory returns a workspaceUnassignManagedIdentityCommand struct.
func NewWorkspaceUnassignManagedIdentityCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceUnassignManagedIdentityCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceUnassignManagedIdentityCommand) validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments, validation.Empty),
		validation.Field(&c.workspaceID, validation.Required),
		validation.Field(&c.managedIdentityID, validation.Required),
	)
}

func (c *workspaceUnassignManagedIdentityCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace unassign-managed-identity"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.RemoveManagedIdentityFromWorkspaceRequest{
		WorkspaceId:       c.workspaceID,
		ManagedIdentityId: c.managedIdentityID,
	}

	c.Logger.Debug("workspace unassign-managed-identity input", "input", input)

	if _, err := c.client.ManagedIdentitiesClient.RemoveManagedIdentityFromWorkspace(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to unassign managed identity from workspace")
		return 1
	}

	c.UI.Successf("Managed identity %s unassigned from workspace successfully", c.managedIdentityID)
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
	return "tharsis [global options] workspace unassign-managed-identity [options]"
}

func (*workspaceUnassignManagedIdentityCommand) Example() string {
	return `
tharsis workspace unassign-managed-identity \
  --workspace-id trn:workspace:ops/my-workspace \
  --managed-identity-id trn:managed_identity:ops/my-identity
`
}

func (c *workspaceUnassignManagedIdentityCommand) Flags() *flag.FlagSet {
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
		"The ID of the managed identity to unassign.",
	)

	return f
}
