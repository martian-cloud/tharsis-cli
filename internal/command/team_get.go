package command

import (
	"errors"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/flag"
)

type teamGetCommand struct {
	*BaseCommand

	toJSON *bool
}

var _ Command = (*teamGetCommand)(nil)

func (c *teamGetCommand) validate() error {
	if len(c.arguments) != 1 {
		return errors.New("expected exactly one argument: id")
	}

	return nil
}

// NewTeamGetCommandFactory returns a teamGetCommand struct.
func NewTeamGetCommandFactory(baseCommand *BaseCommand) func() (Command, error) {
	return func() (Command, error) {
		return &teamGetCommand{BaseCommand: baseCommand}, nil
	}
}

func (c *teamGetCommand) Run(args []string) int {
	if code := c.initialize(
		WithArguments(args),
		WithFlags(c.Flags()),
		WithCommandName("team get"),
		WithInputValidator(c.validate),
		WithClient(true),
	); code != 0 {
		return code
	}

	team, err := c.grpcClient.TeamsClient.GetTeamByID(c.Context, &pb.GetTeamByIDRequest{Id: c.arguments[0]})
	if err != nil {
		c.UI.ErrorWithSummary(err, "failed to get team")
		return 1
	}

	return c.Output(team, c.toJSON)
}

func (*teamGetCommand) Synopsis() string {
	return "Get a team."
}

func (*teamGetCommand) Usage() string {
	return "tharsis [global options] team get [options] <id>"
}

func (*teamGetCommand) Description() string {
	return `
   Retrieves details about a team including
   its name and description.
`
}

func (*teamGetCommand) Example() string {
	return `
tharsis team get <team_id>
`
}

func (c *teamGetCommand) Flags() *flag.Set {
	f := flag.NewSet("Command options")
	f.BoolVar(
		&c.toJSON,
		"json",
		"Show final output as JSON.",
	)

	return f
}
