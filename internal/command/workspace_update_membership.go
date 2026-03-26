package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type workspaceUpdateMembershipCommand struct {
	*BaseCommand

	roleID  *string
	version *int64
	toJSON  *bool
}

// NewWorkspaceUpdateMembershipCommandFactory returns a workspaceUpdateMembershipCommand struct.
func NewWorkspaceUpdateMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceUpdateMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceUpdateMembershipCommand) validate() error {
	const message = "membership-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
		validation.Field(&c.roleID, validation.Required, validation.NotNil),
	)
}

func (c *workspaceUpdateMembershipCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace update-membership"),
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
		c.UI.ErrorWithSummary(err, "failed to update workspace membership")
		return 1
	}

	return c.OutputProto(membership, c.toJSON)
}

func (*workspaceUpdateMembershipCommand) Synopsis() string {
	return "Update a workspace membership."
}

func (*workspaceUpdateMembershipCommand) Description() string {
	return `
   The workspace update-membership command updates a workspace membership's role.
`
}

func (*workspaceUpdateMembershipCommand) Usage() string {
	return "tharsis [global options] workspace update-membership [options] <membership-id>"
}

func (*workspaceUpdateMembershipCommand) Example() string {
	return `
tharsis workspace update-membership \
  -role-id trn:role:<role_name> \
  <id>
`
}

func (c *workspaceUpdateMembershipCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.roleID,
		"role-id",
		"The role ID for the membership.",
	)
	f.StringVar(
		&c.roleID,
		"role",
		"Role name for the membership.",
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
		"Output in JSON format.",
		flag.Default(false),
	)

	return f
}
