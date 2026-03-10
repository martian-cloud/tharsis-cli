package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceGetMembershipCommand struct {
	*BaseCommand

	toJSON bool
}

// NewWorkspaceGetMembershipCommandFactory returns a workspaceGetMembershipCommand struct.
func NewWorkspaceGetMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceGetMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceGetMembershipCommand) validate() error {
	const message = "membership-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *workspaceGetMembershipCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace get-membership"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetNamespaceMembershipByIDRequest{
		Id: c.arguments[0],
	}

	c.Logger.Debug("workspace get-membership input", "input", input)

	membership, err := c.grpcClient.NamespaceMembershipsClient.GetNamespaceMembershipByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get workspace membership")
		return 1
	}

	return outputMembership(c.UI, c.toJSON, membership)
}

func (*workspaceGetMembershipCommand) Synopsis() string {
	return "Get a workspace membership by ID."
}

func (*workspaceGetMembershipCommand) Description() string {
	return `
   The workspace get-membership command retrieves details about a specific workspace membership.
`
}

func (*workspaceGetMembershipCommand) Usage() string {
	return "tharsis [global options] workspace get-membership [options] <membership-id>"
}

func (*workspaceGetMembershipCommand) Example() string {
	return `
tharsis workspace get-membership trn:namespace_membership:ops/my-workspace/Tk1fZj
`
}

func (c *workspaceGetMembershipCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Output in JSON format.",
	)

	return f
}
