package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceGetAssignedManagedIdentitiesCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*workspaceGetAssignedManagedIdentitiesCommand)(nil)

// NewWorkspaceGetAssignedManagedIdentitiesCommandFactory returns a workspaceGetAssignedManagedIdentitiesCommand struct.
func NewWorkspaceGetAssignedManagedIdentitiesCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceGetAssignedManagedIdentitiesCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceGetAssignedManagedIdentitiesCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *workspaceGetAssignedManagedIdentitiesCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace get-assigned-managed-identities"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetManagedIdentitiesForWorkspaceRequest{
		WorkspaceId: trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	}

	result, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentitiesForWorkspace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get assigned managed identities")
		return 1
	}

	return c.OutputList(result.ManagedIdentities, c.toJSON)
}

func (*workspaceGetAssignedManagedIdentitiesCommand) Synopsis() string {
	return "Get assigned managed identities for a workspace."
}

func (*workspaceGetAssignedManagedIdentitiesCommand) Description() string {
	return `
   The workspace get-assigned-managed-identities command lists managed identities assigned to a workspace.
`
}

func (*workspaceGetAssignedManagedIdentitiesCommand) Usage() string {
	return "tharsis [global options] workspace get-assigned-managed-identities [options] <workspace-id>"
}

func (*workspaceGetAssignedManagedIdentitiesCommand) Example() string {
	return `
tharsis workspace get-assigned-managed-identities trn:workspace:<workspace_path>
`
}

func (c *workspaceGetAssignedManagedIdentitiesCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
