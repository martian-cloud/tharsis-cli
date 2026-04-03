package command

import (
	"errors"

	"github.com/aws/smithy-go/ptr"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupGetMembershipCommand struct {
	*BaseCommand

	serviceAccountID *string
	teamID           *string
	userID           *string
	toJSON           *bool
}

var _ Command = (*groupGetMembershipCommand)(nil)

// NewGroupGetMembershipCommandFactory returns a groupGetMembershipCommand struct.
func NewGroupGetMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupGetMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupGetMembershipCommand) validate() error {
	if c.serviceAccountID == nil && c.teamID == nil && c.userID == nil {
		return errors.New("exactly one of service account, team or user ID must be specified")
	}

	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: group id")
	}

	return nil
}

func (c *groupGetMembershipCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group get-membership"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: trn.ToTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	memberships, err := c.grpcClient.NamespaceMembershipsClient.GetNamespaceMembershipsForNamespace(c.Context, &pb.GetNamespaceMembershipsForNamespaceRequest{
		NamespacePath: group.FullPath,
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group memberships")
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

	return c.Output(foundMembership, c.toJSON)
}

func (*groupGetMembershipCommand) Synopsis() string {
	return "Get a group membership."
}

func (*groupGetMembershipCommand) Description() string {
	return `
   Retrieves details about a specific group membership.
`
}

func (*groupGetMembershipCommand) Usage() string {
	return "tharsis [global options] group get-membership [options] <group-id>"
}

func (*groupGetMembershipCommand) Example() string {
	return `
tharsis group get-membership \
  -user-id "trn:user:<username>" \
  trn:group:<group_path>
`
}

func (c *groupGetMembershipCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)
	f.StringVar(
		&c.serviceAccountID,
		"service-account-id",
		"Service account ID to find the group membership for.",
	)
	f.StringVar(
		&c.userID,
		"user-id",
		"User ID to find the group membership for.",
	)
	f.StringVar(
		&c.teamID,
		"team-id",
		"Team ID to find the group membership for.",
	)
	f.StringVar(
		&c.userID,
		"username",
		"Username to find the group membership for.",
		flag.Deprecated("use -user-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeUser, s)
		}),
	)
	f.StringVar(
		&c.teamID,
		"team-name",
		"Team name to find the group membership for.",
		flag.Deprecated("use -team-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeTeam, s)
		}),
	)

	f.MutuallyExclusive("user-id", "service-account-id", "team-id", "username", "team-name")

	return f
}
