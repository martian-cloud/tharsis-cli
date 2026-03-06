package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceAddMembershipCommand struct {
	*BaseCommand

	roleID           string
	userID           *string
	serviceAccountID *string
	teamID           *string
	toJSON           bool
}

// NewWorkspaceAddMembershipCommandFactory returns a workspaceAddMembershipCommand struct.
func NewWorkspaceAddMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceAddMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceAddMembershipCommand) validate() error {
	const message = "workspace-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.roleID, validation.Required),
	)
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

	input := &pb.CreateNamespaceMembershipRequest{
		NamespacePath:    c.arguments[0],
		RoleId:           c.roleID,
		UserId:           c.userID,
		ServiceAccountId: c.serviceAccountID,
		TeamId:           c.teamID,
	}

	c.Logger.Debug("workspace add-membership input", "input", input)

	membership, err := c.client.NamespaceMembershipsClient.CreateNamespaceMembership(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to add workspace membership")
		return 1
	}

	return outputMembership(c.UI, c.toJSON, membership)
}

func (*workspaceAddMembershipCommand) Synopsis() string {
	return "Add a membership to a workspace."
}

func (*workspaceAddMembershipCommand) Description() string {
	return `
   The workspace add-membership command adds a membership to a workspace.
   Exactly one of --user-id, --service-account-id, or --team-id must be specified.
`
}

func (*workspaceAddMembershipCommand) Usage() string {
	return "tharsis [global options] workspace add-membership [options] <workspace-id>"
}

func (*workspaceAddMembershipCommand) Example() string {
	return `
tharsis workspace add-membership \
  --role-id trn:role:owner \
  --user-id trn:user:john.smith \
  trn:workspace:top-level/my-workspace
`
}

func (c *workspaceAddMembershipCommand) Flags() *flag.FlagSet {
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
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}
