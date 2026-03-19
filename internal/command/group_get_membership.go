package command

import (
	"errors"
	"flag"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupGetMembershipCommand struct {
	*BaseCommand

	serviceAccountID *string
	teamID           *string
	userID           *string
	toJSON           bool
}

// NewGroupGetMembershipCommandFactory returns a groupGetMembershipCommand struct.
func NewGroupGetMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupGetMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupGetMembershipCommand) validate() error {
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

	const message = "group id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
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

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: toTRN(trn.ResourceTypeGroup, c.arguments[0])})
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

	return outputMembership(c.UI, c.toJSON, foundMembership)
}

func (*groupGetMembershipCommand) Synopsis() string {
	return "Get a group membership."
}

func (*groupGetMembershipCommand) Description() string {
	return `
   The group get-membership command retrieves details about a specific group membership.
`
}

func (*groupGetMembershipCommand) Usage() string {
	return "tharsis [global options] group get-membership [options] <group-id>"
}

func (*groupGetMembershipCommand) Example() string {
	return `
tharsis group get-membership \
  --user-id trn:user:<username> \
  trn:group:<group_path>
`
}

func (c *groupGetMembershipCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)
	f.Func(
		"service-account-id",
		"Service account ID to find the group membership for.",
		func(s string) error {
			c.serviceAccountID = &s
			return nil
		},
	)
	f.Func(
		"user-id",
		"User ID to find the group membership for.",
		func(s string) error {
			c.userID = &s
			return nil
		},
	)
	f.Func(
		"team-id",
		"Team ID to find the group membership for. Deprecated",
		func(s string) error {
			c.teamID = &s
			return nil
		},
	)
	f.Func(
		"username",
		"Username to find the group membership for. Deprecated",
		func(s string) error {
			c.userID = ptr.String(trn.NewResourceTRN(trn.ResourceTypeUser, s))
			return nil
		},
	)
	f.Func(
		"team-name",
		"Team name to find the group membership for. Deprecated",
		func(s string) error {
			c.teamID = ptr.String(trn.NewResourceTRN(trn.ResourceTypeTeam, s))
			return nil
		},
	)

	return f
}

func outputMembership(ui terminal.UI, toJSON bool, membership *pb.NamespaceMembership) int {
	if toJSON {
		if err := ui.JSON(membership); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		t := terminal.NewTable("id", "role_id", "user_id", "service_account_id", "team_id")

		t.Rich([]string{
			membership.GetMetadata().Id,
			membership.RoleId,
			ptr.ToString(membership.UserId),
			ptr.ToString(membership.ServiceAccountId),
			ptr.ToString(membership.TeamId),
		}, nil)

		ui.Table(t)
	}

	return 0
}
