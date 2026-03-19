package command

import (
	"flag"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// groupGetCommand is the top-level structure for the group get command.
type groupGetCommand struct {
	*BaseCommand

	toJSON bool
}

var _ Command = (*groupGetCommand)(nil)

func (c *groupGetCommand) validate() error {
	const message = "group id is required"
	return validation.ValidateStruct(c,
		validation.Field(&c.arguments,
			validation.Required.Error(message),
			validation.Length(1, 1).Error(message),
		),
	)
}

// NewGroupGetCommandFactory returns a groupGetCommand struct.
func NewGroupGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &groupGetCommand{
			BaseCommand: baseCommand,
		}, nil
	}
}

func (c *groupGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("group get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	input := &pb.GetGroupByIDRequest{
		Id: toTRN(trn.ResourceTypeGroup, c.arguments[0]),
	}

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	return outputGroup(c.UI, c.toJSON, group)
}

func (*groupGetCommand) Synopsis() string {
	return "Get a single group."
}

func (*groupGetCommand) Usage() string {
	return "tharsis [global options] group get [options] <id>"
}

func (*groupGetCommand) Description() string {
	return `
   The group get command retrieves a single group by its ID.
   Shows output as JSON, if specified.
`
}

func (*groupGetCommand) Example() string {
	return `
tharsis group get \
  --json \
  trn:tharsis:group:<group_path>
`
}

func (c *groupGetCommand) Flags() *flag.FlagSet {
	f := flag.NewFlagSet("Command options", flag.ContinueOnError)
	f.BoolVar(
		&c.toJSON,
		"json",
		false,
		"Show output as JSON.",
	)

	return f
}

// outputGroup is the output for most group operations.
func outputGroup(ui terminal.UI, toJSON bool, group *pb.Group) int {
	if toJSON {
		if err := ui.JSON(group); err != nil {
			ui.ErrorWithSummary(err, "failed to get JSON output")
			return 1
		}
	} else {
		ui.NamedValues([]terminal.NamedValue{
			{Name: "ID", Value: group.Metadata.Id},
			{Name: "TRN", Value: group.Metadata.Trn},
			{Name: "Name", Value: group.Name},
			{Name: "Full Path", Value: group.FullPath},
			{Name: "Description", Value: group.Description},
			{Name: "Created By", Value: group.CreatedBy},
			{Name: "Created At", Value: group.Metadata.CreatedAt.AsTime().Local().Format(humanTimeFormat)},
			{Name: "Updated At", Value: group.Metadata.UpdatedAt.AsTime().Local().Format(humanTimeFormat)},
		})
	}

	return 0
}
