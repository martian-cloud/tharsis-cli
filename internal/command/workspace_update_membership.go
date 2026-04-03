package command

import (
	"errors"

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

var _ Command = (*workspaceUpdateMembershipCommand)(nil)

// NewWorkspaceUpdateMembershipCommandFactory returns a workspaceUpdateMembershipCommand struct.
func NewWorkspaceUpdateMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceUpdateMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceUpdateMembershipCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: membership id")
	}

	if c.roleID == nil {
		return errors.New("role id is required")
	}

	return nil
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

	return c.Output(membership, c.toJSON)
}

func (*workspaceUpdateMembershipCommand) Synopsis() string {
	return "Update a workspace membership."
}

func (*workspaceUpdateMembershipCommand) Description() string {
	return `
   Changes the role of an existing workspace membership.
`
}

func (*workspaceUpdateMembershipCommand) Usage() string {
	return "tharsis [global options] workspace update-membership [options] <membership-id>"
}

func (*workspaceUpdateMembershipCommand) Example() string {
	return `
tharsis workspace update-membership \
  -role-id "trn:role:<role_name>" \
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
		"Optimistic locking version. Usually not required.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
