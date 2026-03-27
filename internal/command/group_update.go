package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// groupUpdateCommand is the top-level structure for the group update command.
type groupUpdateCommand struct {
	*BaseCommand

	description *string
	version     *int64
	toJSON      *bool
}

var _ Command = (*groupUpdateCommand)(nil)

func (c *groupUpdateCommand) validate() error {
	const message = "id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewGroupUpdateCommandFactory returns a groupUpdateCommand struct.
func NewGroupUpdateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupUpdateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupUpdateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group update"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.UpdateGroupRequest{
		Id:          trn.ToTRN(trn.ResourceTypeGroup, c.arguments[0]),
		Description: c.description,
		Version:     c.version,
	}

	updatedGroup, err := c.grpcClient.GroupsClient.UpdateGroup(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update a group")
		return 1
	}

	return c.Output(updatedGroup, c.toJSON)
}

func (*groupUpdateCommand) Synopsis() string {
	return "Update a group."
}

func (*groupUpdateCommand) Usage() string {
	return "tharsis [global options] group update [options] <id>"
}

func (*groupUpdateCommand) Description() string {
	return `
   The group update command updates a group. Currently, it
   supports updating the description. Shows final output
   as JSON, if specified.
`
}

func (*groupUpdateCommand) Example() string {
	return `
tharsis group update \
  -description "Updated operations group" \
  trn:group:<group_path>
`
}

func (c *groupUpdateCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.StringVar(
		&c.description,
		"description",
		"Description for the group.",
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
		flag.Default(false),
	)

	return f
}
