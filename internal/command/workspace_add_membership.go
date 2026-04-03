package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceAddMembershipCommand struct {
	*BaseCommand

	roleID           *string
	userID           *string
	serviceAccountID *string
	teamID           *string
	toJSON           *bool
}

var _ Command = (*workspaceAddMembershipCommand)(nil)

// NewWorkspaceAddMembershipCommandFactory returns a workspaceAddMembershipCommand struct.
func NewWorkspaceAddMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceAddMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceAddMembershipCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: workspace id")
	}

	if c.roleID == nil {
		return errors.New("role id is required")
	}

	return nil
}

func (c *workspaceAddMembershipCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace add-membership"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: trn.ToTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	input := &pb.CreateNamespaceMembershipRequest{
		NamespacePath:    workspace.FullPath,
		RoleId:           *c.roleID,
		UserId:           c.userID,
		ServiceAccountId: c.serviceAccountID,
		TeamId:           c.teamID,
	}

	membership, err := c.grpcClient.NamespaceMembershipsClient.CreateNamespaceMembership(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to add workspace membership")
		return 1
	}

	return c.Output(membership, c.toJSON)
}

func (*workspaceAddMembershipCommand) Synopsis() string {
	return "Add a membership to a workspace."
}

func (*workspaceAddMembershipCommand) Description() string {
	return `
   Grants a user, service account, or team access to a
   workspace. Exactly one identity flag must be specified.
`
}

func (*workspaceAddMembershipCommand) Usage() string {
	return "tharsis [global options] workspace add-membership [options] <workspace-id>"
}

func (*workspaceAddMembershipCommand) Example() string {
	return `
tharsis workspace add-membership \
  -role-id "trn:role:owner" \
  -user-id "trn:user:john.smith" \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceAddMembershipCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.roleID,
		"role-id",
		"The role ID for the membership.",
	)
	f.StringVar(
		&c.userID,
		"user-id",
		"The user ID for the membership.",
	)
	f.StringVar(
		&c.roleID,
		"role",
		"Role name for new membership.",
		flag.Deprecated("use -role-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeRole, s)
		}),
	)
	f.StringVar(
		&c.userID,
		"username",
		"Username for the new membership.",
		flag.Deprecated("use -user-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeUser, s)
		}),
	)
	f.StringVar(
		&c.serviceAccountID,
		"service-account-id",
		"The service account ID for the membership.",
	)
	f.StringVar(
		&c.teamID,
		"team-id",
		"The team ID for the membership.",
	)
	f.StringVar(
		&c.teamID,
		"team-name",
		"Team name for the new membership.",
		flag.Deprecated("use -team-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeTeam, s)
		}),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	f.MutuallyExclusive("user-id", "service-account-id", "team-id", "username", "team-name")

	return f
}
