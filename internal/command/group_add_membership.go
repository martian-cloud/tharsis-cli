package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupAddMembershipCommand struct {
	*BaseCommand

	roleID           *string
	userID           *string
	serviceAccountID *string
	teamID           *string
	toJSON           *bool
}

var _ Command = (*groupAddMembershipCommand)(nil)

// NewGroupAddMembershipCommandFactory returns a groupAddMembershipCommand struct.
func NewGroupAddMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupAddMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupAddMembershipCommand) validate() error {
	const message = "group-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *groupAddMembershipCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group add-membership"),
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

	input := &pb.CreateNamespaceMembershipRequest{
		NamespacePath:    group.FullPath,
		RoleId:           *c.roleID,
		UserId:           c.userID,
		ServiceAccountId: c.serviceAccountID,
		TeamId:           c.teamID,
	}

	membership, err := c.grpcClient.NamespaceMembershipsClient.CreateNamespaceMembership(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to add group membership")
		return 1
	}

	return c.Output(membership, c.toJSON)
}

func (*groupAddMembershipCommand) Synopsis() string {
	return "Add a membership to a group."
}

func (*groupAddMembershipCommand) Description() string {
	return `
   The group add-membership command adds a membership to a group.
   Exactly one of -user-id, -service-account-id, or -team-id must be specified.
`
}

func (*groupAddMembershipCommand) Usage() string {
	return "tharsis [global options] group add-membership [options] <group-id>"
}

func (*groupAddMembershipCommand) Example() string {
	return `
tharsis group add-membership \
  -role-id "trn:role:<role_name>" \
  -user-id "trn:user:<username>" \
  trn:group:<group_path>
`
}

func (c *groupAddMembershipCommand) Flags() *flag.Set {
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
		"The team name for the membership.",
		flag.Deprecated("use -team-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeTeam, s)
		}),
	)
	f.StringVar(
		&c.userID,
		"username",
		"The username for the membership.",
		flag.Deprecated("use -user-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeUser, s)
		}),
	)
	f.StringVar(
		&c.roleID,
		"role",
		"The role for the membership.",
		flag.Deprecated("use -role-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeRole, s)
		}),
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Output in JSON format.",
		flag.Default(false),
	)

	return f
}
