package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceListMembershipsCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*workspaceListMembershipsCommand)(nil)

// NewWorkspaceListMembershipsCommandFactory returns a workspaceListMembershipsCommand struct.
func NewWorkspaceListMembershipsCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceListMembershipsCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceListMembershipsCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *workspaceListMembershipsCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace list-memberships"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	// Ensure it's a workspace.
	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	input := &pb.GetNamespaceMembershipsForNamespaceRequest{
		NamespacePath: workspace.FullPath,
	}

	result, err := c.grpcClient.NamespaceMembershipsClient.GetNamespaceMembershipsForNamespace(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get a list of workspace memberships")
		return 1
	}

	return c.OutputList(result, c.toJSON)
}

func (*workspaceListMembershipsCommand) Synopsis() string {
	return "Retrieve a list of workspace memberships."
}

func (*workspaceListMembershipsCommand) Description() string {
	return `
   The workspace list-memberships command prints information about
   memberships for a specific workspace.
`
}

func (*workspaceListMembershipsCommand) Usage() string {
	return "tharsis [global options] workspace list-memberships [options] <id>"
}

func (*workspaceListMembershipsCommand) Example() string {
	return `
tharsis workspace list-memberships trn:workspace:<workspace_path>
`
}

func (c *workspaceListMembershipsCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
		flag.Default(false),
	)

	return f
}
