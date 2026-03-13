package command

import (
	"flag"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupAddMembershipCommand struct {
	*BaseCommand

	roleID           string
	userID           *string
	serviceAccountID *string
	teamID           *string
	toJSON           bool
}

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
		validation.Field(&c.roleID, validation.Required),
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

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: toTRN(trn.ResourceTypeGroup, c.arguments[0])})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	input := &pb.CreateNamespaceMembershipRequest{
		NamespacePath:    group.FullPath,
		RoleId:           c.roleID,
		UserId:           c.userID,
		ServiceAccountId: c.serviceAccountID,
		TeamId:           c.teamID,
	}

	c.Logger.Debug("group add-membership input", "input", input)

	membership, err := c.grpcClient.NamespaceMembershipsClient.CreateNamespaceMembership(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to add group membership")
		return 1
	}

	return outputMembership(c.UI, c.toJSON, membership)
}

func (*groupAddMembershipCommand) Synopsis() string {
	return "Add a membership to a group."
}

func (*groupAddMembershipCommand) Description() string {
	return `
   The group add-membership command adds a membership to a group.
   Exactly one of --user-id, --service-account-id, or --team-id must be specified.
`
}

func (*groupAddMembershipCommand) Usage() string {
	return "tharsis [global options] group add-membership [options] <group-id>"
}

func (*groupAddMembershipCommand) Example() string {
	return `
tharsis group add-membership \
  --role-id trn:role:<role_name> \
  --user-id trn:user:<username> \
  trn:group:<group_path>
`
}

func (c *groupAddMembershipCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.roleID,
		"role-id",
		"",
		"The role ID for the membership.",
	)
	f.Func(
		"user-id",
		"The user ID for the membership.",
		func(s string) error {
			c.userID = &s
			return nil
		},
	)
	f.Func(
		"service-account-id",
		"The service account ID for the membership.",
		func(s string) error {
			c.serviceAccountID = &s
			return nil
		},
	)
	f.Func(
		"team-id",
		"The team ID for the membership.",
		func(s string) error {
			c.teamID = &s
			return nil
		},
	)
	f.Func(
		"team-name",
		"The team name for the membership. Deprecated.",
		func(s string) error {
			c.teamID = ptr.String(trn.NewResourceTRN(trn.ResourceTypeTeam, s))
			return nil
		},
	)
	f.Func(
		"username",
		"The username for the membership. Deprecated.",
		func(s string) error {
			c.userID = ptr.String(trn.NewResourceTRN(trn.ResourceTypeUser, s))
			return nil
		},
	)
	f.Func(
		"role",
		"The role for the membership. Deprecated.",
		func(s string) error {
			c.roleID = trn.NewResourceTRN(trn.ResourceTypeRole, s)
			return nil
		},
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}
