package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// groupDeleteCommand is the top-level structure for the group delete command.
type groupDeleteCommand struct {
	*BaseCommand

	version *int64
	force   *bool
}

var _ Command = (*groupDeleteCommand)(nil)

func (c *groupDeleteCommand) validate() error {
	const message = "group id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewGroupDeleteCommandFactory returns a groupDeleteCommand struct.
func NewGroupDeleteCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupDeleteCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupDeleteCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group delete"),
		WithInputValidator(c.validate),
		WithClient(true),
		WithForcePrompt("Are you sure you want to delete this group?"),
	); code != 0 {
		return code
	}

	input := &pb.DeleteGroupRequest{
		Id:      trn.ToTRN(trn.ResourceTypeGroup, c.arguments[0]),
		Force:   c.force,
		Version: c.version,
	}

	if _, err := c.grpcClient.GroupsClient.DeleteGroup(c.Context, input); err != nil {
		c.UI.ErrorWithSummary(err, "failed to delete group")
		return 1
	}

	c.UI.Successf("Group deleted successfully!")
	return 0
}

func (*groupDeleteCommand) Synopsis() string {
	return "Delete a group."
}

func (*groupDeleteCommand) Usage() string {
	return "tharsis [global options] group delete [options] <id>"
}

func (*groupDeleteCommand) Description() string {
	return `
   The group delete command deletes a group by its ID. Includes
   a force flag to delete the group even if resources are
   deployed (dangerous!).
`
}

func (*groupDeleteCommand) Example() string {
	return `
tharsis group delete \
  -force \
  trn:group:<group_path>
`
}

func (c *groupDeleteCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.Int64Var(
		&c.version,
		"version",
		"Metadata version of the resource to be deleted. In most cases, this is not required.",
	)
	f.BoolVar(
		&c.force,
		"force",
		"Force delete the group.",
	)

	return f
}
