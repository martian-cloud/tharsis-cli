package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type groupRemoveMembershipCommand struct {
	*BaseCommand

	version *int64
}

// NewGroupRemoveMembershipCommandFactory returns a groupRemoveMembershipCommand struct.
func NewGroupRemoveMembershipCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupRemoveMembershipCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupRemoveMembershipCommand) validate() error {
	const message = "membership-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *groupRemoveMembershipCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group remove-membership"),
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

	if membership.Namespace == nil || membership.Namespace.GroupId == nil {
		c.UI.Errorf("namespace membership not found for group")
		return 1
	}

	input := &pb.DeleteNamespaceMembershipRequest{
		Id:      membership.Metadata.Id,
		Version: c.version,
	}

	if _, err := c.grpcClient.NamespaceMembershipsClient.DeleteNamespaceMembership(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to remove group membership")
		return 1
	}

	c.UI.Successf("Group membership removed successfully!")
	return 0
}

func (*groupRemoveMembershipCommand) Synopsis() string {
	return "Remove a group membership."
}

func (*groupRemoveMembershipCommand) Description() string {
	return `
   The group remove-membership command removes a membership from a group.
`
}

func (*groupRemoveMembershipCommand) Usage() string {
	return "tharsis [global options] group remove-membership [options] <membership-id>"
}

func (*groupRemoveMembershipCommand) Example() string {
	return `
tharsis group remove-membership <id>
`
}

func (c *groupRemoveMembershipCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Metadata version of the resource to be deleted. In most cases, this is not required.",
	)

	return f
}
