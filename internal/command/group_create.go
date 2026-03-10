package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// groupCreateCommand is the top-level structure for the group create command.
type groupCreateCommand struct {
	*BaseCommand

	parentGroupID *string
	description   string
	toJSON        bool
	ifNotExists   bool
}

var _ Command = (*groupCreateCommand)(nil)

func (c *groupCreateCommand) validate() error {
	const message = "name is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewGroupCreateCommandFactory returns a groupCreateCommand struct.
func NewGroupCreateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupCreateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupCreateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group create"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	name := c.arguments[0]

	if c.ifNotExists {
		var checkID string
		if c.parentGroupID != nil {
			c.Logger.Debug("getting parent group", "value", *c.parentGroupID)

			group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: *c.parentGroupID})
			if err != nil {
				c.UI.ErrorWithSummary(err, "failed to get parent group")
				return 1
			}

			checkID = trn.NewResourceTRN(trn.ResourceTypeGroup, group.FullPath, name)
		} else {
			checkID = trn.NewResourceTRN(trn.ResourceTypeGroup, name)
		}

		c.Logger.Debug("checking if group exists", "value", checkID)

		existingGroup, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, &pb.GetGroupByIDRequest{Id: checkID})
		if err != nil && status.Code(err) != codes.NotFound {
			c.UI.ErrorWithSummary(err, "failed to check group")
			return 1
		}

		if existingGroup != nil {
			c.Logger.Debug("group already exists, returning existing group")
			return outputGroup(c.UI, c.toJSON, existingGroup)
		}
	}

	input := &pb.CreateGroupRequest{
		Name:        name,
		ParentId:    c.parentGroupID,
		Description: c.description,
	}

	c.Logger.Debug("group create input", "input", input)

	createdGroup, err := c.grpcClient.GroupsClient.CreateGroup(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to create a group")
		return 1
	}

	return outputGroup(c.UI, c.toJSON, createdGroup)
}

func (*groupCreateCommand) Synopsis() string {
	return "Create a new group."
}

func (*groupCreateCommand) Usage() string {
	return "tharsis [global options] group create [options] <name>"
}

func (*groupCreateCommand) Description() string {
	return `
   The group create command creates a new group. It allows
   setting a group's description (optional). Shows final
   output as JSON, if specified. Idempotent when used with
   --if-not-exists option.
`
}

func (*groupCreateCommand) Example() string {
	return `
tharsis group create \
  --parent-group-id trn:group:ops \
  --description "Operations group" \
  my-group
`
}

func (c *groupCreateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"parent-group-id",
		"Parent group ID.",
		func(s string) error {
			c.parentGroupID = &s
			return nil
		},
	)
	f.StringVar(
		&c.description,
		"description",
		"",
		"Description for the new group.",
	)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show final output as JSON.",
	)
	f.BoolVar(
		&c.ifNotExists,
		"if-not-exists",
		false,
		"Create a group if it does not already exist.",
	)

	return f
}
