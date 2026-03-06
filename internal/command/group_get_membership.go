package command

import (
	"flag"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type groupGetMembershipCommand struct {
	*BaseCommand

	toJSON bool
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

	input := &pb.GetNamespaceMembershipByIDRequest{
		Id: c.arguments[0],
	}

	c.Logger.Debug("group get-membership input", "input", input)

	membership, err := c.client.NamespaceMembershipsClient.GetNamespaceMembershipByID(c.Context, input)
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
tharsis group get-membership trn:namespace_membership:ops/Tk1fZj
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
