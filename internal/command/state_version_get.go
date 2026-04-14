package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type stateVersionGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*stateVersionGetCommand)(nil)

func (c *stateVersionGetCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewStateVersionGetCommandFactory returns a stateVersionGetCommand struct.
func NewStateVersionGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &stateVersionGetCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *stateVersionGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("state-version get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	sv, err := c.grpcClient.StateVersionsClient.GetStateVersionByID(c.Context, &pb.GetStateVersionByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get state version")
		return 1
	}

	return c.Output(sv, c.toJSON)
}

func (*stateVersionGetCommand) Synopsis() string {
	return "Get a state version."
}

func (*stateVersionGetCommand) Usage() string {
	return "tharsis [global options] state-version get [options] <id>"
}

func (*stateVersionGetCommand) Description() string {
	return `
   Returns details about a Terraform state
   version including its status and
   associated workspace.
`
}

func (*stateVersionGetCommand) Example() string {
	return `
tharsis state-version get <state_version_id>
`
}

func (c *stateVersionGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
