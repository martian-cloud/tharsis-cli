package command

import (
	"flag"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

// groupUpdateCommand is the top-level structure for the group update command.
type groupUpdateCommand struct {
	*BaseCommand

	description *string
	version     *int64
	toJSON      bool
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

	groupID := c.arguments[0]

	input := &pb.UpdateGroupRequest{
		Id:          groupID,
		Description: c.description,
		Version:     c.version,
	}

	c.Logger.Debug("group update input", "input", input)

	updatedGroup, err := c.client.GroupsClient.UpdateGroup(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to update a group")
		return 1
	}

	return outputGroup(c.UI, c.toJSON, updatedGroup)
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
  --description "Updated operations group" \
  trn:group:ops/my-group
`
}

func (c *groupUpdateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"description",
		"Description for the group.",
		func(s string) error {
			c.description = &s
			return nil
		},
	)
	f.Func(
		"version",
		"Metadata version of the resource to be updated. "+
			"In most cases, this is not required.",
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
		"Show final output as JSON.",
	)

	return f
}
