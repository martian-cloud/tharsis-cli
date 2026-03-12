package command

import (
	"flag"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupGetMembershipCommand struct {
	*BaseCommand

	serviceAccountID *string
	teamName         *string
	username         *string
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
	const message = "membership-id is required"
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

	// Check if this using the deprecated group path argument.
	if g := toTRN(trn.ResourceTypeGroup, c.arguments[0]); g != "" {
		group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: g})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get group")
			return 1
		}

		// This is passing in a group path argument (deprecated)
		// so we'll need to lookup the membership a different way.
		memberships, err := c.grpcClient.NamespaceMembershipsClient.GetNamespaceMembershipsForNamespace(c.Context, &pb.GetNamespaceMembershipsForNamespaceRequest{
			NamespacePath: group.FullPath,
		})
		if err != nil {
			c.UI.ErrorWithSummary(err, "failed to get group memberships")
			return 1
		}

		var foundMembership *pb.NamespaceMembership
		for _, membership := range memberships.NamespaceMemberships {
			if c.username != nil && membership.UserId != nil {
				user, err := c.grpcClient.UsersClient.GetUserByID(c.Context, &pb.GetUserByIDRequest{
					Id: trn.NewResourceTRN(trn.ResourceTypeUser, *c.username),
				})
				if err != nil {
					c.UI.ErrorWithSummary(err, "failed to get user")
					return 1
				}

				if user.Metadata.Id == *membership.UserId {
					foundMembership = membership
					break
				}
			}

			if c.teamName != nil && membership.TeamId != nil {
				team, err := c.grpcClient.TeamsClient.GetTeamByID(c.Context, &pb.GetTeamByIDRequest{
					Id: trn.NewResourceTRN(trn.ResourceTypeTeam, *c.teamName),
				})
				if err != nil {
					c.UI.ErrorWithSummary(err, "failed to get team")
					return 1
				}

				if team.Metadata.Id == *membership.TeamId {
					foundMembership = membership
					break
				}
			}

			if c.serviceAccountID != nil && ptr.ToString(membership.ServiceAccountId) == *c.serviceAccountID {
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

	input := &pb.GetNamespaceMembershipByIDRequest{
		Id: c.arguments[0],
	}

	c.Logger.Debug("group get-membership input", "input", input)

	membership, err := c.grpcClient.NamespaceMembershipsClient.GetNamespaceMembershipByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group membership")
		return 1
	}

	return outputMembership(c.UI, c.toJSON, membership)
}

func (*groupGetMembershipCommand) Synopsis() string {
	return "Get a group membership by ID."
}

func (*groupGetMembershipCommand) Description() string {
	return `
   The group get-membership command retrieves details about a specific group membership.
`
}

func (*groupGetMembershipCommand) Usage() string {
	return "tharsis [global options] group get-membership [options] <membership-id>"
}

func (*groupGetMembershipCommand) Example() string {
	return `
tharsis group get-membership <id>
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
		"Service account ID to find the group membership. Deprecated",
		func(s string) error {
			c.serviceAccountID = &s
			return nil
		},
	)
	f.Func(
		"username",
		"Username to find the group membership. Deprecated",
		func(s string) error {
			c.username = &s
			return nil
		},
	)
	f.Func(
		"team-name",
		"Team name to find the group membership. Deprecated",
		func(s string) error {
			c.teamName = &s
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
