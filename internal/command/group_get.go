package command

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/trn"
)

// groupGetCommand is the top-level structure for the group get command.
type groupGetCommand struct {
	*BaseCommand

	toJSON *bool
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
		Id: trn.ToTRN(trn.ResourceTypeGroup, c.arguments[0]),
	}

	group, err := c.grpcClient.GroupsClient.GetGroupByID(c.Context, input)
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get group")
		return 1
	}

	return c.OutputProto(group, c.toJSON)
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
  -json \
  trn:tharsis:group:<group_path>
`
}

func (c *groupGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show output as JSON.",
		flag.Default(false),
	)

	return f
}
