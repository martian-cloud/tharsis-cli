package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupUpdateMembershipCommand struct {
	*BaseCommand

	roleID  *string
	version *int64
	toJSON  *bool
}

var _ Command = (*groupUpdateMembershipCommand)(nil)

// NewGroupUpdateMembershipCommandFactory returns a groupUpdateMembershipCommand struct.
func NewGroupUpdateMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupUpdateMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupUpdateMembershipCommand) validate() error {
	const message = "membership-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *groupUpdateMembershipCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group update-membership"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.UpdateNamespaceMembershipRequest{
		Id:      c.arguments[0],
		RoleId:  *c.roleID,
		Version: c.version,
	}

	membership, err := c.grpcClient.NamespaceMembershipsClient.UpdateNamespaceMembership(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update group membership")
		return 1
	}

	return c.Output(membership, c.toJSON)
}

func (*groupUpdateMembershipCommand) Synopsis() string {
	return "Update a group membership."
}

func (*groupUpdateMembershipCommand) Description() string {
	return `
   The group update-membership command updates a group membership's role.
`
}

func (*groupUpdateMembershipCommand) Usage() string {
	return "tharsis [global options] group update-membership [options] <membership-id>"
}

func (*groupUpdateMembershipCommand) Example() string {
	return `
tharsis group update-membership \
  -role-id "trn:role:<role_name>" \
  <id>
`
}

func (c *groupUpdateMembershipCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.roleID,
		"role-id",
		"The role ID for the membership.",
		flag.Required(),
	)
	f.StringVar(
		&c.roleID,
		"role",
		"New role for the membership.",
		flag.Deprecated("use -role-id"),
		flag.TransformString(func(s string) string {
			return trn.NewResourceTRN(trn.ResourceTypeRole, s)
		}),
	)
	f.Int64Var(
		&c.version,
		"version",
		"Metadata version of the resource to be updated. In most cases, this is not required.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
