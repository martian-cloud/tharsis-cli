package command

import (
	"errors"
	"flag"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceGetMembershipCommand struct {
	*BaseCommand

	serviceAccountID *string
	teamID           *string
	userID           *string
	toJSON           bool
}

// NewWorkspaceGetMembershipCommandFactory returns a workspaceGetMembershipCommand struct.
func NewWorkspaceGetMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceGetMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceGetMembershipCommand) validate() error {
	count := 0
	if c.serviceAccountID != nil {
		count++
	}

	if c.teamID != nil {
		count++
	}

	if c.userID != nil {
		count++
	}

	if count != 1 {
		return errors.New("exactly one of service account, team or user ID must be specified")
	}

	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *workspaceGetMembershipCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace get-membership"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	workspace, err := c.grpcClient.WorkspacesClient.GetWorkspaceByID(c.Context, &pb.GetWorkspaceByIDRequest{
		Id: toTRN(trn.ResourceTypeWorkspace, c.arguments[0]),
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace")
		return 1
	}

	memberships, err := c.grpcClient.NamespaceMembershipsClient.GetNamespaceMembershipsForNamespace(c.Context, &pb.GetNamespaceMembershipsForNamespaceRequest{
		NamespacePath: workspace.FullPath,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace memberships")
		return 1
	}

	var id string
	if c.userID != nil {
		user, err := c.grpcClient.UsersClient.GetUserByID(c.Context, &pb.GetUserByIDRequest{
			Id: *c.userID,
		})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get user")
			return 1
		}

		id = user.Metadata.Id
	}

	if c.teamID != nil {
		team, err := c.grpcClient.TeamsClient.GetTeamByID(c.Context, &pb.GetTeamByIDRequest{
			Id: *c.teamID,
		})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get team")
			return 1
		}

		id = team.Metadata.Id
	}

	if c.serviceAccountID != nil {
		serviceAccount, err := c.grpcClient.ServiceAccountsClient.GetServiceAccountByID(c.Context, &pb.GetServiceAccountByIDRequest{
			Id: *c.serviceAccountID,
		})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get service account")
			return 1
		}

		id = serviceAccount.Metadata.Id
	}

	var foundMembership *pb.NamespaceMembership
	for _, membership := range memberships.NamespaceMemberships {
		if ptr.ToString(membership.TeamId) == id ||
			ptr.ToString(membership.ServiceAccountId) == id ||
			ptr.ToString(membership.UserId) == id {
			foundMembership = membership
			break
		}
	}

	if foundMembership == nil {
		c.UI.Errorf("no membership found for the specified principal")
		return 1
	}

	return outputMembership(c.UI, c.toJSON, foundMembership)
}

func (*workspaceGetMembershipCommand) Synopsis() string {
	return "Get a workspace membership."
}

func (*workspaceGetMembershipCommand) Description() string {
	return `
   The workspace get-membership command retrieves details about a specific workspace membership.
`
}

func (*workspaceGetMembershipCommand) Usage() string {
	return "tharsis [global options] workspace get-membership [options] <workspace-id>"
}

func (*workspaceGetMembershipCommand) Example() string {
	return `
tharsis workspace get-membership \
  --user-id trn:user:<username> \
  trn:workspace:<workspace_path>
`
}

func (c *workspaceGetMembershipCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)
	f.Func(
		"service-account-id",
		"Service account ID to find the workspace membership for.",
		func(s string) error {
			c.serviceAccountID = &s
			return nil
		},
	)
	f.Func(
		"user-id",
		"User ID to find the workspace membership for.",
		func(s string) error {
			c.userID = &s
			return nil
		},
	)
	f.Func(
		"team-id",
		"Team ID to find the workspace membership for. Deprecated",
		func(s string) error {
			c.teamID = &s
			return nil
		},
	)
	f.Func(
		"username",
		"Username to find the workspace membership for. Deprecated",
		func(s string) error {
			c.userID = ptr.String(trn.NewResourceTRN(trn.ResourceTypeUser, s))
			return nil
		},
	)
	f.Func(
		"team-name",
		"Team name to find the workspace membership for. Deprecated",
		func(s string) error {
			c.teamID = ptr.String(trn.NewResourceTRN(trn.ResourceTypeTeam, s))
			return nil
		},
	)

	return f
}
