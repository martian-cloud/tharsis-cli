package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

type workspaceRemoveMembershipCommand struct {
	*BaseCommand

	version *int64
}

// NewWorkspaceRemoveMembershipCommandFactory returns a workspaceRemoveMembershipCommand struct.
func NewWorkspaceRemoveMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &workspaceRemoveMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *workspaceRemoveMembershipCommand) validate() error {
	const message = "membership-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *workspaceRemoveMembershipCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("workspace remove-membership"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	membership, err := c.grpcClient.NamespaceMembershipsClient.GetNamespaceMembershipByID(c.Context, &pb.GetNamespaceMembershipByIDRequest{
		Id: c.arguments[0],
	})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get namespace membership")
		return 1
	}

	if membership.Namespace == nil || membership.Namespace.WorkspaceId == nil {
		c.UI.Errorf("namespace membership not found for workspace")
		return 1
	}

	input := &pb.DeleteNamespaceMembershipRequest{
		Id:      membership.Metadata.Id,
		Version: c.version,
	}

	if _, err := c.grpcClient.NamespaceMembershipsClient.DeleteNamespaceMembership(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to remove workspace membership")
		return 1
	}

	c.UI.Successf("Workspace membership removed successfully!")
	return 0
}

func (*workspaceRemoveMembershipCommand) Synopsis() string {
	return "Remove a workspace membership."
}

func (*workspaceRemoveMembershipCommand) Description() string {
	return `
   The workspace remove-membership command removes a membership from a workspace.
`
}

func (*workspaceRemoveMembershipCommand) Usage() string {
	return "tharsis [global options] workspace remove-membership [options] <membership-id>"
}

func (*workspaceRemoveMembershipCommand) Example() string {
	return `
tharsis workspace remove-membership <id>
`
}

func (c *workspaceRemoveMembershipCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"version",
		"Metadata version of the resource to be deleted. In most cases, this is not required.",
		func(s string) error {
			v, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return err
			}
			c.version = &v
			return nil
		},
	)

	return f
}
