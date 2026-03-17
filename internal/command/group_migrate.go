package command

import (
	"errors"
	"flag"
	"strconv"

	"github.com/aws/smithy-go/ptr"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

type groupMigrateCommand struct {
	*BaseCommand

	toTopLevel  *bool
	newParentID *string
	toJSON      bool
}

// NewGroupMigrateCommandFactory returns a groupMigrateCommand struct.
func NewGroupMigrateCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupMigrateCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupMigrateCommand) validate() error {
	if c.newParentID != nil && c.toTopLevel != nil {
		return errors.New("must supply only one of --new-parent-path and --to-top-level")
	}

	const message = "group-id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

func (c *groupMigrateCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group migrate"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.MigrateGroupRequest{
		GroupId:     toTRN(trn.ResourceTypeGroup, c.arguments[0]),
		NewParentId: c.newParentID,
	}

	group, err := c.grpcClient.GroupsClient.MigrateGroup(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to migrate group")
		return 1
	}

	return outputGroup(c.UI, c.toJSON, group)
}

func (*groupMigrateCommand) Synopsis() string {
	return "Migrate a group to a new parent or to top-level."
}

func (*groupMigrateCommand) Description() string {
	return `
   The group migrate command migrates a group to another parent group or to top-level.
   Omit --new-parent-id to migrate to top-level.
`
}

func (*groupMigrateCommand) Usage() string {
	return "tharsis [global options] group migrate [options] <group-id>"
}

func (*groupMigrateCommand) Example() string {
	return `
tharsis group migrate \
  --new-parent-id trn:group:<parent_group_path> \
  trn:group:<group_path>
`
}

func (c *groupMigrateCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.Func(
		"new-parent-id",
		"New parent group ID. Omit to migrate to top-level.",
		func(s string) error {
			c.newParentID = &s
			return nil
		},
	)
	f.Func(
		"new-parent-path",
		"New parent path for the group. Deprecated",
		func(s string) error {
			c.newParentID = ptr.String(trn.NewResourceTRN(trn.ResourceTypeGroup, s))
			return nil
		},
	)
	f.BoolFunc(
		"to-top-level",
		"Migrate group to top level. Deprecated.",
		func(s string) error {
			b, err := strconv.ParseBool(s)
			if err != nil {
				return err
			}

			c.toTopLevel = &b
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
