package command

import (
	"errors"

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
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: group id")
	}

	return nil
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

	return c.Output(group, c.toJSON)
}

func (*groupGetCommand) Synopsis() string {
	return "Get a single group."
}

func (*groupGetCommand) Usage() string {
	return "tharsis [global options] group get [options] <id>"
}

func (*groupGetCommand) Description() string {
	return `
   Retrieves details about a group by ID or path.
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
		"Show final output as JSON.",
	)

	return f
}
