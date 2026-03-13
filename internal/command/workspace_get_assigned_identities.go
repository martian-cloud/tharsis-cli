package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceGetAssignedManagedIdentitiesCommand struct {
	*BaseCommand

	toJSON bool
}

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
		WorkspaceId: toTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	}

	c.Logger.Debug("workspace get-assigned-managed-identities input", "input", input)

	result, err := c.grpcClient.ManagedIdentitiesClient.GetManagedIdentitiesForWorkspace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get assigned managed identities")
		return 1
	}

	if c.toJSON {
		if err := c.UI.JSON(result.ManagedIdentities); err != nil {
			c.UI.ErrorWithSummary(err, "failed to output JSON")
			return 1
		}
		return 0
	}

	t := terminal.NewTable("id", "name", "group id", "type")
	for _, identity := range result.ManagedIdentities {
		t.Rich([]string{
			identity.Metadata.Id,
			identity.Name,
			identity.GroupId,
			identity.Type,
		}, nil)
	}

	c.UI.Table(t)
	return 0
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

func (c *workspaceGetAssignedManagedIdentitiesCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}
