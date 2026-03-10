package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceUpdateMembershipCommand struct {
	*BaseCommand

	roleID  string
	version *int64
	toJSON  bool
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
		validation.Field(&c.roleID, validation.Required),
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
		RoleId:  c.roleID,
		Version: c.version,
	}

	c.Logger.Debug("workspace update-membership input", "input", input)

	membership, err := c.grpcClient.NamespaceMembershipsClient.UpdateNamespaceMembership(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update workspace membership")
		return 1
	}

	return outputMembership(c.UI, c.toJSON, membership)
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
  --role-id trn:role:deployer \
  trn:namespace_membership:ops/my-workspace/Tk1fZ...
`
}

func (c *workspaceUpdateMembershipCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.StringVar(
		&c.roleID,
		"role-id",
		"",
		"The role ID for the membership.",
	)
	f.Func(
		"version",
		"Metadata version of the resource to be updated. In most cases, this is not required.",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			c.version = &v
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
