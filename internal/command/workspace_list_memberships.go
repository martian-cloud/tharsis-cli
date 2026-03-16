package command

import (
	"flag"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceListMembershipsCommand struct {
	*BaseCommand

	toJSON bool
}

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
		Id: toTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
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

	if c.toJSON {
		if err := c.UI.JSON(result); err != nil {
			c.UI.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "role_id", "user_id", "service_account_id", "team_id")

		for _, membership := range result.NamespaceMemberships {
			t.Rich([]string{
				membership.GetMetadata().Id,
				membership.RoleId,
				ptr.ToString(membership.UserId),
				ptr.ToString(membership.ServiceAccountId),
				ptr.ToString(membership.TeamId),
			}, nil)
		}

		c.UI.Table(t)
	}

	return 0
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

func (c *workspaceListMembershipsCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)

	return f
}
