package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/infor-cloud/martian-cloud/phobos/phobos-cli/pkg/terminal"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
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

	input := &pb.DeleteNamespaceMembershipRequest{
		Id:      c.arguments[0],
		Version: c.version,
	}

	c.Logger.Debug("group remove-membership input", "input", input)

	if _, err := c.client.NamespaceMembershipsClient.DeleteNamespaceMembership(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to remove group membership")
		return 1
	}

	c.UI.Output("Group membership removed successfully!", terminal.WithSuccessStyle())
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
tharsis group remove-membership trn:namespace_membership:ops/Tk1fZ...
`
}

func (c *groupRemoveMembershipCommand) Flags() *flag.FlagSet {
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
